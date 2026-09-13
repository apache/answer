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
	"context"
	"strings"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/auth"
	pat "github.com/apache/answer/internal/service/personal_access_token"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

const ctxPATKey = "ctxPersonalAccessToken"

type patAuthenticator interface {
	Authenticate(ctx context.Context, rawToken string) (*pat.AuthenticatedToken, error)
}

type patUserLoader interface {
	GetUserCacheInfoByUserID(ctx context.Context, userID string) (*entity.UserCacheInfo, error)
}

type patSecurityLoader interface {
	GetSiteSecurity(ctx context.Context) (*schema.SiteSecurityResp, error)
}

// PATRequestAuthorizer authenticates PAT requests and applies their route scopes.
type PATRequestAuthorizer struct {
	tokens   patAuthenticator
	users    patUserLoader
	security patSecurityLoader
}

func NewPATRequestAuthorizer(
	tokens *pat.Service,
	users *auth.AuthService,
	security siteinfo_common.SiteInfoCommonService,
) *PATRequestAuthorizer {
	return newPATRequestAuthorizer(tokens, users, security)
}

func newPATRequestAuthorizer(
	tokens patAuthenticator,
	users patUserLoader,
	security patSecurityLoader,
) *PATRequestAuthorizer {
	return &PATRequestAuthorizer{tokens: tokens, users: users, security: security}
}

// Resolve reports whether the request presented a PAT. A non-nil user means the PAT passed authentication and scope checks.
func (a *PATRequestAuthorizer) Resolve(ctx *gin.Context) (handled bool, user *entity.UserCacheInfo) {
	rawToken, candidate, validTransport := extractPAT(ctx)
	if !candidate {
		return false, nil
	}
	if !validTransport {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		ctx.Abort()
		return true, nil
	}

	tokenInfo, err := a.tokens.Authenticate(ctx, rawToken)
	if err != nil {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		ctx.Abort()
		return true, nil
	}

	security, err := a.security.GetSiteSecurity(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		ctx.Abort()
		return true, nil
	}
	if !security.PersonalAccessTokensEnabled {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ErrFeatureDisabled), gin.H{"feature": "personal_access_tokens"})
		ctx.Abort()
		return true, nil
	}

	allowed, required := pat.AuthorizeRoute(ctx.Request.Method, ctx.FullPath(), tokenInfo.Scopes)
	if !allowed {
		handler.HandleResponse(ctx, errors.Forbidden(reason.PersonalAccessTokenInsufficientScope), gin.H{"required_scopes": required})
		ctx.Abort()
		return true, nil
	}

	userInfo, err := a.users.GetUserCacheInfoByUserID(ctx, tokenInfo.UserID)
	if err != nil || userInfo == nil {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		ctx.Abort()
		return true, nil
	}
	ctx.Set(ctxPATKey, tokenInfo)
	return true, userInfo
}

func GetPATFromContext(ctx *gin.Context) *pat.AuthenticatedToken {
	value, exists := ctx.Get(ctxPATKey)
	if !exists {
		return nil
	}
	token, _ := value.(*pat.AuthenticatedToken)
	return token
}

func extractPAT(ctx *gin.Context) (rawToken string, candidate bool, validTransport bool) {
	queryToken := ctx.Query("Authorization")
	if strings.HasPrefix(queryToken, pat.TokenPrefix) || strings.HasPrefix(strings.TrimPrefix(queryToken, "Bearer "), pat.TokenPrefix) {
		return "", true, false
	}

	header := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if strings.HasPrefix(header, pat.TokenPrefix) {
		return "", true, false
	}
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false, false
	}
	if !strings.HasPrefix(parts[1], pat.TokenPrefix) {
		return "", false, false
	}
	return parts[1], true, true
}

