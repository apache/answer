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
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/meta"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
)

type MetaController struct {
	metaService *meta.MetaService
}

func NewMetaController(
	metaService *meta.MetaService,
) *MetaController {
	return &MetaController{
		metaService: metaService,
	}
}

// AddOrUpdateReaction add or update reaction
// @Summary add or update reaction
// @Description update reaction. if not exist, add one
// @Tags Meta
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateReactionReq true "reaction"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/meta/reaction [put]
func (mc *MetaController) AddOrUpdateReaction(ctx *gin.Context) {
	req := &schema.UpdateReactionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := mc.metaService.AddOrUpdateReaction(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetReaction get reaction
// @Summary get reaction
// @Description get reaction for an object
// @Tags Meta
// @Accept json
// @Produce json
// @Param object_id query string true "object_id"
// @Success 200 {object} handler.RespBody{data=schema.ReactionRespItem}
// @Router /answer/api/v1/meta/reaction [get]
func (mc *MetaController) GetReaction(ctx *gin.Context) {
	req := &schema.GetReactionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := mc.metaService.GetReactionByObjectId(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
