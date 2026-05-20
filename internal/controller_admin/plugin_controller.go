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
	"encoding/json"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/plugin_common"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
)

// PluginController role controller
type PluginController struct {
	pluginCommonService *plugin_common.PluginCommonService
}

// NewPluginController new controller
func NewPluginController(pluginCommonService *plugin_common.PluginCommonService) *PluginController {
	return &PluginController{pluginCommonService: pluginCommonService}
}

// GetAllPluginStatus get all plugins status
// @Summary get all plugins status
// @Description get all plugins status
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Success 200 {object} handler.RespBody{data=[]schema.GetPluginListResp}
// @Router /answer/api/v1/plugin/status [get]
func (pc *PluginController) GetAllPluginStatus(ctx *gin.Context) {
	resp := make([]*schema.GetAllPluginStatusResp, 0)
	_ = plugin.CallBase(func(base plugin.Base) error {
		info := base.Info()
		resp = append(resp, &schema.GetAllPluginStatusResp{
			SlugName: info.SlugName,
			Enabled:  plugin.StatusManager.IsEnabled(info.SlugName),
		})
		return nil
	})
	handler.HandleResponse(ctx, nil, resp)
}

// GetPluginList get plugin list
// @Summary get plugin list
// @Description get plugin list
// @Tags AdminPlugin
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param status query string false "status: active/inactive"
// @Param have_config query boolean false "have config"
// @Success 200 {object} handler.RespBody{data=[]schema.GetPluginListResp}
// @Router /answer/admin/api/plugins [get]
func (pc *PluginController) GetPluginList(ctx *gin.Context) {
	req := &schema.GetPluginListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	pluginConfigMapping := make(map[string]bool)
	_ = plugin.CallConfig(func(fn plugin.Config) error {
		if len(fn.ConfigFields()) > 0 {
			pluginConfigMapping[fn.Info().SlugName] = true
		}
		return nil
	})

	resp := make([]*schema.GetPluginListResp, 0)
	_ = plugin.CallBase(func(base plugin.Base) error {
		info := base.Info()
		resp = append(resp, &schema.GetPluginListResp{
			Name:        info.Name.Translate(ctx),
			SlugName:    info.SlugName,
			Description: info.Description.Translate(ctx),
			Version:     info.Version,
			Enabled:     plugin.StatusManager.IsEnabled(info.SlugName),
			HaveConfig:  pluginConfigMapping[info.SlugName],
			Link:        info.Link,
		})
		return nil
	})

	if len(req.Status) > 0 {
		resp = pc.filterPluginByStatus(resp, req.Status)
	}
	if req.HaveConfig {
		resp = pc.filterNoConfigPlugin(resp)
	}
	handler.HandleResponse(ctx, nil, resp)
}

func (pc *PluginController) filterNoConfigPlugin(list []*schema.GetPluginListResp) []*schema.GetPluginListResp {
	resp := make([]*schema.GetPluginListResp, 0)
	for _, t := range list {
		if t.HaveConfig {
			resp = append(resp, t)
		}
	}
	return resp
}

func (pc *PluginController) filterPluginByStatus(list []*schema.GetPluginListResp, status schema.PluginStatus,
) []*schema.GetPluginListResp {
	resp := make([]*schema.GetPluginListResp, 0)
	for _, t := range list {
		if status == schema.PluginStatusActive && t.Enabled {
			resp = append(resp, t)
		} else if status == schema.PluginStatusInactive && !t.Enabled {
			resp = append(resp, t)
		}
	}
	return resp
}

// UpdatePluginStatus update plugin status
// @Summary update plugin status
// @Description update plugin status
// @Tags AdminPlugin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdatePluginStatusReq true "UpdatePluginStatusReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/plugin/status [put]
func (pc *PluginController) UpdatePluginStatus(ctx *gin.Context) {
	req := &schema.UpdatePluginStatusReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	plugin.StatusManager.Enable(req.PluginSlugName, req.Enabled)
	err := pc.pluginCommonService.UpdatePluginStatus(ctx)
	handler.HandleResponse(ctx, err, nil)
}

// GetPluginConfig get plugin config
// @Summary get plugin config
// @Description get plugin config
// @Tags AdminPlugin
// @Security ApiKeyAuth
// @Produce  json
// @Param plugin_slug_name query string true "plugin_slug_name"
// @Success 200 {object} handler.RespBody{data=schema.GetPluginConfigResp}
// @Router /answer/admin/api/plugin/config [get]
func (pc *PluginController) GetPluginConfig(ctx *gin.Context) {
	req := &schema.GetPluginConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp := &schema.GetPluginConfigResp{}
	_ = plugin.CallBase(func(base plugin.Base) error {
		if base.Info().SlugName != req.PluginSlugName {
			return nil
		}
		info := base.Info()
		resp.Name = info.Name.Translate(ctx)
		resp.SlugName = info.SlugName
		resp.Description = info.Description.Translate(ctx)
		resp.Version = info.Version
		return nil
	})

	_ = plugin.CallConfig(func(fn plugin.Config) error {
		if fn.Info().SlugName != req.PluginSlugName {
			return nil
		}
		resp.SetConfigFields(ctx, fn.ConfigFields())
		return nil
	})
	handler.HandleResponse(ctx, nil, resp)
}

// UpdatePluginConfig update plugin config
// @Summary update plugin config
// @Description update plugin config
// @Tags AdminPlugin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdatePluginConfigReq true "UpdatePluginConfigReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/plugin/config [put]
func (pc *PluginController) UpdatePluginConfig(ctx *gin.Context) {
	req := &schema.UpdatePluginConfigReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	configFields, _ := json.Marshal(req.ConfigFields)
	err := plugin.CallConfig(func(fn plugin.Config) error {
		if fn.Info().SlugName == req.PluginSlugName {
			return fn.ConfigReceiver(configFields)
		}
		return nil
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	err = pc.pluginCommonService.UpdatePluginConfig(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
