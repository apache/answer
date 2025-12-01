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

package user_admin

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	"github.com/apache/answer/internal/service/badge"
	"github.com/apache/answer/internal/service/comment_common"
	"github.com/apache/answer/internal/service/export"
	notificationcommon "github.com/apache/answer/internal/service/notification_common"
	"github.com/apache/answer/internal/service/plugin_common"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/pkg/token"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/internal/service/role"
	"github.com/apache/answer/internal/service/siteinfo_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/internal/service/user_external_login"
	"github.com/apache/answer/pkg/checker"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"golang.org/x/crypto/bcrypt"
)

// UserAdminRepo user repository
type UserAdminRepo interface {
	UpdateUserStatus(ctx context.Context, userID string, userStatus, mailStatus int, email string, suspendedUntil time.Time) (err error)
	GetUserInfo(ctx context.Context, userID string) (user *entity.User, exist bool, err error)
	GetUserInfoByEmail(ctx context.Context, email string) (user *entity.User, exist bool, err error)
	GetUserPage(ctx context.Context, page, pageSize int, user *entity.User,
		usernameOrDisplayName string, isStaff bool) (users []*entity.User, total int64, err error)
	AddUser(ctx context.Context, user *entity.User) (err error)
	AddUsers(ctx context.Context, users []*entity.User) (err error)
	UpdateUserPassword(ctx context.Context, userID string, password string) (err error)
	DeletePermanentlyUsers(ctx context.Context) (err error)
	GetExpiredSuspendedUsers(ctx context.Context) (users []*entity.User, err error)
}

// UserAdminService user service
type UserAdminService struct {
	userRepo              UserAdminRepo
	userRoleRelService    *role.UserRoleRelService
	authService           *auth.AuthService
	userCommonService     *usercommon.UserCommon
	userActivity          activity.UserActiveActivityRepo
	siteInfoCommonService siteinfo_common.SiteInfoCommonService
	emailService          *export.EmailService
	questionCommonRepo    questioncommon.QuestionRepo
	answerCommonRepo      answercommon.AnswerRepo
	commentCommonRepo     comment_common.CommentCommonRepo
	userExternalLoginRepo user_external_login.UserExternalLoginRepo
	notificationRepo      notificationcommon.NotificationRepo
	pluginUserConfigRepo  plugin_common.PluginUserConfigRepo
	badgeAwardRepo        badge.BadgeAwardRepo
}

// NewUserAdminService new user admin service
func NewUserAdminService(
	userRepo UserAdminRepo,
	userRoleRelService *role.UserRoleRelService,
	authService *auth.AuthService,
	userCommonService *usercommon.UserCommon,
	userActivity activity.UserActiveActivityRepo,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
	emailService *export.EmailService,
	questionCommonRepo questioncommon.QuestionRepo,
	answerCommonRepo answercommon.AnswerRepo,
	commentCommonRepo comment_common.CommentCommonRepo,
	userExternalLoginRepo user_external_login.UserExternalLoginRepo,
	notificationRepo notificationcommon.NotificationRepo,
	pluginUserConfigRepo plugin_common.PluginUserConfigRepo,
	badgeAwardRepo badge.BadgeAwardRepo,
) *UserAdminService {
	return &UserAdminService{
		userRepo:              userRepo,
		userRoleRelService:    userRoleRelService,
		authService:           authService,
		userCommonService:     userCommonService,
		userActivity:          userActivity,
		siteInfoCommonService: siteInfoCommonService,
		emailService:          emailService,
		questionCommonRepo:    questionCommonRepo,
		answerCommonRepo:      answerCommonRepo,
		commentCommonRepo:     commentCommonRepo,
		userExternalLoginRepo: userExternalLoginRepo,
		notificationRepo:      notificationRepo,
		pluginUserConfigRepo:  pluginUserConfigRepo,
		badgeAwardRepo:        badgeAwardRepo,
	}
}

// UpdateUserStatus update user
func (us *UserAdminService) UpdateUserStatus(ctx context.Context, req *schema.UpdateUserStatusReq) (err error) {
	// Admin cannot modify their status
	if req.UserID == req.LoginUserID {
		return errors.BadRequest(reason.AdminCannotModifySelfStatus)
	}
	userInfo, exist, err := us.userRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}
	// if user status is deleted
	if userInfo.Status == entity.UserStatusDeleted {
		return nil
	}

	if req.IsInactive() {
		userInfo.MailStatus = entity.EmailStatusToBeVerified
	}
	if req.IsDeleted() {
		userInfo.Status = entity.UserStatusDeleted
		userInfo.EMail = fmt.Sprintf("%s.%d", userInfo.EMail, time.Now().Unix())
	}
	if req.IsSuspended() {
		userInfo.Status = entity.UserStatusSuspended
	}
	if req.IsNormal() {
		userInfo.Status = entity.UserStatusAvailable
		userInfo.MailStatus = entity.EmailStatusAvailable
	}

	suspendedUntil := req.GetSuspendedUntil()
	err = us.userRepo.UpdateUserStatus(ctx, userInfo.ID, userInfo.Status, userInfo.MailStatus, userInfo.EMail, suspendedUntil)
	if err != nil {
		return err
	}

	// remove all content that user created, such as question, answer, comment, etc.
	if req.RemoveAllContent {
		us.removeAllUserCreatedContent(ctx, userInfo.ID)
	}

	if req.IsDeleted() {
		us.removeAllUserConfiguration(ctx, userInfo.ID)
	}

	// if user reputation is zero means this user is inactive, so try to activate this user.
	if req.IsNormal() && userInfo.Rank == 0 {
		return us.userActivity.UserActive(ctx, userInfo.ID)
	}
	return nil
}

// removeAllUserConfiguration remove all user configuration
func (us *UserAdminService) removeAllUserConfiguration(ctx context.Context, userID string) {
	err := us.userExternalLoginRepo.DeleteUserExternalLoginByUserID(ctx, userID)
	if err != nil {
		log.Errorf("remove all user external login error: %v", err)
	}
	err = us.notificationRepo.DeleteNotification(ctx, userID)
	if err != nil {
		log.Errorf("remove all user notification error: %v", err)
	}
	err = us.notificationRepo.DeleteUserNotificationConfig(ctx, userID)
	if err != nil {
		log.Errorf("remove all user notification config error: %v", err)
	}
	err = us.pluginUserConfigRepo.DeleteUserPluginConfig(ctx, userID)
	if err != nil {
		log.Errorf("remove all user plugin config error: %v", err)
	}
	err = us.badgeAwardRepo.DeleteUserBadgeAward(ctx, userID)
	if err != nil {
		log.Errorf("remove all user badge award error: %v", err)
	}
}

// removeAllUserCreatedContent remove all user created content
func (us *UserAdminService) removeAllUserCreatedContent(ctx context.Context, userID string) {
	if err := us.questionCommonRepo.RemoveAllUserQuestion(ctx, userID); err != nil {
		log.Errorf("remove all user question error: %v", err)
	}
	if err := us.answerCommonRepo.RemoveAllUserAnswer(ctx, userID); err != nil {
		log.Errorf("remove all user answer error: %v", err)
	}
	if err := us.commentCommonRepo.RemoveAllUserComment(ctx, userID); err != nil {
		log.Errorf("remove all user comment error: %v", err)
	}
}

// UpdateUserRole update user role
func (us *UserAdminService) UpdateUserRole(ctx context.Context, req *schema.UpdateUserRoleReq) (err error) {
	// Users cannot modify their roles
	if req.UserID == req.LoginUserID {
		return errors.BadRequest(reason.UserCannotUpdateYourRole)
	}

	err = us.userRoleRelService.SaveUserRole(ctx, req.UserID, req.RoleID)
	if err != nil {
		return err
	}

	us.authService.RemoveUserAllTokens(ctx, req.UserID)
	return
}

// AddUser add user
func (us *UserAdminService) AddUser(ctx context.Context, req *schema.AddUserReq) (err error) {
	_, has, err := us.userRepo.GetUserInfoByEmail(ctx, req.Email)
	if err != nil {
		return err
	}
	if has {
		return errors.BadRequest(reason.EmailDuplicate)
	}

	hashPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userInfo := &entity.User{}
	userInfo.EMail = req.Email
	userInfo.DisplayName = req.DisplayName
	userInfo.Pass = string(hashPwd)

	userInfo.Username, err = us.userCommonService.MakeUsername(ctx, userInfo.DisplayName)
	if err != nil {
		return err
	}
	userInfo.MailStatus = entity.EmailStatusAvailable
	userInfo.Status = entity.UserStatusAvailable
	userInfo.Rank = 1

	err = us.userRepo.AddUser(ctx, userInfo)
	if err != nil {
		return err
	}
	return
}

// AddUsers add users
func (us *UserAdminService) AddUsers(ctx context.Context, req *schema.AddUsersReq) (
	resp []*validator.FormErrorField, err error) {
	resp, err = req.ParseUsers(ctx)
	if err != nil {
		return resp, err
	}
	errData := us.checkUserDuplicateInner(ctx, req.Users)
	if errData != nil {
		return errData.GetErrField(ctx), errors.BadRequest(reason.RequestFormatError)
	}
	users, errData, err := us.formatBulkAddUsers(ctx, req)
	if err != nil {
		return resp, err
	}
	if errData != nil {
		return errData.GetErrField(ctx), errors.BadRequest(reason.RequestFormatError)
	}
	return nil, us.userRepo.AddUsers(ctx, users)
}

func (us *UserAdminService) checkUserDuplicateInner(ctx context.Context, users []*schema.AddUserReq) (
	errorData *schema.AddUsersErrorData) {
	lang := handler.GetLangByCtx(ctx)
	val := validator.GetValidatorByLang(lang)

	emails := make(map[string]bool)
	displayNames := make(map[string]bool)
	for line, user := range users {
		if errFields, e := val.Check(user); e != nil {
			errorData = &schema.AddUsersErrorData{}
			if len(errFields) > 0 {
				errorData.Field = errFields[0].ErrorField
				errorData.ExtraMessage = errFields[0].ErrorMsg
			}
			errorData.Line = line + 1
			errorData.Content = fmt.Sprintf("%s, %s, %s", user.DisplayName, user.Email, user.Password)
			return errorData
		}
		if emails[user.Email] {
			return &schema.AddUsersErrorData{
				Field:        "email",
				Line:         line + 1,
				Content:      user.Email,
				ExtraMessage: translator.Tr(lang, reason.EmailDuplicate),
			}
		}
		if displayNames[user.DisplayName] {
			return &schema.AddUsersErrorData{
				Field:        "name",
				Line:         line + 1,
				Content:      user.DisplayName,
				ExtraMessage: translator.Tr(lang, reason.UsernameDuplicate),
			}
		}
		emails[user.Email] = true
		displayNames[user.DisplayName] = true
	}
	return nil
}

func (us *UserAdminService) formatBulkAddUsers(ctx context.Context, req *schema.AddUsersReq) (
	users []*entity.User, errorData *schema.AddUsersErrorData, err error) {
	lang := handler.GetLangByCtx(ctx)
	errorData = &schema.AddUsersErrorData{Line: -1}
	for line, user := range req.Users {
		_, has, e := us.userRepo.GetUserInfoByEmail(ctx, user.Email)
		if e != nil {
			return nil, nil, e
		}
		if has {
			errorData.Field = "email"
			errorData.Line = line + 1
			errorData.Content = user.Email
			errorData.ExtraMessage = translator.Tr(lang, reason.EmailDuplicate)
			return nil, errorData, nil
		}

		userInfo := &entity.User{}
		userInfo.EMail = user.Email
		userInfo.DisplayName = user.DisplayName
		hashPwd, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		userInfo.Pass = string(hashPwd)
		userInfo.Username, err = us.userCommonService.MakeUsername(ctx, userInfo.DisplayName)
		if err != nil {
			errorData.Field = "name"
			errorData.Line = line + 1
			errorData.Content = user.DisplayName
			errorData.ExtraMessage = translator.Tr(lang, reason.UsernameInvalid)
			return nil, errorData, nil
		}
		userInfo.MailStatus = entity.EmailStatusAvailable
		userInfo.Status = entity.UserStatusAvailable
		userInfo.Rank = 1
		users = append(users, userInfo)
	}
	return users, nil, nil
}

// UpdateUserPassword update user password
func (us *UserAdminService) UpdateUserPassword(ctx context.Context, req *schema.UpdateUserPasswordReq) (err error) {
	// Users cannot modify their password
	if req.UserID == req.LoginUserID {
		return errors.BadRequest(reason.AdminCannotUpdateTheirPassword)
	}
	userInfo, exist, err := us.userRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}

	hashPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = us.userRepo.UpdateUserPassword(ctx, userInfo.ID, string(hashPwd))
	if err != nil {
		return err
	}
	// logout this user
	us.authService.RemoveUserAllTokens(ctx, req.UserID)
	return
}

// EditUserProfile edit user profile
func (us *UserAdminService) EditUserProfile(ctx context.Context, req *schema.EditUserProfileReq) (
	errFields []*validator.FormErrorField, err error) {
	if req.UserID == req.LoginUserID {
		return nil, errors.BadRequest(reason.AdminCannotEditTheirProfile)
	}
	_, exist, err := us.userRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	if checker.IsInvalidUsername(req.Username) || checker.IsUsersIgnorePath(req.Username) {
		return append(errFields, &validator.FormErrorField{
			ErrorField: "username",
			ErrorMsg:   reason.UsernameInvalid,
		}), errors.BadRequest(reason.UsernameInvalid)
	}

	userInfo, exist, err := us.userCommonService.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exist && userInfo.ID != req.UserID {
		return append(errFields, &validator.FormErrorField{
			ErrorField: "username",
			ErrorMsg:   reason.UsernameDuplicate,
		}), errors.BadRequest(reason.UsernameDuplicate)
	}

	userInfo, exist, err = us.userCommonService.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exist && userInfo.ID != req.UserID {
		return append(errFields, &validator.FormErrorField{
			ErrorField: "email",
			ErrorMsg:   reason.EmailDuplicate,
		}), errors.BadRequest(reason.EmailDuplicate)
	}

	user := &entity.User{}
	user.ID = req.UserID
	user.DisplayName = req.DisplayName
	user.Username = req.Username
	user.EMail = req.Email
	user.MailStatus = entity.EmailStatusAvailable
	err = us.userCommonService.UpdateUserProfile(ctx, user)
	if err != nil {
		return nil, err
	}
	return
}

// GetUserInfo get user one
func (us *UserAdminService) GetUserInfo(ctx context.Context, userID string) (resp *schema.GetUserInfoResp, err error) {
	user, exist, err := us.userRepo.GetUserInfo(ctx, userID)
	if err != nil {
		return
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	resp = &schema.GetUserInfoResp{}
	_ = copier.Copy(resp, user)
	return resp, nil
}

// GetUserPage get user list page
func (us *UserAdminService) GetUserPage(ctx context.Context, req *schema.GetUserPageReq) (pageModel *pager.PageModel, err error) {
	user := &entity.User{}
	_ = copier.Copy(user, req)

	if req.IsInactive() {
		user.MailStatus = entity.EmailStatusToBeVerified
		user.Status = entity.UserStatusAvailable
	} else if req.IsSuspended() {
		user.Status = entity.UserStatusSuspended
	} else if req.IsDeleted() {
		user.Status = entity.UserStatusDeleted
	} else {
		user.MailStatus = entity.EmailStatusAvailable
		user.Status = entity.UserStatusAvailable
	}

	if len(req.Query) > 0 {
		if email, e := mail.ParseAddress(req.Query); e == nil {
			user.EMail = email.Address
			req.Query = ""
		} else if strings.HasPrefix(req.Query, "user:") {
			id := strings.TrimSpace(strings.TrimPrefix(req.Query, "user:"))
			idSearch := true
			for _, r := range id {
				if !unicode.IsDigit(r) {
					idSearch = false
					break
				}
			}
			if idSearch {
				user.ID = id
				req.Query = ""
			} else {
				req.Query = id
			}
		}
	}

	users, total, err := us.userRepo.GetUserPage(ctx, req.Page, req.PageSize, user, req.Query, req.Staff)
	if err != nil {
		return
	}
	avatarMapping := us.siteInfoCommonService.FormatListAvatar(ctx, users)

	resp := make([]*schema.GetUserPageResp, 0)
	for _, u := range users {
		t := &schema.GetUserPageResp{
			UserID:      u.ID,
			CreatedAt:   u.CreatedAt.Unix(),
			Username:    u.Username,
			EMail:       u.EMail,
			Rank:        u.Rank,
			DisplayName: u.DisplayName,
			Avatar:      avatarMapping[u.ID].GetURL(),
		}
		if u.Status == entity.UserStatusDeleted {
			t.Status = constant.UserDeleted
			t.DeletedAt = u.DeletedAt.Unix()
		} else if u.Status == entity.UserStatusSuspended {
			t.Status = constant.UserSuspended
			t.SuspendedAt = u.SuspendedAt.Unix()
			if !u.SuspendedUntil.IsZero() {
				t.SuspendedUntil = u.SuspendedUntil.Unix()
			}
		} else if u.MailStatus == entity.EmailStatusToBeVerified {
			t.Status = constant.UserInactive
		} else {
			t.Status = constant.UserNormal
		}
		resp = append(resp, t)
	}
	us.setUserRoleInfo(ctx, resp)
	return pager.NewPageModel(total, resp), nil
}

func (us *UserAdminService) setUserRoleInfo(ctx context.Context, resp []*schema.GetUserPageResp) {
	var userIDs []string
	for _, u := range resp {
		userIDs = append(userIDs, u.UserID)
	}

	userRoleMapping, err := us.userRoleRelService.GetUserRoleMapping(ctx, userIDs)
	if err != nil {
		log.Error(err)
		return
	}

	for _, u := range resp {
		r := userRoleMapping[u.UserID]
		if r == nil {
			continue
		}
		u.RoleID = r.ID
		u.RoleName = r.Name
	}
}

func (us *UserAdminService) GetUserActivation(ctx context.Context, req *schema.GetUserActivationReq) (
	resp *schema.GetUserActivationResp, err error) {
	userInfo, exist, err := us.userRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	general, err := us.siteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return nil, err
	}

	data := &schema.EmailCodeContent{
		Email:  userInfo.EMail,
		UserID: userInfo.ID,
	}
	code := token.GenerateToken()
	us.emailService.SaveCode(ctx, userInfo.ID, code, data.ToJSONString())
	resp = &schema.GetUserActivationResp{
		ActivationURL: fmt.Sprintf("%s/users/account-activation?code=%s", general.SiteUrl, code),
	}
	return resp, nil
}

// SendUserActivation send user activation email
func (us *UserAdminService) SendUserActivation(ctx context.Context, req *schema.SendUserActivationReq) (err error) {
	userInfo, exist, err := us.userRepo.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.BadRequest(reason.UserNotFound)
	}

	general, err := us.siteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		return err
	}

	data := &schema.EmailCodeContent{
		Email:  userInfo.EMail,
		UserID: userInfo.ID,
	}
	code := token.GenerateToken()
	verifyEmailURL := fmt.Sprintf("%s/users/account-activation?code=%s", general.SiteUrl, code)
	title, body, err := us.emailService.RegisterTemplate(ctx, verifyEmailURL)
	if err != nil {
		return err
	}
	go us.emailService.SendAndSaveCode(ctx, userInfo.ID, userInfo.EMail, title, body, code, data.ToJSONString())
	return nil
}

func (us *UserAdminService) DeletePermanently(ctx context.Context, req *schema.DeletePermanentlyReq) (err error) {
	switch req.Type {
	case constant.DeletePermanentlyUsers:
		return us.userRepo.DeletePermanentlyUsers(ctx)
	case constant.DeletePermanentlyQuestions:
		return us.questionCommonRepo.DeletePermanentlyQuestions(ctx)
	case constant.DeletePermanentlyAnswers:
		return us.answerCommonRepo.DeletePermanentlyAnswers(ctx)
	}

	return errors.BadRequest(reason.RequestFormatError)
}

// CheckAndUnsuspendExpiredUsers checks for users whose suspension has expired and restores them to normal status
func (us *UserAdminService) CheckAndUnsuspendExpiredUsers(ctx context.Context) error {
	// Find all suspended users whose suspension time has expired
	expiredUsers, err := us.userRepo.GetExpiredSuspendedUsers(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, user := range expiredUsers {
		// Check if suspension has expired (not permanent and time has passed)
		if user.Status == entity.UserStatusSuspended &&
			!user.SuspendedUntil.IsZero() &&
			user.SuspendedUntil.Before(now) {

			log.Infof("Unsuspending user %s (ID: %s) - suspension expired at %v",
				user.Username, user.ID, user.SuspendedUntil)

			// Update user status to normal
			err = us.userRepo.UpdateUserStatus(ctx, user.ID, entity.UserStatusAvailable,
				entity.EmailStatusAvailable, user.EMail, time.Time{})
			if err != nil {
				log.Errorf("Failed to unsuspend user %s (ID: %s): %v",
					user.Username, user.ID, err)
				continue
			}
		}
	}

	return nil
}
