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

package controller_admin

import (
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/apikey"
	"github.com/gin-gonic/gin"
)

// AdminAPIKeyController site info controller
type AdminAPIKeyController struct {
	apiKeyService *apikey.APIKeyService
}

// NewAdminAPIKeyController new site info controller
func NewAdminAPIKeyController(apiKeyService *apikey.APIKeyService) *AdminAPIKeyController {
	return &AdminAPIKeyController{
		apiKeyService: apiKeyService,
	}
}

// GetAllAPIKeys get all api keys
// @Summary get all api keys
// @Description get all api keys
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]schema.GetAPIKeyResp}
// @Router /answer/admin/api/api-key/all [get]
func (sc *AdminAPIKeyController) GetAllAPIKeys(ctx *gin.Context) {
	resp, err := sc.apiKeyService.GetAPIKeyList(ctx, &schema.GetAPIKeyReq{})
	handler.HandleResponse(ctx, err, resp)
}

// AddAPIKey add apikey
// @Summary add apikey
// @Description add apikey
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.AddAPIKeyReq true "apikey"
// @Success 200 {object} handler.RespBody{data=schema.AddAPIKeyResp}
// @Router /answer/admin/api/api-key [post]
func (sc *AdminAPIKeyController) AddAPIKey(ctx *gin.Context) {
	req := &schema.AddAPIKeyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := sc.apiKeyService.AddAPIKey(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateAPIKey update apikey
// @Summary update apikey
// @Description update apikey
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.UpdateAPIKeyReq true "apikey"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/api-key [put]
func (sc *AdminAPIKeyController) UpdateAPIKey(ctx *gin.Context) {
	req := &schema.UpdateAPIKeyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	err := sc.apiKeyService.UpdateAPIKey(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// DeleteAPIKey delete apikey
// @Summary delete apikey
// @Description delete apikey
// @Security ApiKeyAuth
// @Tags admin
// @Param data body schema.DeleteAPIKeyReq true "apikey"
// @Produce json
// @Success 200 {object} handler.RespBody{}
// @Router /answer/admin/api/api-key [delete]
func (sc *AdminAPIKeyController) DeleteAPIKey(ctx *gin.Context) {
	req := &schema.DeleteAPIKeyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	err := sc.apiKeyService.DeleteAPIKey(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
