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

package middleware

import (
	"net/http"
	"strings"

	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/role"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/ui"
	"github.com/gin-gonic/gin"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

var ctxUUIDKey = "ctxUuidKey"

// AuthUserMiddleware auth user middleware
type AuthUserMiddleware struct {
	authService           *auth.AuthService
	siteInfoCommonService siteinfo_common.SiteInfoCommonService
}

// NewAuthUserMiddleware new auth user middleware
func NewAuthUserMiddleware(
	authService *auth.AuthService,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService) *AuthUserMiddleware {
	return &AuthUserMiddleware{
		authService:           authService,
		siteInfoCommonService: siteInfoCommonService,
	}
}

// Auth get token and auth user, set user info to context if user is already login
func (am *AuthUserMiddleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			ctx.Next()
			return
		}
		userInfo, err := am.authService.GetUserCacheInfo(ctx, token)
		if err != nil {
			ctx.Next()
			return
		}
		if userInfo != nil {
			ctx.Set(ctxUUIDKey, userInfo)
		}
		ctx.Next()
	}
}

// EjectUserBySiteInfo if admin config the site can access by nologin user, eject user.
func (am *AuthUserMiddleware) EjectUserBySiteInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mustLogin := false
		siteInfo, _ := am.siteInfoCommonService.GetSiteLogin(ctx)
		if siteInfo != nil {
			mustLogin = siteInfo.LoginRequired
		}
		if !mustLogin {
			ctx.Next()
			return
		}

		// If site in private mode, user must login.
		userInfo := GetUserInfoFromContext(ctx)
		if userInfo == nil {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		// If user is not active, eject user.
		if userInfo.EmailStatus != entity.EmailStatusAvailable {
			handler.HandleResponse(ctx, errors.Forbidden(reason.EmailNeedToBeVerified),
				&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeInactive})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// MustAuthWithoutAccountAvailable auth user info, any login user can access though user is not active.
func (am *AuthUserMiddleware) MustAuthWithoutAccountAvailable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		userInfo, err := am.authService.GetUserCacheInfo(ctx, token)
		if err != nil || userInfo == nil {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		if userInfo.UserStatus == entity.UserStatusDeleted {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		ctx.Set(ctxUUIDKey, userInfo)
		ctx.Next()
	}
}

// MustAuthAndAccountAvailable auth user info and check user status, only allow active user access.
func (am *AuthUserMiddleware) MustAuthAndAccountAvailable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		userInfo, err := am.authService.GetUserCacheInfo(ctx, token)
		if err != nil || userInfo == nil {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		if userInfo.EmailStatus != entity.EmailStatusAvailable {
			handler.HandleResponse(ctx, errors.Forbidden(reason.EmailNeedToBeVerified),
				&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeInactive})
			ctx.Abort()
			return
		}
		if userInfo.UserStatus == entity.UserStatusSuspended {
			handler.HandleResponse(ctx, errors.Forbidden(reason.UserSuspended),
				&schema.ForbiddenResp{Type: schema.ForbiddenReasonTypeUserSuspended})
			ctx.Abort()
			return
		}
		if userInfo.UserStatus == entity.UserStatusDeleted {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		ctx.Set(ctxUUIDKey, userInfo)
		ctx.Next()
	}
}

func (am *AuthUserMiddleware) AdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		userInfo, err := am.authService.GetAdminUserCacheInfo(ctx, token)
		if err != nil || userInfo == nil {
			handler.HandleResponse(ctx, errors.Forbidden(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		if userInfo != nil {
			if userInfo.UserStatus == entity.UserStatusDeleted {
				handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
				ctx.Abort()
				return
			}
			ctx.Set(ctxUUIDKey, userInfo)
		}
		ctx.Next()
	}
}

func (am *AuthUserMiddleware) CheckPrivateMode() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := am.siteInfoCommonService.GetSiteLogin(ctx)
		if err != nil {
			ShowIndexPage(ctx)
			ctx.Abort()
			return
		}
		if resp.LoginRequired {
			ShowIndexPage(ctx)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
func ShowIndexPage(ctx *gin.Context) {
	ctx.Header("content-type", "text/html;charset=utf-8")
	ctx.Header("X-Frame-Options", "DENY")
	file, err := ui.Build.ReadFile("build/index.html")
	if err != nil {
		log.Error(err)
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.String(http.StatusOK, string(file))
}

// GetLoginUserIDFromContext get user id from context
func GetLoginUserIDFromContext(ctx *gin.Context) (userID string) {
	userInfo := GetUserInfoFromContext(ctx)
	if userInfo == nil {
		return ""
	}
	return userInfo.UserID
}

// GetIsAdminFromContext get user is admin from context
func GetIsAdminFromContext(ctx *gin.Context) (isAdmin bool) {
	userInfo := GetUserInfoFromContext(ctx)
	if userInfo == nil {
		return false
	}
	return userInfo.RoleID == role.RoleAdminID
}

// GetUserInfoFromContext get user info from context
func GetUserInfoFromContext(ctx *gin.Context) (u *entity.UserCacheInfo) {
	userInfo, exist := ctx.Get(ctxUUIDKey)
	if !exist {
		return nil
	}
	u, ok := userInfo.(*entity.UserCacheInfo)
	if !ok {
		return nil
	}
	return u
}

func GetUserIsAdminModerator(ctx *gin.Context) (isAdminModerator bool) {
	userInfo, exist := ctx.Get(ctxUUIDKey)
	if !exist {
		return false
	}
	u, ok := userInfo.(*entity.UserCacheInfo)
	if !ok {
		return false
	}
	if u.RoleID == role.RoleAdminID || u.RoleID == role.RoleModeratorID {
		return true
	}
	return false
}

func GetLoginUserIDInt64FromContext(ctx *gin.Context) (userID int64) {
	userIDStr := GetLoginUserIDFromContext(ctx)
	return converter.StringToInt64(userIDStr)
}

// ExtractToken extract token from context
func ExtractToken(ctx *gin.Context) (token string) {
	token = ctx.GetHeader("Authorization")
	if len(token) == 0 {
		token = ctx.Query("Authorization")
	}
	return strings.TrimPrefix(token, "Bearer ")
}
