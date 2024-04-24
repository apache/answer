/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package user_external_login

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/internal/base/handler"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/base/translator"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/activity"
	"github.com/apache/incubator-answer/internal/service/export"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	"github.com/apache/incubator-answer/internal/service/user_notification_config"
	"github.com/apache/incubator-answer/pkg/checker"
	"github.com/apache/incubator-answer/pkg/random"
	"github.com/apache/incubator-answer/pkg/token"
	"github.com/apache/incubator-answer/plugin"
	"github.com/google/uuid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type UserExternalLoginRepo interface {
	AddUserExternalLogin(ctx context.Context, user *entity.UserExternalLogin) (err error)
	UpdateInfo(ctx context.Context, userInfo *entity.UserExternalLogin) (err error)
	GetByExternalID(ctx context.Context, provider, externalID string) (userInfo *entity.UserExternalLogin, exist bool, err error)
	GetByUserID(ctx context.Context, provider, userID string) (userInfo *entity.UserExternalLogin, exist bool, err error)
	GetUserExternalLoginList(ctx context.Context, userID string) (resp []*entity.UserExternalLogin, err error)
	DeleteUserExternalLogin(ctx context.Context, userID, externalID string) (err error)
	SetCacheUserExternalLoginInfo(ctx context.Context, key string, info *schema.ExternalLoginUserInfoCache) (err error)
	GetCacheUserExternalLoginInfo(ctx context.Context, key string) (info *schema.ExternalLoginUserInfoCache, err error)
}

// UserExternalLoginService user external login service
type UserExternalLoginService struct {
	userRepo                      usercommon.UserRepo
	userExternalLoginRepo         UserExternalLoginRepo
	userCommonService             *usercommon.UserCommon
	emailService                  *export.EmailService
	siteInfoCommonService         siteinfo_common.SiteInfoCommonService
	userActivity                  activity.UserActiveActivityRepo
	userNotificationConfigService *user_notification_config.UserNotificationConfigService
}

// NewUserExternalLoginService new user external login service
func NewUserExternalLoginService(
	userRepo usercommon.UserRepo,
	userCommonService *usercommon.UserCommon,
	userExternalLoginRepo UserExternalLoginRepo,
	emailService *export.EmailService,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
	userActivity activity.UserActiveActivityRepo,
	userNotificationConfigService *user_notification_config.UserNotificationConfigService,
) *UserExternalLoginService {
	return &UserExternalLoginService{
		userRepo:                      userRepo,
		userCommonService:             userCommonService,
		userExternalLoginRepo:         userExternalLoginRepo,
		emailService:                  emailService,
		siteInfoCommonService:         siteInfoCommonService,
		userActivity:                  userActivity,
		userNotificationConfigService: userNotificationConfigService,
	}
}

// ExternalLogin if user is already a member logged in
func (us *UserExternalLoginService) ExternalLogin(
	ctx context.Context, externalUserInfo *schema.ExternalLoginUserInfoCache) (
	resp *schema.UserExternalLoginResp, err error) {
	if len(externalUserInfo.ExternalID) == 0 {
		return &schema.UserExternalLoginResp{
			ErrTitle: translator.Tr(handler.GetLangByCtx(ctx), reason.UserAccessDenied),
			ErrMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.UserExternalLoginMissingUserID),
		}, nil
	}

	oldExternalLoginUserInfo, exist, err := us.userExternalLoginRepo.GetByExternalID(ctx,
		externalUserInfo.Provider, externalUserInfo.ExternalID)
	if err != nil {
		return nil, err
	}
	if exist {
		// if user is already a member, login directly
		oldUserInfo, exist, err := us.userRepo.GetByUserID(ctx, oldExternalLoginUserInfo.UserID)
		if err != nil {
			return nil, err
		}
		if exist && oldUserInfo.Status != entity.UserStatusDeleted {
			if err := us.userRepo.UpdateLastLoginDate(ctx, oldUserInfo.ID); err != nil {
				log.Errorf("update user last login date failed: %v", err)
			}
			newMailStatus, err := us.activeUser(ctx, oldUserInfo, externalUserInfo)
			if err != nil {
				log.Error(err)
			}
			accessToken, _, err := us.userCommonService.CacheLoginUserInfo(
				ctx, oldUserInfo.ID, newMailStatus, oldUserInfo.Status, oldExternalLoginUserInfo.ExternalID)
			return &schema.UserExternalLoginResp{AccessToken: accessToken}, err
		}
	}

	// cache external user info, waiting for user enter email address.
	if len(externalUserInfo.Email) == 0 {
		bindingKey := token.GenerateToken()
		err = us.userExternalLoginRepo.SetCacheUserExternalLoginInfo(ctx, bindingKey, externalUserInfo)
		if err != nil {
			return nil, err
		}
		return &schema.UserExternalLoginResp{BindingKey: bindingKey}, nil
	}

	// check whether site allow register or not
	siteInfo, err := us.siteInfoCommonService.GetSiteLogin(ctx)
	if err != nil {
		return nil, err
	}
	if !checker.EmailInAllowEmailDomain(externalUserInfo.Email, siteInfo.AllowEmailDomains) {
		log.Debugf("email domain not allowed: %s", externalUserInfo.Email)
		return &schema.UserExternalLoginResp{
			ErrTitle: translator.Tr(handler.GetLangByCtx(ctx), reason.UserAccessDenied),
			ErrMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.EmailIllegalDomainError),
		}, nil
	}

	oldUserInfo, exist, err := us.userRepo.GetByEmail(ctx, externalUserInfo.Email)
	if err != nil {
		return nil, err
	}
	// if user is not a member, register a new user
	if !exist {
		oldUserInfo, err = us.registerNewUser(ctx, externalUserInfo)
		if err != nil {
			return nil, err
		}
	}
	// bind external user info to user
	err = us.bindOldUser(ctx, externalUserInfo, oldUserInfo)
	if err != nil {
		return nil, err
	}

	// If user login with external account and email is exist, active user directly.
	newMailStatus, err := us.activeUser(ctx, oldUserInfo, externalUserInfo)
	if err != nil {
		log.Error(err)
	}

	// set default user notification config for external user
	if err := us.userNotificationConfigService.SetDefaultUserNotificationConfig(ctx, []string{oldUserInfo.ID}); err != nil {
		log.Errorf("set default user notification config failed, err: %v", err)
	}

	accessToken, _, err := us.userCommonService.CacheLoginUserInfo(
		ctx, oldUserInfo.ID, newMailStatus, oldUserInfo.Status, oldExternalLoginUserInfo.ExternalID)
	return &schema.UserExternalLoginResp{AccessToken: accessToken}, err
}

func (us *UserExternalLoginService) registerNewUser(ctx context.Context,
	externalUserInfo *schema.ExternalLoginUserInfoCache) (userInfo *entity.User, err error) {
	userInfo = &entity.User{}
	userInfo.EMail = externalUserInfo.Email
	userInfo.DisplayName = externalUserInfo.DisplayName

	userInfo.Username, err = us.userCommonService.MakeUsername(ctx, externalUserInfo.Username)
	if err != nil {
		log.Error(err)
		userInfo.Username = random.Username()
	}

	if len(externalUserInfo.Avatar) > 0 {
		avatarInfo := &schema.AvatarInfo{
			Type:   constant.AvatarTypeCustom,
			Custom: externalUserInfo.Avatar,
		}
		avatar, _ := json.Marshal(avatarInfo)
		userInfo.Avatar = string(avatar)
	}

	userInfo.MailStatus = entity.EmailStatusToBeVerified
	userInfo.Status = entity.UserStatusAvailable
	userInfo.LastLoginDate = time.Now()
	userInfo.Bio = externalUserInfo.Bio
	userInfo.BioHTML = externalUserInfo.Bio
	err = us.userRepo.AddUser(ctx, userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}

func (us *UserExternalLoginService) bindOldUser(ctx context.Context,
	externalUserInfo *schema.ExternalLoginUserInfoCache, oldUserInfo *entity.User) (err error) {
	oldExternalUserInfo, exist, err := us.userExternalLoginRepo.GetByExternalID(ctx,
		externalUserInfo.Provider,
		externalUserInfo.ExternalID)
	if err != nil {
		return err
	}
	if exist {
		oldExternalUserInfo.MetaInfo = externalUserInfo.MetaInfo
		oldExternalUserInfo.UserID = oldUserInfo.ID
		err = us.userExternalLoginRepo.UpdateInfo(ctx, oldExternalUserInfo)
	} else {
		newExternalUserInfo := &entity.UserExternalLogin{
			UserID:     oldUserInfo.ID,
			Provider:   externalUserInfo.Provider,
			ExternalID: externalUserInfo.ExternalID,
			MetaInfo:   externalUserInfo.MetaInfo,
		}
		err = us.userExternalLoginRepo.AddUserExternalLogin(ctx, newExternalUserInfo)
	}
	return err
}

func (us *UserExternalLoginService) activeUser(ctx context.Context, oldUserInfo *entity.User,
	externalUserInfo *schema.ExternalLoginUserInfoCache) (
	mailStatus int, err error) {
	log.Infof("user %s login with external account, try to active email, old status is %d",
		oldUserInfo.ID, oldUserInfo.MailStatus)

	// try to active user email
	if oldUserInfo.MailStatus == entity.EmailStatusToBeVerified {
		err = us.userRepo.UpdateEmailStatus(ctx, oldUserInfo.ID, entity.EmailStatusAvailable)
		if err != nil {
			return oldUserInfo.MailStatus, err
		}
	}

	// try to update user avatar
	if oldUserInfo.Avatar == "" && len(externalUserInfo.Avatar) > 0 {
		avatarInfo := &schema.AvatarInfo{
			Type:   constant.AvatarTypeCustom,
			Custom: externalUserInfo.Avatar,
		}
		avatar, _ := json.Marshal(avatarInfo)
		oldUserInfo.Avatar = string(avatar)
		err = us.userRepo.UpdateInfo(ctx, oldUserInfo)
		if err != nil {
			log.Error(err)
		}
	}

	if err = us.userActivity.UserActive(ctx, oldUserInfo.ID); err != nil {
		return oldUserInfo.MailStatus, err
	}
	return entity.EmailStatusAvailable, nil
}

// ExternalLoginBindingUserSendEmail Send an email for third-party account login for binding user
func (us *UserExternalLoginService) ExternalLoginBindingUserSendEmail(
	ctx context.Context, req *schema.ExternalLoginBindingUserSendEmailReq) (
	resp *schema.ExternalLoginBindingUserSendEmailResp, err error) {
	siteGeneral, err := us.siteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return nil, err
	}
	resp = &schema.ExternalLoginBindingUserSendEmailResp{}
	externalLoginInfo, err := us.userExternalLoginRepo.GetCacheUserExternalLoginInfo(ctx, req.BindingKey)
	if err != nil || externalLoginInfo == nil {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	if len(externalLoginInfo.Email) > 0 {
		log.Warnf("the binding email has been sent %s", req.BindingKey)
		return &schema.ExternalLoginBindingUserSendEmailResp{}, nil
	}

	userInfo, exist, err := us.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exist && !req.Must {
		resp.EmailExistAndMustBeConfirmed = true
		return resp, nil
	}

	if !exist {
		externalLoginInfo.Email = req.Email
		userInfo, err = us.registerNewUser(ctx, externalLoginInfo)
		if err != nil {
			return nil, err
		}
		resp.AccessToken, _, err = us.userCommonService.CacheLoginUserInfo(
			ctx, userInfo.ID, userInfo.MailStatus, userInfo.Status, externalLoginInfo.ExternalID)
		if err != nil {
			log.Error(err)
		}
	}
	err = us.userExternalLoginRepo.SetCacheUserExternalLoginInfo(ctx, req.BindingKey, externalLoginInfo)
	if err != nil {
		return nil, err
	}

	// send bind confirmation email
	data := &schema.EmailCodeContent{
		SourceType: schema.BindingSourceType,
		Email:      req.Email,
		UserID:     userInfo.ID,
		BindingKey: req.BindingKey,
	}
	code := uuid.NewString()
	verifyEmailURL := fmt.Sprintf("%s/users/account-activation?code=%s", siteGeneral.SiteUrl, code)
	title, body, err := us.emailService.RegisterTemplate(ctx, verifyEmailURL)
	if err != nil {
		return nil, err
	}
	go us.emailService.SendAndSaveCode(ctx, userInfo.EMail, title, body, code, data.ToJSONString())
	return resp, nil
}

// ExternalLoginBindingUser
// The user clicks on the email link of the bound account and requests the API to bind the user officially
func (us *UserExternalLoginService) ExternalLoginBindingUser(
	ctx context.Context, bindingKey string, oldUserInfo *entity.User) (err error) {
	externalLoginInfo, err := us.userExternalLoginRepo.GetCacheUserExternalLoginInfo(ctx, bindingKey)
	if err != nil || externalLoginInfo == nil {
		return errors.BadRequest(reason.UserNotFound)
	}
	return us.bindOldUser(ctx, externalLoginInfo, oldUserInfo)
}

// GetExternalLoginUserInfoList get external login user info list
func (us *UserExternalLoginService) GetExternalLoginUserInfoList(
	ctx context.Context, userID string) (resp []*entity.UserExternalLogin, err error) {
	return us.userExternalLoginRepo.GetUserExternalLoginList(ctx, userID)
}

// ExternalLoginUnbinding external login unbinding
func (us *UserExternalLoginService) ExternalLoginUnbinding(
	ctx context.Context, req *schema.ExternalLoginUnbindingReq) (resp any, err error) {
	// If user has only one external login and never set password, he can't unbind it.
	userInfo, exist, err := us.userRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	if len(userInfo.Pass) == 0 {
		loginList, err := us.userExternalLoginRepo.GetUserExternalLoginList(ctx, req.UserID)
		if err != nil {
			return nil, err
		}
		if len(loginList) <= 1 {
			return schema.ErrTypeToast, errors.BadRequest(reason.UserExternalLoginUnbindingForbidden)
		}
	}

	return nil, us.userExternalLoginRepo.DeleteUserExternalLogin(ctx, req.UserID, req.ExternalID)
}

// CheckUserStatusInUserCenter check user status in user center
func (us *UserExternalLoginService) CheckUserStatusInUserCenter(ctx context.Context, userID string) (
	valid bool, externalID string, err error) {
	// If enable user center plugin, user status should be checked by user center
	userCenter, ok := plugin.GetUserCenter()
	if !ok {
		return true, "", nil
	}
	userInfoList, err := us.GetExternalLoginUserInfoList(ctx, userID)
	if err != nil {
		return false, "", err
	}
	var thisUcUserInfo *entity.UserExternalLogin
	for _, t := range userInfoList {
		if t.Provider == userCenter.Info().SlugName {
			thisUcUserInfo = t
			break
		}
	}
	// If this user not login by user center, no need to check user status
	if thisUcUserInfo == nil {
		return true, "", nil
	}
	userStatus := userCenter.UserStatus(thisUcUserInfo.ExternalID)
	if userStatus == plugin.UserStatusDeleted {
		return false, thisUcUserInfo.ExternalID, nil
	}
	return true, thisUcUserInfo.ExternalID, nil
}
