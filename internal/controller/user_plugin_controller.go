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
	"encoding/json"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/segmentfault/pacman/errors"
	"net/http"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/plugin_common"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
)

// UserPluginController role controller
type UserPluginController struct {
	pluginCommonService *plugin_common.PluginCommonService
}

// NewUserPluginController new controller
func NewUserPluginController(pluginCommonService *plugin_common.PluginCommonService) *UserPluginController {
	return &UserPluginController{pluginCommonService: pluginCommonService}
}

// GetUserPluginList get plugin list that used for user.
// @Summary get plugin list that used for user.
// @Description get plugin list that used for user.
// @Tags UserPlugin
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Success 200 {object} handler.RespBody{data=[]schema.GetUserPluginListResp}
// @Router /answer/api/v1/user/plugin/configs [get]
func (pc *UserPluginController) GetUserPluginList(ctx *gin.Context) {
	resp := make([]*schema.GetUserPluginListResp, 0)
	_ = plugin.CallUserConfig(func(base plugin.UserConfig) error {
		info := base.Info()
		if plugin.StatusManager.IsEnabled(info.SlugName) {
			resp = append(resp, &schema.GetUserPluginListResp{
				Name:     info.Name.Translate(ctx),
				SlugName: info.SlugName,
			})
		}
		return nil
	})
	handler.HandleResponse(ctx, nil, resp)
}

// GetUserPluginConfig get user plugin config
// @Summary get user plugin config
// @Description get user plugin config
// @Tags UserPlugin
// @Security ApiKeyAuth
// @Produce  json
// @Param plugin_slug_name query string true "plugin_slug_name"
// @Success 200 {object} handler.RespBody{data=schema.GetPluginConfigResp}
// @Router /answer/api/v1/user/plugin/config [get]
func (pc *UserPluginController) GetUserPluginConfig(ctx *gin.Context) {
	req := &schema.GetUserPluginConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp := &schema.GetUserPluginConfigResp{}
	_ = plugin.CallUserConfig(func(fn plugin.UserConfig) error {
		if fn.Info().SlugName != req.PluginSlugName {
			return nil
		}
		info := fn.Info()
		resp.Name = info.Name.Translate(ctx)
		resp.SlugName = info.SlugName
		resp.SetConfigFields(ctx, fn.UserConfigFields())
		return nil
	})

	configValue, err := pc.pluginCommonService.GetUserPluginConfig(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if len(configValue) > 0 {
		configValueMapping := make(map[string]any)
		_ = json.Unmarshal([]byte(configValue), &configValueMapping)
		for _, field := range resp.ConfigFields {
			if value, ok := configValueMapping[field.Name]; ok {
				field.Value = value
			}
		}
	}

	handler.HandleResponse(ctx, err, resp)
}

// UpdatePluginUserConfig update user plugin config
// @Summary update user plugin config
// @Description update user plugin config
// @Tags UserPlugin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateUserPluginConfigReq true "UpdatePluginConfigReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/user/plugin/config [put]
func (pc *UserPluginController) UpdatePluginUserConfig(ctx *gin.Context) {
	req := &schema.UpdateUserPluginConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	if !plugin.StatusManager.IsEnabled(req.PluginSlugName) {
		handler.HandleResponse(ctx, errors.New(http.StatusBadRequest, reason.RequestFormatError), nil)
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	configFields, _ := json.Marshal(req.ConfigFields)
	err := plugin.CallUserConfig(func(fn plugin.UserConfig) error {
		if fn.Info().SlugName == req.PluginSlugName {
			return fn.UserConfigReceiver(req.UserID, configFields)
		}
		return nil
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	err = pc.pluginCommonService.UpdatePluginUserConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
