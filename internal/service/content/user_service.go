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

package content

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/answer/internal/service/event_queue"
	"github.com/apache/answer/pkg/token"
	"time"

	"github.com/apache/answer/internal/base/constant"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/user_notification_config"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/internal/service/export"
	"github.com/apache/answer/internal/service/role"
	"github.com/apache/answer/internal/service/siteinfo_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/internal/service/user_external_login"
	"github.com/apache/answer/pkg/checker"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/crypto/bcrypt"
)

// UserService user service
type UserService struct {
	userCommonService             *usercommon.UserCommon
	userRepo                      usercommon.UserRepo
	userActivity                  activity.UserActiveActivityRepo
	activityRepo                  activity_common.ActivityRepo
	emailService                  *export.EmailService
	authService                   *auth.AuthService
	siteInfoService               siteinfo_common.SiteInfoCommonService
	userRoleService               *role.UserRoleRelService
	userExternalLoginService      *user_external_login.UserExternalLoginService
	userNotificationConfigRepo    user_notification_config.UserNotificationConfigRepo
	userNotificationConfigService *user_notification_config.UserNotificationConfigService
	questionService               *questioncommon.QuestionCommon
	eventQueueService             event_queue.EventQueueService
}

func NewUserService(userRepo usercommon.UserRepo,
	userActivity activity.UserActiveActivityRepo,
	activityRepo activity_common.ActivityRepo,
	emailService *export.EmailService,
	authService *auth.AuthService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	userRoleService *role.UserRoleRelService,
	userCommonService *usercommon.UserCommon,
	userExternalLoginService *user_external_login.UserExternalLoginService,
	userNotificationConfigRepo user_notification_config.UserNotificationConfigRepo,
	userNotificationConfigService *user_notification_config.UserNotificationConfigService,
	questionService *questioncommon.QuestionCommon,
	eventQueueService event_queue.EventQueueService,
) *UserService {
	return &UserService{
		userCommonService:             userCommonService,
		userRepo:                      userRepo,
		userActivity:                  userActivity,
		activityRepo:                  activityRepo,
		emailService:                  emailService,
		authService:                   authService,
		siteInfoService:               siteInfoService,
		userRoleService:               userRoleService,
		userExternalLoginService:      userExternalLoginService,
		userNotificationConfigRepo:    userNotificationConfigRepo,
		userNotificationConfigService: userNotificationConfigService,
		questionService:               questionService,
		eventQueueService:             eventQueueService,
	}
}

// GetUserInfoByUserID get user info by user id
func (us *UserService) GetUserInfoByUserID(ctx context.Context, token, userID string) (
	resp *schema.GetCurrentLoginUserInfoResp, err error) {
	userInfo, exist, err := us.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	if userInfo.Status == entity.UserStatusDeleted {
		return nil, errors.Unauthorized(reason.UnauthorizedError)
	}

	resp = &schema.GetCurrentLoginUserInfoResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.RoleID, err = us.userRoleService.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		log.Error(err)
	}
	resp.Avatar = us.siteInfoService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status)
	resp.AccessToken = token
	resp.HavePassword = len(userInfo.Pass) > 0
	return resp, nil
}

func (us *UserService) GetOtherUserInfoByUsername(ctx context.Context, req *schema.GetOtherUserInfoByUsernameReq) (
	resp *schema.GetOtherUserInfoByUsernameResp, err error) {
	userInfo, exist, err := us.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.NotFound(reason.UserNotFound)
	}
	resp = &schema.GetOtherUserInfoByUsernameResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = us.siteInfoService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()

	// Only the user himself and the administrator can see the hidden questions
	questionCount, err := us.questionService.GetPersonalUserQuestionCount(ctx, req.UserID, userInfo.ID, req.IsAdmin)
	if err != nil {
		return nil, err
	}
	resp.QuestionCount = int(questionCount)
	return resp, nil
}

// EmailLogin email login
func (us *UserService) EmailLogin(ctx context.Context, req *schema.UserEmailLoginReq) (resp *schema.UserLoginResp, err error) {
	siteLogin, err := us.siteInfoService.GetSiteLogin(ctx)
	if err != nil {
		return nil, err
	}
	if !siteLogin.AllowPasswordLogin {
		return nil, errors.BadRequest(reason.NotAllowedLoginViaPassword)
	}
	userInfo, exist, err := us.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if !exist || userInfo.Status == entity.UserStatusDeleted {
		return nil, errors.BadRequest(reason.EmailOrPasswordWrong)
	}
	if !us.verifyPassword(ctx, req.Pass, userInfo.Pass) {
		return nil, errors.BadRequest(reason.EmailOrPasswordWrong)
	}
	ok, externalID, err := us.userExternalLoginService.CheckUserStatusInUserCenter(ctx, userInfo.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.BadRequest(reason.EmailOrPasswordWrong)
	}

	err = us.userRepo.UpdateLastLoginDate(ctx, userInfo.ID)
	if err != nil {
		log.Errorf("update last login data failed, err: %v", err)
	}

	roleID, err := us.userRoleService.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		log.Error(err)
	}

	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = us.siteInfoService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userInfo.ID,
		EmailStatus: userInfo.MailStatus,
		UserStatus:  userInfo.Status,
		RoleID:      roleID,
		ExternalID:  externalID,
	}
	resp.AccessToken, resp.VisitToken, err = us.authService.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return nil, err
	}
	resp.RoleID = userCacheInfo.RoleID
	if resp.RoleID == role.RoleAdminID {
		err = us.authService.SetAdminUserCacheInfo(ctx, resp.AccessToken, userCacheInfo)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// RetrievePassWord .
func (us *UserService) RetrievePassWord(ctx context.Context, req *schema.UserRetrievePassWordRequest) error {
	userInfo, has, err := us.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	// send email
	data := &schema.EmailCodeContent{
		Email:  req.Email,
		UserID: userInfo.ID,
	}
	code := token.GenerateToken()
	verifyEmailURL := fmt.Sprintf("%s/users/password-reset?code=%s", us.getSiteUrl(ctx), code)
	title, body, err := us.emailService.PassResetTemplate(ctx, verifyEmailURL)
	if err != nil {
		return err
	}
	go us.emailService.SendAndSaveCode(ctx, userInfo.ID, req.Email, title, body, code, data.ToJSONString())
	return nil
}

// UpdatePasswordWhenForgot update user password when user forgot password
func (us *UserService) UpdatePasswordWhenForgot(ctx context.Context, req *schema.UserRePassWordRequest) (err error) {
	data := &schema.EmailCodeContent{}
	err = data.FromJSONString(req.Content)
	if err != nil {
		return errors.BadRequest(reason.EmailVerifyURLExpired)
	}

	userInfo, exist, err := us.userRepo.GetByEmail(ctx, data.Email)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}
	enpass, err := us.encryptPassword(ctx, req.Pass)
	if err != nil {
		return err
	}
	err = us.userRepo.UpdatePass(ctx, userInfo.ID, enpass)
	if err != nil {
		return err
	}
	// When the user changes the password, all the current user's tokens are invalid.
	us.authService.RemoveUserAllTokens(ctx, userInfo.ID)
	return nil
}

func (us *UserService) UserModifyPassWordVerification(ctx context.Context, req *schema.UserModifyPasswordReq) (bool, error) {
	userInfo, has, err := us.userRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return false, err
	}
	if !has {
		return false, errors.BadRequest(reason.UserNotFound)
	}
	isPass := us.verifyPassword(ctx, req.OldPass, userInfo.Pass)
	if !isPass {
		return false, nil
	}

	return true, nil
}

// UserModifyPassword user modify password
func (us *UserService) UserModifyPassword(ctx context.Context, req *schema.UserModifyPasswordReq) error {
	enpass, err := us.encryptPassword(ctx, req.Pass)
	if err != nil {
		return err
	}
	userInfo, exist, err := us.userRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}

	isPass := us.verifyPassword(ctx, req.OldPass, userInfo.Pass)
	if !isPass {
		return errors.BadRequest(reason.OldPasswordVerificationFailed)
	}
	err = us.userRepo.UpdatePass(ctx, userInfo.ID, enpass)
	if err != nil {
		return err
	}

	us.authService.RemoveTokensExceptCurrentUser(ctx, userInfo.ID, req.AccessToken)
	return nil
}

// UpdateInfo update user info
func (us *UserService) UpdateInfo(ctx context.Context, req *schema.UpdateInfoRequest) (
	errFields []*validator.FormErrorField, err error) {
	siteUsers, err := us.siteInfoService.GetSiteUsers(ctx)
	if err != nil {
		return nil, err
	}

	if siteUsers.AllowUpdateUsername && len(req.Username) > 0 {
		if checker.IsInvalidUsername(req.Username) {
			return append(errFields, &validator.FormErrorField{
				ErrorField: "username",
				ErrorMsg:   reason.UsernameInvalid,
			}), errors.BadRequest(reason.UsernameInvalid)
		}
		// admin can use reserved username
		if !req.IsAdmin && checker.IsReservedUsername(req.Username) {
			return append(errFields, &validator.FormErrorField{
				ErrorField: "username",
				ErrorMsg:   reason.UsernameInvalid,
			}), errors.BadRequest(reason.UsernameInvalid)
		} else if req.IsAdmin && checker.IsUsersIgnorePath(req.Username) {
			return append(errFields, &validator.FormErrorField{
				ErrorField: "username",
				ErrorMsg:   reason.UsernameInvalid,
			}), errors.BadRequest(reason.UsernameInvalid)
		}

		userInfo, exist, err := us.userRepo.GetByUsername(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		if exist && userInfo.ID != req.UserID {
			return append(errFields, &validator.FormErrorField{
				ErrorField: "username",
				ErrorMsg:   reason.UsernameDuplicate,
			}), errors.BadRequest(reason.UsernameDuplicate)
		}
	}

	oldUserInfo, exist, err := us.userRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	cond := us.formatUserInfoForUpdateInfo(oldUserInfo, req, siteUsers)
	err = us.userRepo.UpdateInfo(ctx, cond)
	if err != nil {
		return nil, err
	}
	us.eventQueueService.Send(ctx, schema.NewEvent(constant.EventUserUpdate, req.UserID))
	return nil, err
}

func (us *UserService) formatUserInfoForUpdateInfo(
	oldUserInfo *entity.User, req *schema.UpdateInfoRequest, siteUsersConf *schema.SiteUsersResp) *entity.User {
	avatar, _ := json.Marshal(req.Avatar)

	userInfo := &entity.User{}
	userInfo.DisplayName = oldUserInfo.DisplayName
	userInfo.Username = oldUserInfo.Username
	userInfo.Avatar = oldUserInfo.Avatar
	userInfo.Bio = oldUserInfo.Bio
	userInfo.BioHTML = oldUserInfo.BioHTML
	userInfo.Website = oldUserInfo.Website
	userInfo.Location = oldUserInfo.Location
	userInfo.ID = req.UserID

	if len(req.DisplayName) > 0 && siteUsersConf.AllowUpdateDisplayName {
		userInfo.DisplayName = req.DisplayName
	}
	if len(req.Username) > 0 && siteUsersConf.AllowUpdateUsername {
		userInfo.Username = req.Username
	}
	if len(avatar) > 0 && siteUsersConf.AllowUpdateAvatar {
		userInfo.Avatar = string(avatar)
	}
	if siteUsersConf.AllowUpdateBio {
		userInfo.Bio = req.Bio
		userInfo.BioHTML = req.BioHTML
	}
	if siteUsersConf.AllowUpdateWebsite {
		userInfo.Website = req.Website
	}
	if siteUsersConf.AllowUpdateLocation {
		userInfo.Location = req.Location
	}
	return userInfo
}

// UserUpdateInterface update user interface
func (us *UserService) UserUpdateInterface(ctx context.Context, req *schema.UpdateUserInterfaceRequest) (err error) {
	return us.userRepo.UpdateUserInterface(ctx, req.UserId, req.Language, req.ColorScheme)
}

// UserRegisterByEmail user register
func (us *UserService) UserRegisterByEmail(ctx context.Context, registerUserInfo *schema.UserRegisterReq) (
	resp *schema.UserLoginResp, errFields []*validator.FormErrorField, err error,
) {
	_, has, err := us.userRepo.GetByEmail(ctx, registerUserInfo.Email)
	if err != nil {
		return nil, nil, err
	}
	if has {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "e_mail",
			ErrorMsg:   reason.EmailDuplicate,
		})
		return nil, errFields, errors.BadRequest(reason.EmailDuplicate)
	}

	userInfo := &entity.User{}
	userInfo.EMail = registerUserInfo.Email
	userInfo.DisplayName = registerUserInfo.Name
	userInfo.Pass, err = us.encryptPassword(ctx, registerUserInfo.Pass)
	if err != nil {
		return nil, nil, err
	}
	userInfo.Username, err = us.userCommonService.MakeUsername(ctx, registerUserInfo.Name)
	if err != nil {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "name",
			ErrorMsg:   reason.UsernameInvalid,
		})
		return nil, errFields, err
	}
	userInfo.IPInfo = registerUserInfo.IP
	userInfo.MailStatus = entity.EmailStatusToBeVerified
	userInfo.Status = entity.UserStatusAvailable
	userInfo.LastLoginDate = time.Now()
	err = us.userRepo.AddUser(ctx, userInfo)
	if err != nil {
		return nil, nil, err
	}
	if err := us.userNotificationConfigService.SetDefaultUserNotificationConfig(ctx, []string{userInfo.ID}); err != nil {
		log.Errorf("set default user notification config failed, err: %v", err)
	}

	// send email
	data := &schema.EmailCodeContent{
		Email:  registerUserInfo.Email,
		UserID: userInfo.ID,
	}
	code := token.GenerateToken()
	verifyEmailURL := fmt.Sprintf("%s/users/account-activation?code=%s", us.getSiteUrl(ctx), code)
	title, body, err := us.emailService.RegisterTemplate(ctx, verifyEmailURL)
	if err != nil {
		return nil, nil, err
	}
	go us.emailService.SendAndSaveCode(ctx, userInfo.ID, userInfo.EMail, title, body, code, data.ToJSONString())

	roleID, err := us.userRoleService.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		log.Error(err)
	}

	// return user info and token
	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = us.siteInfoService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userInfo.ID,
		EmailStatus: userInfo.MailStatus,
		UserStatus:  userInfo.Status,
		RoleID:      roleID,
	}
	resp.AccessToken, resp.VisitToken, err = us.authService.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return nil, nil, err
	}
	resp.RoleID = userCacheInfo.RoleID
	if resp.RoleID == role.RoleAdminID {
		err = us.authService.SetAdminUserCacheInfo(ctx, resp.AccessToken, &entity.UserCacheInfo{UserID: userInfo.ID})
		if err != nil {
			return nil, nil, err
		}
	}
	return resp, nil, nil
}

func (us *UserService) UserVerifyEmailSend(ctx context.Context, userID string) error {
	userInfo, has, err := us.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if !has {
		return errors.BadRequest(reason.UserNotFound)
	}

	data := &schema.EmailCodeContent{
		Email:  userInfo.EMail,
		UserID: userInfo.ID,
	}
	code := token.GenerateToken()
	verifyEmailURL := fmt.Sprintf("%s/users/account-activation?code=%s", us.getSiteUrl(ctx), code)
	title, body, err := us.emailService.RegisterTemplate(ctx, verifyEmailURL)
	if err != nil {
		return err
	}
	go us.emailService.SendAndSaveCode(ctx, userInfo.ID, userInfo.EMail, title, body, code, data.ToJSONString())
	return nil
}

func (us *UserService) UserVerifyEmail(ctx context.Context, req *schema.UserVerifyEmailReq) (resp *schema.UserLoginResp, err error) {
	data := &schema.EmailCodeContent{}
	err = data.FromJSONString(req.Content)
	if err != nil {
		return nil, errors.BadRequest(reason.EmailVerifyURLExpired)
	}

	userInfo, has, err := us.userRepo.GetByEmail(ctx, data.Email)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	if userInfo.MailStatus == entity.EmailStatusToBeVerified {
		userInfo.MailStatus = entity.EmailStatusAvailable
		err = us.userRepo.UpdateEmailStatus(ctx, userInfo.ID, userInfo.MailStatus)
		if err != nil {
			return nil, err
		}
	}
	if err = us.userActivity.UserActive(ctx, userInfo.ID); err != nil {
		log.Error(err)
		return nil, err
	}

	// In the case of three-party login, the associated users are bound
	if len(data.BindingKey) > 0 {
		err = us.userExternalLoginService.ExternalLoginBindingUser(ctx, data.BindingKey, userInfo)
		if err != nil {
			return nil, err
		}
	}

	accessToken, userCacheInfo, err := us.userCommonService.CacheLoginUserInfo(
		ctx, userInfo.ID, userInfo.MailStatus, userInfo.Status, "")
	if err != nil {
		return nil, err
	}

	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = us.siteInfoService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	resp.AccessToken = accessToken
	// User verified email will update user email status. So user status cache should be updated.
	if err = us.authService.SetUserStatus(ctx, userCacheInfo); err != nil {
		return nil, err
	}
	return resp, nil
}

// verifyPassword
// Compare whether the password is correct
func (us *UserService) verifyPassword(ctx context.Context, loginPass, userPass string) bool {
	if len(loginPass) == 0 && len(userPass) == 0 {
		return true
	}
	err := bcrypt.CompareHashAndPassword([]byte(userPass), []byte(loginPass))
	return err == nil
}

// encryptPassword
// The password does irreversible encryption.
func (us *UserService) encryptPassword(ctx context.Context, Pass string) (string, error) {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(Pass), bcrypt.DefaultCost)
	// This encrypted string can be saved to the database and can be used as password matching verification
	return string(hashPwd), err
}

// UserChangeEmailSendCode user change email verification
func (us *UserService) UserChangeEmailSendCode(ctx context.Context, req *schema.UserChangeEmailSendCodeReq) (
	resp []*validator.FormErrorField, err error) {
	userInfo, exist, err := us.userRepo.GetByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	// If user's email already verified, then must verify password first.
	if userInfo.MailStatus == entity.EmailStatusAvailable && !us.verifyPassword(ctx, req.Pass, userInfo.Pass) {
		resp = append(resp, &validator.FormErrorField{
			ErrorField: "pass",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.OldPasswordVerificationFailed),
		})
		return resp, errors.BadRequest(reason.OldPasswordVerificationFailed)
	}

	_, exist, err = us.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exist {
		resp = append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "e_mail",
			ErrorMsg:   translator.Tr(handler.GetLangByCtx(ctx), reason.EmailDuplicate),
		})
		return resp, errors.BadRequest(reason.EmailDuplicate)
	}

	data := &schema.EmailCodeContent{
		Email:  req.Email,
		UserID: req.UserID,
	}
	code := token.GenerateToken()
	var title, body string
	verifyEmailURL := fmt.Sprintf("%s/users/confirm-new-email?code=%s", us.getSiteUrl(ctx), code)
	if userInfo.MailStatus == entity.EmailStatusToBeVerified {
		title, body, err = us.emailService.RegisterTemplate(ctx, verifyEmailURL)
	} else {
		title, body, err = us.emailService.ChangeEmailTemplate(ctx, verifyEmailURL)
	}
	if err != nil {
		return nil, err
	}
	log.Infof("send email confirmation %s", verifyEmailURL)

	go us.emailService.SendAndSaveCode(ctx, userInfo.ID, req.Email, title, body, code, data.ToJSONString())
	return nil, nil
}

// UserChangeEmailVerify user change email verify code
func (us *UserService) UserChangeEmailVerify(ctx context.Context, content string) (resp *schema.UserLoginResp, err error) {
	data := &schema.EmailCodeContent{}
	err = data.FromJSONString(content)
	if err != nil {
		return nil, errors.BadRequest(reason.EmailVerifyURLExpired)
	}

	_, exist, err := us.userRepo.GetByEmail(ctx, data.Email)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errors.BadRequest(reason.EmailDuplicate)
	}

	userInfo, exist, err := us.userRepo.GetByUserID(ctx, data.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	err = us.userRepo.UpdateEmail(ctx, data.UserID, data.Email)
	if err != nil {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	err = us.userRepo.UpdateEmailStatus(ctx, data.UserID, entity.EmailStatusAvailable)
	if err != nil {
		return nil, err
	}
	// if email status is to be verified, active user as well
	if userInfo.MailStatus == entity.EmailStatusToBeVerified {
		if err = us.userActivity.UserActive(ctx, userInfo.ID); err != nil {
			log.Error(err)
			return nil, err
		}
	}

	roleID, err := us.userRoleService.GetUserRole(ctx, userInfo.ID)
	if err != nil {
		log.Error(err)
	}

	resp = &schema.UserLoginResp{}
	resp.ConvertFromUserEntity(userInfo)
	resp.Avatar = us.siteInfoService.FormatAvatar(ctx, userInfo.Avatar, userInfo.EMail, userInfo.Status).GetURL()
	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userInfo.ID,
		EmailStatus: entity.EmailStatusAvailable,
		UserStatus:  userInfo.Status,
		RoleID:      roleID,
	}
	resp.AccessToken, resp.VisitToken, err = us.authService.SetUserCacheInfo(ctx, userCacheInfo)
	if err != nil {
		return nil, err
	}
	// User verified email will update user email status. So user status cache should be updated.
	if err = us.authService.SetUserStatus(ctx, userCacheInfo); err != nil {
		return nil, err
	}
	resp.RoleID = userCacheInfo.RoleID
	if resp.RoleID == role.RoleAdminID {
		err = us.authService.SetAdminUserCacheInfo(ctx, resp.AccessToken, &entity.UserCacheInfo{UserID: userInfo.ID})
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

// getSiteUrl get site url
func (us *UserService) getSiteUrl(ctx context.Context) string {
	siteGeneral, err := us.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site general failed: %s", err)
		return ""
	}
	return siteGeneral.SiteUrl
}

// UserRanking get user ranking
func (us *UserService) UserRanking(ctx context.Context) (resp *schema.UserRankingResp, err error) {
	limit := 20
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -7)
	userIDs, userIDExist := make([]string, 0), make(map[string]bool, 0)

	// get most reputation users
	rankStat, rankStatUserIDs, err := us.getActivityUserRankStat(ctx, startTime, endTime, limit, userIDExist)
	if err != nil {
		return nil, err
	}
	userIDs = append(userIDs, rankStatUserIDs...)

	// get most vote users
	voteStat, voteStatUserIDs, err := us.getActivityUserVoteStat(ctx, startTime, endTime, limit, userIDExist)
	if err != nil {
		return nil, err
	}
	userIDs = append(userIDs, voteStatUserIDs...)

	// get all staff members
	userRoleRels, staffUserIDs, err := us.getStaff(ctx, userIDExist)
	if err != nil {
		return nil, err
	}
	userIDs = append(userIDs, staffUserIDs...)

	// get user information
	userInfoMapping, err := us.getUserInfoMapping(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return us.warpStatRankingResp(userInfoMapping, rankStat, voteStat, userRoleRels), nil
}

// GetUserStaff get user staff
func (us *UserService) GetUserStaff(ctx context.Context, req *schema.GetUserStaffReq) (
	resp []*schema.GetUserStaffResp, err error) {
	userList, err := us.userRepo.SearchUserListByName(ctx, req.Username, req.PageSize, true)
	if err != nil {
		return nil, err
	}
	avatarMapping := us.siteInfoService.FormatListAvatar(ctx, userList)
	for _, u := range userList {
		resp = append(resp, &schema.GetUserStaffResp{
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Avatar:      avatarMapping[u.ID].GetURL(),
		})
	}
	return resp, nil
}

// UserUnsubscribeNotification user unsubscribe email notification
func (us *UserService) UserUnsubscribeNotification(
	ctx context.Context, req *schema.UserUnsubscribeNotificationReq) (err error) {
	data := &schema.EmailCodeContent{}
	err = data.FromJSONString(req.Content)
	if err != nil || len(data.UserID) == 0 {
		return errors.BadRequest(reason.EmailVerifyURLExpired)
	}

	for _, source := range data.NotificationSources {
		notificationConfig, exist, err := us.userNotificationConfigRepo.GetByUserIDAndSource(
			ctx, data.UserID, source)
		if err != nil {
			return err
		}
		if !exist {
			continue
		}
		channels := schema.NewNotificationChannelsFormJson(notificationConfig.Channels)
		// unsubscribe email notification
		for _, channel := range channels {
			if channel.Key == constant.EmailChannel {
				channel.Enable = false
			}
		}
		notificationConfig.Channels = channels.ToJsonString()
		if err = us.userNotificationConfigRepo.Save(ctx, notificationConfig); err != nil {
			return err
		}
	}
	return nil
}

func (us *UserService) getActivityUserRankStat(ctx context.Context, startTime, endTime time.Time, limit int,
	userIDExist map[string]bool) (rankStat []*entity.ActivityUserRankStat, userIDs []string, err error) {
	if plugin.RankAgentEnabled() {
		return make([]*entity.ActivityUserRankStat, 0), make([]string, 0), nil
	}
	rankStat, err = us.activityRepo.GetUsersWhoHasGainedTheMostReputation(ctx, startTime, endTime, limit)
	if err != nil {
		return nil, nil, err
	}
	for _, stat := range rankStat {
		if stat.Rank <= 0 {
			continue
		}
		if userIDExist[stat.UserID] {
			continue
		}
		userIDs = append(userIDs, stat.UserID)
		userIDExist[stat.UserID] = true
	}
	return rankStat, userIDs, nil
}

func (us *UserService) getActivityUserVoteStat(ctx context.Context, startTime, endTime time.Time, limit int,
	userIDExist map[string]bool) (voteStat []*entity.ActivityUserVoteStat, userIDs []string, err error) {
	if plugin.RankAgentEnabled() {
		return make([]*entity.ActivityUserVoteStat, 0), make([]string, 0), nil
	}
	voteStat, err = us.activityRepo.GetUsersWhoHasVoteMost(ctx, startTime, endTime, limit)
	if err != nil {
		return nil, nil, err
	}
	for _, stat := range voteStat {
		if stat.VoteCount <= 0 {
			continue
		}
		if userIDExist[stat.UserID] {
			continue
		}
		userIDs = append(userIDs, stat.UserID)
		userIDExist[stat.UserID] = true
	}
	return voteStat, userIDs, nil
}

func (us *UserService) getStaff(ctx context.Context, userIDExist map[string]bool) (
	userRoleRels []*entity.UserRoleRel, userIDs []string, err error) {
	userRoleRels, err = us.userRoleService.GetUserByRoleID(ctx, []int{role.RoleAdminID, role.RoleModeratorID})
	if err != nil {
		return nil, nil, err
	}
	for _, rel := range userRoleRels {
		if userIDExist[rel.UserID] {
			continue
		}
		userIDs = append(userIDs, rel.UserID)
		userIDExist[rel.UserID] = true
	}
	return userRoleRels, userIDs, nil
}

func (us *UserService) getUserInfoMapping(ctx context.Context, userIDs []string) (
	userInfoMapping map[string]*entity.User, err error) {
	userInfoMapping = make(map[string]*entity.User, 0)
	if len(userIDs) == 0 {
		return userInfoMapping, nil
	}
	userInfoList, err := us.userRepo.BatchGetByID(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	avatarMapping := us.siteInfoService.FormatListAvatar(ctx, userInfoList)
	for _, user := range userInfoList {
		user.Avatar = avatarMapping[user.ID].GetURL()
		userInfoMapping[user.ID] = user
	}
	return userInfoMapping, nil
}

func (us *UserService) SearchUserListByName(ctx context.Context, req *schema.GetOtherUserInfoByUsernameReq) (
	resp []*schema.UserBasicInfo, err error) {
	resp = make([]*schema.UserBasicInfo, 0)
	if len(req.Username) == 0 {
		return resp, nil
	}
	userList, err := us.userRepo.SearchUserListByName(ctx, req.Username, 5, false)
	if err != nil {
		return resp, err
	}
	avatarMapping := us.siteInfoService.FormatListAvatar(ctx, userList)
	for _, u := range userList {
		if req.UserID == u.ID {
			continue
		}
		basicInfo := us.userCommonService.FormatUserBasicInfo(ctx, u)
		basicInfo.Avatar = avatarMapping[u.ID].GetURL()
		resp = append(resp, basicInfo)
	}
	return resp, nil
}

func (us *UserService) warpStatRankingResp(
	userInfoMapping map[string]*entity.User,
	rankStat []*entity.ActivityUserRankStat,
	voteStat []*entity.ActivityUserVoteStat,
	userRoleRels []*entity.UserRoleRel) (resp *schema.UserRankingResp) {
	resp = &schema.UserRankingResp{
		UsersWithTheMostReputation: make([]*schema.UserRankingSimpleInfo, 0),
		UsersWithTheMostVote:       make([]*schema.UserRankingSimpleInfo, 0),
		Staffs:                     make([]*schema.UserRankingSimpleInfo, 0),
	}
	for _, stat := range rankStat {
		if stat.Rank <= 0 {
			continue
		}
		if userInfo := userInfoMapping[stat.UserID]; userInfo != nil && userInfo.Status != entity.UserStatusDeleted {
			resp.UsersWithTheMostReputation = append(resp.UsersWithTheMostReputation, &schema.UserRankingSimpleInfo{
				Username:    userInfo.Username,
				Rank:        stat.Rank,
				DisplayName: userInfo.DisplayName,
				Avatar:      userInfo.Avatar,
			})
		}
	}
	for _, stat := range voteStat {
		if stat.VoteCount <= 0 {
			continue
		}
		if userInfo := userInfoMapping[stat.UserID]; userInfo != nil && userInfo.Status != entity.UserStatusDeleted {
			resp.UsersWithTheMostVote = append(resp.UsersWithTheMostVote, &schema.UserRankingSimpleInfo{
				Username:    userInfo.Username,
				VoteCount:   stat.VoteCount,
				DisplayName: userInfo.DisplayName,
				Avatar:      userInfo.Avatar,
			})
		}
	}
	for _, rel := range userRoleRels {
		if userInfo := userInfoMapping[rel.UserID]; userInfo != nil && userInfo.Status != entity.UserStatusDeleted {
			resp.Staffs = append(resp.Staffs, &schema.UserRankingSimpleInfo{
				Username:    userInfo.Username,
				Rank:        userInfo.Rank,
				DisplayName: userInfo.DisplayName,
				Avatar:      userInfo.Avatar,
			})
		}
	}
	return resp
}
