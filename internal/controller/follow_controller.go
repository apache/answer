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
	"github.com/apache/answer/internal/service/follow"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

// FollowController activity controller
type FollowController struct {
	followService *follow.FollowService
}

// NewFollowController new controller
func NewFollowController(followService *follow.FollowService) *FollowController {
	return &FollowController{followService: followService}
}

// Follow godoc
// @Summary follow object or cancel follow operation
// @Description follow object or cancel follow operation
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.FollowReq true "follow"
// @Success 200 {object} handler.RespBody{data=schema.FollowResp}
// @Router /answer/api/v1/follow [post]
func (fc *FollowController) Follow(ctx *gin.Context) {
	req := &schema.FollowReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	dto := &schema.FollowDTO{}
	_ = copier.Copy(dto, req)
	dto.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := fc.followService.Follow(ctx, dto)
	if err != nil {
		handler.HandleResponse(ctx, err, schema.ErrTypeToast)
	} else {
		handler.HandleResponse(ctx, err, resp)
	}
}

// UpdateFollowTags update user follow tags
// @Summary update user follow tags
// @Description update user follow tags
// @Tags Activity
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateFollowTagsReq true "follow"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/follow/tags [put]
func (fc *FollowController) UpdateFollowTags(ctx *gin.Context) {
	req := &schema.UpdateFollowTagsReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	err := fc.followService.UpdateFollowTags(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
