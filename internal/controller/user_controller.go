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

package controller

import (
	"net/url"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/action"
	"github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/export"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/internal/service/user_notification_config"
	"github.com/apache/answer/pkg/checker"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// UserController user controller
type UserController struct {
	userService                   *content.UserService
	authService                   *auth.AuthService
	actionService                 *action.CaptchaService
	emailService                  *export.EmailService
	siteInfoCommonService         siteinfo_common.SiteInfoCommonService
	userNotificationConfigService *user_notification_config.UserNotificationConfigService
}

// NewUserController new controller
func NewUserController(
	authService *auth.AuthService,
	userService *content.UserService,
	actionService *action.CaptchaService,
	emailService *export.EmailService,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
	userNotificationConfigService *user_notification_config.UserNotificationConfigService,
) *UserController {
	return &UserController{
		authService:                   authService,
		userService:                   userService,
		actionService:                 actionService,
		emailService:                  emailService,
		siteInfoCommonService:         siteInfoCommonService,
		userNotificationConfigService: userNotificationConfigService,
	}
}

// GetUserInfoByUserID get user info, if user no login response http code is 200, but user info is null
// @Summary GetUserInfoByUserID
// @Description get user info, if user no login response http code is 200, but user info is null
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=schema.GetCurrentLoginUserInfoResp}
// @Router /answer/api/v1/user/info [get]
func (uc *UserController) GetUserInfoByUserID(ctx *gin.Context) {
	token := middleware.ExtractToken(ctx)
	if len(token) == 0 {
		handler.HandleResponse(ctx, nil, nil)
		return
	}

	// if user is no login return null in data
	userInfo, _ := uc.authService.GetUserCacheInfo(ctx, token)
	if userInfo == nil {
		handler.HandleResponse(ctx, nil, nil)
		return
	}

	resp, err := uc.userService.GetUserInfoByUserID(ctx, token, userInfo.UserID)
	uc.setVisitCookies(ctx, userInfo.VisitToken, false)
	handler.HandleResponse(ctx, err, resp)
}

// GetOtherUserInfoByUsername godoc
// @Summary GetOtherUserInfoByUsername
// @Description GetOtherUserInfoByUsername
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param username query string true "username"
// @Success 200 {object} handler.RespBody{data=schema.GetOtherUserInfoResp}
// @Router /answer/api/v1/personal/user/info [get]
func (uc *UserController) GetOtherUserInfoByUsername(ctx *gin.Context) {
	req := &schema.GetOtherUserInfoByUsernameReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)

	resp, err := uc.userService.GetOtherUserInfoByUsername(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UserEmailLogin godoc
// @Summary UserEmailLogin
// @Description UserEmailLogin
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserEmailLoginReq true "UserEmailLogin"
// @Success 200 {object} handler.RespBody{data=schema.UserLoginResp}
// @Router /answer/api/v1/user/login/email [post]
func (uc *UserController) UserEmailLogin(ctx *gin.Context) {
	req := &schema.UserEmailLoginReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := uc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionPassword, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	resp, err := uc.userService.EmailLogin(ctx, req)
	if err != nil {
		_, _ = uc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
		errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "e_mail",
			ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.EmailOrPasswordWrong),
		})
		handler.HandleResponse(ctx, errors.BadRequest(reason.EmailOrPasswordWrong), errFields)
		return
	}
	if !isAdmin {
		uc.actionService.ActionRecordDel(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
	}
	if resp.Status == constant.UserSuspended {
		handler.HandleResponse(ctx, errors.Forbidden(reason.UserSuspended),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeUserSuspended})
		return
	}
	uc.setVisitCookies(ctx, resp.VisitToken, true)
	handler.HandleResponse(ctx, nil, resp)
}

// RetrievePassWord godoc
// @Summary RetrievePassWord
// @Description RetrievePassWord
// @Tags User
// @Accept  json
// @Produce  json
// @Param data body schema.UserRetrievePassWordRequest  true "UserRetrievePassWordRequest"
// @Success 200 {string} string ""
// @Router /answer/api/v1/user/password/reset [post]
func (uc *UserController) RetrievePassWord(ctx *gin.Context) {
	req := &schema.UserRetrievePassWordRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := uc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}
	err := uc.userService.RetrievePassWord(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UseRePassWord godoc
// @Summary UseRePassWord
// @Description UseRePassWord
// @Tags User
// @Accept  json
// @Produce  json
// @Param data body schema.UserRePassWordRequest  true "UserRePassWordRequest"
// @Success 200 {string} string ""
// @Router /answer/api/v1/user/password/replacement [post]
func (uc *UserController) UseRePassWord(ctx *gin.Context) {
	req := &schema.UserRePassWordRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.Content = uc.emailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}

	err := uc.userService.UpdatePasswordWhenForgot(ctx, req)
	uc.actionService.ActionRecordDel(ctx, entity.CaptchaActionPassword, ctx.ClientIP())
	handler.HandleResponse(ctx, err, nil)
}

// UserLogout user logout
// @Summary user logout
// @Description user logout
// @Security ApiKeyAuth
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/logout [get]
func (uc *UserController) UserLogout(ctx *gin.Context) {
	accessToken := middleware.ExtractToken(ctx)
	if len(accessToken) == 0 {
		handler.HandleResponse(ctx, nil, nil)
		return
	}
	_ = uc.authService.RemoveUserCacheInfo(ctx, accessToken)
	_ = uc.authService.RemoveAdminUserCacheInfo(ctx, accessToken)
	visitToken, _ := ctx.Cookie(constant.UserVisitCookiesCacheKey)
	_ = uc.authService.RemoveUserVisitCacheInfo(ctx, visitToken)
	handler.HandleResponse(ctx, nil, nil)
}

// UserRegisterByEmail godoc
// @Summary UserRegisterByEmail
// @Description UserRegisterByEmail
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserRegisterReq true "UserRegisterReq"
// @Success 200 {object} handler.RespBody{data=schema.UserLoginResp}
// @Router /answer/api/v1/user/register/email [post]
func (uc *UserController) UserRegisterByEmail(ctx *gin.Context) {
	// check whether site allow register or not
	siteInfo, err := uc.siteInfoCommonService.GetSiteLogin(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !siteInfo.AllowNewRegistrations || !siteInfo.AllowEmailRegistrations {
		handler.HandleResponse(ctx, errors.BadRequest(reason.NotAllowedRegistration), nil)
		return
	}

	req := &schema.UserRegisterReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	if !checker.EmailInAllowEmailDomain(req.Email, siteInfo.AllowEmailDomains) {
		handler.HandleResponse(ctx, errors.BadRequest(reason.EmailIllegalDomainError), nil)
		return
	}
	req.IP = ctx.ClientIP()
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := uc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, req.IP, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	resp, errFields, err := uc.userService.UserRegisterByEmail(ctx, req)
	if len(errFields) > 0 {
		for _, field := range errFields {
			field.ErrorMsg = translator.
				Tr(handler.GetLang(ctx), field.ErrorMsg)
		}
		handler.HandleResponse(ctx, err, errFields)
	} else {
		handler.HandleResponse(ctx, err, resp)
	}
}

// UserVerifyEmail godoc
// @Summary UserVerifyEmail
// @Description UserVerifyEmail
// @Tags User
// @Accept json
// @Produce json
// @Param code query string true "code" default()
// @Success 200 {object} handler.RespBody{data=schema.UserLoginResp}
// @Router /answer/api/v1/user/email/verification [post]
func (uc *UserController) UserVerifyEmail(ctx *gin.Context) {
	req := &schema.UserVerifyEmailReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.Content = uc.emailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}

	resp, err := uc.userService.UserVerifyEmail(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	uc.actionService.ActionRecordDel(ctx, entity.CaptchaActionEmail, ctx.ClientIP())
	handler.HandleResponse(ctx, err, resp)
}

// UserVerifyEmailSend godoc
// @Summary UserVerifyEmailSend
// @Description UserVerifyEmailSend
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param captcha_id query string false "captcha_id"  default()
// @Param captcha_code query string false "captcha_code"  default()
// @Success 200 {string} string ""
// @Router /answer/api/v1/user/email/verification/send [post]
func (uc *UserController) UserVerifyEmailSend(ctx *gin.Context) {
	req := &schema.UserVerifyEmailSendReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userInfo := middleware.GetUserInfoFromContext(ctx)
	if userInfo == nil {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := uc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEmail, ctx.ClientIP(), req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	err := uc.userService.UserVerifyEmailSend(ctx, userInfo.UserID)
	handler.HandleResponse(ctx, err, nil)
}

// UserModifyPassWord godoc
// @Summary UserModifyPassWord
// @Description UserModifyPassWord
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UserModifyPasswordReq  true "UserModifyPasswordReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/password [put]
func (uc *UserController) UserModifyPassWord(ctx *gin.Context) {
	req := &schema.UserModifyPasswordReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.AccessToken = middleware.ExtractToken(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := uc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEditUserinfo, req.UserID,
			req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
		_, err := uc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionEditUserinfo, req.UserID)
		if err != nil {
			log.Error(err)
		}
	}

	oldPassVerification, err := uc.userService.UserModifyPassWordVerification(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !oldPassVerification {
		errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "old_pass",
			ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.OldPasswordVerificationFailed),
		})
		handler.HandleResponse(ctx, errors.BadRequest(reason.OldPasswordVerificationFailed), errFields)
		return
	}

	if req.OldPass == req.Pass {
		errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "pass",
			ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.NewPasswordSameAsPreviousSetting),
		})
		handler.HandleResponse(ctx, errors.BadRequest(reason.NewPasswordSameAsPreviousSetting), errFields)
		return
	}
	err = uc.userService.UserModifyPassword(ctx, req)
	if err == nil {
		uc.actionService.ActionRecordDel(ctx, entity.CaptchaActionEditUserinfo, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// UserUpdateInfo update user info
// @Summary UserUpdateInfo update user info
// @Description UserUpdateInfo update user info
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "access-token"
// @Param data body schema.UpdateInfoRequest true "UpdateInfoRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/info [put]
func (uc *UserController) UserUpdateInfo(ctx *gin.Context) {
	req := &schema.UpdateInfoRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)
	errFields, err := uc.userService.UpdateInfo(ctx, req)
	for _, field := range errFields {
		field.ErrorMsg = translator.Tr(handler.GetLang(ctx), field.ErrorMsg)
	}
	handler.HandleResponse(ctx, err, errFields)
}

// UserUpdateInterface update user interface config
// @Summary UserUpdateInterface update user interface config
// @Description UserUpdateInterface update user interface config
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "access-token"
// @Param data body schema.UpdateUserInterfaceRequest true "UpdateInfoRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/interface [put]
func (uc *UserController) UserUpdateInterface(ctx *gin.Context) {
	req := &schema.UpdateUserInterfaceRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserId = middleware.GetLoginUserIDFromContext(ctx)
	err := uc.userService.UserUpdateInterface(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// ActionRecord godoc
// @Summary ActionRecord
// @Description ActionRecord
// @Tags User
// @Param action query string true "action" Enums(login, e_mail, find_pass)
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=schema.ActionRecordResp}
// @Router /answer/api/v1/user/action/record [get]
func (uc *UserController) ActionRecord(ctx *gin.Context) {
	req := &schema.ActionRecordReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userinfo := middleware.GetUserInfoFromContext(ctx)
	if userinfo != nil {
		req.UserID = userinfo.UserID
	}
	req.IP = ctx.ClientIP()
	resp := &schema.ActionRecordResp{}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if isAdmin {
		resp.Verify = false
		handler.HandleResponse(ctx, nil, resp)
	} else {
		resp, err := uc.actionService.ActionRecord(ctx, req)
		handler.HandleResponse(ctx, err, resp)
	}

}

// GetUserNotificationConfig get user's notification config
// @Summary get user's notification config
// @Description get user's notification config
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=schema.GetUserNotificationConfigResp}
// @Router /answer/api/v1/user/notification/config [post]
func (uc *UserController) GetUserNotificationConfig(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := uc.userNotificationConfigService.GetUserNotificationConfig(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateUserNotificationConfig update user's notification config
// @Summary update user's notification config
// @Description update user's notification config
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateUserNotificationConfigReq true "UpdateUserNotificationConfigReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/notification/config [put]
func (uc *UserController) UpdateUserNotificationConfig(ctx *gin.Context) {
	req := &schema.UpdateUserNotificationConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	err := uc.userNotificationConfigService.UpdateUserNotificationConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UserChangeEmailSendCode send email to the user email then change their email
// @Summary send email to the user email then change their email
// @Description send email to the user email then change their email
// @Security ApiKeyAuth
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserChangeEmailSendCodeReq true "UserChangeEmailSendCodeReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/email/change/code [post]
func (uc *UserController) UserChangeEmailSendCode(ctx *gin.Context) {
	req := &schema.UserChangeEmailSendCodeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	// If the user is not logged in, the api cannot be used.
	// If the user email is not verified, that also can use this api to modify the email.
	if len(req.UserID) == 0 {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}
	// check whether email allow register or not
	siteInfo, err := uc.siteInfoCommonService.GetSiteLogin(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !checker.EmailInAllowEmailDomain(req.Email, siteInfo.AllowEmailDomains) {
		handler.HandleResponse(ctx, errors.BadRequest(reason.EmailIllegalDomainError), nil)
		return
	}
	isAdmin := middleware.GetUserIsAdminModerator(ctx)

	if !isAdmin {
		captchaPass := uc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEditUserinfo, req.UserID, req.CaptchaID, req.CaptchaCode)
		uc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionEditUserinfo, req.UserID)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	resp, err := uc.userService.UserChangeEmailSendCode(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, resp)
		return
	}
	if !isAdmin {
		uc.actionService.ActionRecordDel(ctx, entity.CaptchaActionEditUserinfo, ctx.ClientIP())
	}

	handler.HandleResponse(ctx, err, nil)
}

// UserChangeEmailVerify user change email verification
// @Summary user change email verification
// @Description user change email verification
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UserChangeEmailVerifyReq true "UserChangeEmailVerifyReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/email [put]
func (uc *UserController) UserChangeEmailVerify(ctx *gin.Context) {
	req := &schema.UserChangeEmailVerifyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.Content = uc.emailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}

	resp, err := uc.userService.UserChangeEmailVerify(ctx, req.Content)
	uc.actionService.ActionRecordDel(ctx, entity.CaptchaActionEmail, ctx.ClientIP())
	handler.HandleResponse(ctx, err, resp)
}

// UserRanking get user ranking
// @Summary get user ranking
// @Description get user ranking
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} handler.RespBody{data=schema.UserRankingResp}
// @Router /answer/api/v1/user/ranking [get]
func (uc *UserController) UserRanking(ctx *gin.Context) {
	resp, err := uc.userService.UserRanking(ctx)
	handler.HandleResponse(ctx, err, resp)
}

// UserStaff get user staff
// @Summary get user staff
// @Description get user staff
// @Tags User
// @Accept json
// @Produce json
// @Param username query string true "username"
// @Param page_size query string true "page_size"
// @Success 200 {object} handler.RespBody{data=schema.GetUserStaffResp}
// @Router /answer/api/v1/user/staff [get]
func (uc *UserController) UserStaff(ctx *gin.Context) {
	req := &schema.GetUserStaffReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := uc.userService.GetUserStaff(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UserUnsubscribeNotification unsubscribe notification
// @Summary unsubscribe notification
// @Description unsubscribe notification
// @Tags User
// @Accept json
// @Produce json
// @Param data body schema.UserUnsubscribeNotificationReq true "UserUnsubscribeNotificationReq"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/user/notification/unsubscribe [put]
func (uc *UserController) UserUnsubscribeNotification(ctx *gin.Context) {
	req := &schema.UserUnsubscribeNotificationReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.Content = uc.emailService.VerifyUrlExpired(ctx, req.Code)
	if len(req.Content) == 0 {
		handler.HandleResponse(ctx, errors.Forbidden(reason.EmailVerifyURLExpired),
			&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeURLExpired})
		return
	}

	err := uc.userService.UserUnsubscribeNotification(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// SearchUserListByName godoc
// @Summary SearchUserListByName
// @Description SearchUserListByName
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param username query string true "username"
// @Success 200 {object} handler.RespBody{data=schema.GetOtherUserInfoResp}
// @Router /answer/api/v1/user/info/search [get]
func (uc *UserController) SearchUserListByName(ctx *gin.Context) {
	req := &schema.GetOtherUserInfoByUsernameReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := uc.userService.SearchUserListByName(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

func (uc *UserController) setVisitCookies(ctx *gin.Context, visitToken string, force bool) {
	if !force {
		cookie, _ := ctx.Cookie(constant.UserVisitCookiesCacheKey)
		// If the cookie is the same as the visitToken, no need to set it again
		if cookie == visitToken {
			return
		}
	}
	general, err := uc.siteInfoCommonService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site general error: %v", err)
		return
	}
	parsedURL, err := url.Parse(general.SiteUrl)
	if err != nil {
		log.Errorf("parse url error: %v", err)
		return
	}
	ctx.SetCookie(constant.UserVisitCookiesCacheKey,
		visitToken, constant.UserVisitCacheTime, "/", parsedURL.Hostname(), true, true)
}
