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
	"errors"
	"time"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/schema"
	pat "github.com/apache/answer/internal/service/personal_access_token"
	"github.com/apache/answer/internal/service/siteinfo_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/gin-gonic/gin"
	pacmanerrors "github.com/segmentfault/pacman/errors"
)

type PersonalAccessTokenController struct {
	tokens     *pat.Service
	siteInfo   siteinfo_common.SiteInfoCommonService
	userCommon *usercommon.UserCommon
}

func NewPersonalAccessTokenController(
	tokens *pat.Service,
	siteInfo siteinfo_common.SiteInfoCommonService,
	userCommon *usercommon.UserCommon,
) *PersonalAccessTokenController {
	return &PersonalAccessTokenController{tokens: tokens, siteInfo: siteInfo, userCommon: userCommon}
}

// List returns the current user's Personal Access Tokens.
// @Summary List personal access tokens
// @Tags Personal Access Token
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=[]schema.PersonalAccessTokenInfo}
// @Router /answer/api/v1/personal-access-tokens [get]
func (c *PersonalAccessTokenController) List(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	items, err := c.tokens.List(ctx, userID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	security, err := c.siteInfo.GetSiteSecurity(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	resp := make([]schema.PersonalAccessTokenInfo, 0, len(items))
	for _, item := range items {
		converted := convertPATInfo(item)
		if !security.PersonalAccessTokensEnabled && converted.Status == string(pat.StatusActive) {
			converted.Status = "temporarily_unavailable"
		}
		resp = append(resp, converted)
	}
	handler.HandleResponse(ctx, nil, resp)
}

// Create creates a Personal Access Token and returns its secret once.
// @Summary Create a personal access token
// @Tags Personal Access Token
// @Security ApiKeyAuth
// @Param data body schema.PersonalAccessTokenCreateReq true "personal access token"
// @Success 200 {object} handler.RespBody{data=schema.PersonalAccessTokenCreateResp}
// @Router /answer/api/v1/personal-access-tokens [post]
func (c *PersonalAccessTokenController) Create(ctx *gin.Context) {
	request := &schema.PersonalAccessTokenCreateReq{}
	if handler.BindAndCheck(ctx, request) {
		return
	}
	userInfo := middleware.GetUserInfoFromContext(ctx)
	if userInfo == nil {
		handler.HandleResponse(ctx, pacmanerrors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}
	security, err := c.siteInfo.GetSiteSecurity(ctx)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !security.PersonalAccessTokensEnabled {
		handler.HandleResponse(ctx, pacmanerrors.Forbidden(reason.ErrFeatureDisabled), gin.H{"feature": "personal_access_tokens"})
		return
	}
	if userInfo.AuthenticatedAt <= 0 || time.Since(time.Unix(userInfo.AuthenticatedAt, 0)) >
		time.Duration(security.PATReauthenticationWindow())*time.Minute {
		handler.HandleResponse(ctx, pacmanerrors.Forbidden(reason.PersonalAccessTokenReauthenticationRequired), nil)
		return
	}

	created, err := c.tokens.Create(ctx, pat.CreateInput{
		UserID: userInfo.UserID, Name: request.Name, Scopes: request.Scopes,
		ExpiresAt: time.Unix(request.ExpiresAt, 0),
	})
	if errors.Is(err, pat.ErrInvalidScope) || errors.Is(err, pat.ErrInvalidExpiry) {
		handler.HandleResponse(ctx, pacmanerrors.BadRequest(reason.RequestFormatError), nil)
		return
	}
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	resp := &schema.PersonalAccessTokenCreateResp{
		Token: created.Token, PersonalAccessTokenInfo: convertPATInfo(created.Info),
	}
	handler.HandleResponse(ctx, nil, resp)
}

// Revoke permanently revokes one of the current user's Personal Access Tokens.
// @Summary Revoke a personal access token
// @Tags Personal Access Token
// @Security ApiKeyAuth
// @Param id query int true "personal access token id"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/personal-access-tokens [delete]
func (c *PersonalAccessTokenController) Revoke(ctx *gin.Context) {
	request := &schema.PersonalAccessTokenRevokeReq{}
	if handler.BindAndCheck(ctx, request) {
		return
	}
	err := c.tokens.Revoke(ctx, middleware.GetLoginUserIDFromContext(ctx), request.ID)
	handler.HandleResponse(ctx, err, nil)
}

// Current returns metadata for the PAT authenticating the request.
// @Summary Inspect the current personal access token
// @Tags Personal Access Token
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=schema.PersonalAccessTokenCurrentResp}
// @Router /answer/api/v1/personal-access-tokens/current [get]
func (c *PersonalAccessTokenController) Current(ctx *gin.Context) {
	token := middleware.GetPATFromContext(ctx)
	if token == nil {
		handler.HandleResponse(ctx, pacmanerrors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}
	user, exists, err := c.userCommon.GetUserBasicInfoByID(ctx, token.UserID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !exists {
		handler.HandleResponse(ctx, pacmanerrors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}
	resp := &schema.PersonalAccessTokenCurrentResp{
		User: schema.PersonalAccessTokenUserInfo{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName},
		Token: schema.PersonalAccessTokenInfo{
			ID: token.ID, Name: token.Name, TokenSuffix: token.TokenSuffix, Scopes: token.Scopes,
			CreatedAt: token.CreatedAt.Unix(), ExpiresAt: token.ExpiresAt.Unix(), Status: string(pat.StatusActive),
		},
	}
	handler.HandleResponse(ctx, nil, resp)
}

func convertPATInfo(info pat.TokenInfo) schema.PersonalAccessTokenInfo {
	result := schema.PersonalAccessTokenInfo{
		ID: info.ID, Name: info.Name, TokenSuffix: info.TokenSuffix, Scopes: info.Scopes,
		CreatedAt: info.CreatedAt.Unix(), ExpiresAt: info.ExpiresAt.Unix(), Status: string(info.Status),
	}
	if !info.RevokedAt.IsZero() {
		result.RevokedAt = info.RevokedAt.Unix()
	}
	return result
}
