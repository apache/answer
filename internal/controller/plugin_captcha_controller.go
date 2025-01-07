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

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
)

// CaptchaController comment controller
type CaptchaController struct {
}

// NewCaptchaController new controller
func NewCaptchaController() *CaptchaController {
	return &CaptchaController{}
}

type GetCaptchaConfigResp struct {
	SlugName string         `json:"slug_name"`
	Config   map[string]any `json:"config"`
}

// GetCaptchaConfig get captcha config
func (uc *CaptchaController) GetCaptchaConfig(ctx *gin.Context) {
	resp := &GetCaptchaConfigResp{}
	_ = plugin.CallCaptcha(func(fn plugin.Captcha) error {
		resp.SlugName = fn.Info().SlugName
		_ = json.Unmarshal([]byte(fn.GetConfig()), &resp.Config)
		return nil
	})
	handler.HandleResponse(ctx, nil, resp)
}
