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
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
)

type EmbedController struct {
}

func NewEmbedController() *EmbedController {
	return &EmbedController{}
}

// GetEmbedConfig get embed plugin config
// @Summary get embed plugin config
// @Description get embed plugin config
// @Tags Plugin
// @Accept json
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]plugin.EmbedConfig}
// @Router /answer/api/v1/embed/config [get]
func (c *EmbedController) GetEmbedConfig(ctx *gin.Context) {
	resp := make([]*plugin.EmbedConfig, 0)

	err := plugin.CallEmbed(func(embed plugin.Embed) (err error) {
		resp, err = embed.GetEmbedConfigs(ctx)
		return err
	})

	handler.HandleResponse(ctx, err, resp)
}
