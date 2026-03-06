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
	"strings"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// apiKeyAllowedPrefixes lists the URL path prefixes accessible via API key.
// Routes not matching any prefix require a session token.
var apiKeyAllowedPrefixes = []string{
	"/answer/api/v1/question",
	"/answer/api/v1/answer",
	"/answer/api/v1/comment",
	"/answer/api/v1/tag",
	"/answer/api/v1/search",
	"/answer/api/v1/collection",
	"/answer/api/v1/vote",
	"/answer/api/v1/follow",
	"/answer/api/v1/revisions",
	"/answer/api/v1/chat/completions",
	"/answer/api/v1/ai/conversation",
	"/answer/api/v1/mcp",
}

func isAPIKeyAllowed(path string) bool {
	for _, prefix := range apiKeyAllowedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// AuthSessionOrAPIKey tries session-based auth first, then falls back to API key auth.
// In both cases it injects a UserCacheInfo into the Gin context so that downstream
// handlers can use GetLoginUserIDFromContext() as usual.
func (am *AuthUserMiddleware) AuthSessionOrAPIKey() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ExtractToken(ctx)
		if len(token) == 0 {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		// 1. Try session-based auth
		userInfo, err := am.authService.GetUserCacheInfo(ctx, token)
		if err == nil && userInfo != nil {
			if !am.validateUserStatus(ctx, userInfo) {
				return
			}
			ctx.Set(ctxUUIDKey, userInfo)
			ctx.Next()
			return
		}

		// 2. Fallback to API key auth (only for whitelisted routes)
		if !isAPIKeyAllowed(ctx.Request.URL.Path) {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		isRead := ctx.Request.Method == "GET"
		apiKeyInfo, err := am.authService.GetAPIKeyInfo(ctx, isRead, token)
		if err != nil || apiKeyInfo == nil {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		// Resolve user from the API key's UserID
		userEntity, exist, err := am.userRepo.GetByUserID(ctx, apiKeyInfo.UserID)
		if err != nil || !exist {
			log.Errorf("API key %s references unknown user %s", apiKeyInfo.AccessKey, apiKeyInfo.UserID)
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}
		if userEntity.Status == entity.UserStatusDeleted || userEntity.Status == entity.UserStatusSuspended {
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		roleID, err := am.userRoleService.GetUserRole(ctx, userEntity.ID)
		if err != nil {
			log.Errorf("failed to get role for user %s: %v", userEntity.ID, err)
			handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
			ctx.Abort()
			return
		}

		ctx.Set(ctxUUIDKey, &entity.UserCacheInfo{
			UserID:      userEntity.ID,
			UserStatus:  userEntity.Status,
			EmailStatus: userEntity.MailStatus,
			RoleID:      roleID,
		})
		ctx.Next()
	}
}
