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
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/badge"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
)

type BadgeController struct {
	badgeService      *badge.BadgeService
	badgeAwardService *badge.BadgeAwardService
}

func NewBadgeController(
	badgeService *badge.BadgeService,
	badgeAwardService *badge.BadgeAwardService) *BadgeController {
	return &BadgeController{
		badgeService:      badgeService,
		badgeAwardService: badgeAwardService,
	}
}

// GetBadgeList list all badges
// @Summary list all badges group by group
// @Description list all badges group by group
// @Tags api-badge
// @Accept json
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]schema.GetBadgeListResp}
// @Router /answer/api/v1/badges [get]
func (b *BadgeController) GetBadgeList(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := b.badgeService.ListByGroup(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// GetBadgeInfo get badge info
// @Summary get badge info
// @Description get badge info
// @Tags api-badge
// @Accept json
// @Produce json
// @Param id query string true "id" default(string)
// @Success 200 {object} handler.RespBody{data=schema.GetBadgeInfoResp}
// @Router /answer/api/v1/badge [get]
func (b *BadgeController) GetBadgeInfo(ctx *gin.Context) {
	id := ctx.Query("id")
	id = uid.DeShortID(id)

	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := b.badgeService.GetBadgeInfo(ctx, id, userID)
	handler.HandleResponse(ctx, err, resp)
}

// GetBadgeAwardList get badge award list
// @Summary get badge award list
// @Description get badge award list
// @Tags api-badge
// @Accept json
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param badge_id query string true "badge id"
// @Param username query string false "only list the award by username"
// @Success 200 {object} handler.RespBody{data=schema.GetBadgeInfoResp}
// @Router /answer/api/v1/badge/awards/page [get]
func (b *BadgeController) GetBadgeAwardList(ctx *gin.Context) {
	req := &schema.GetBadgeAwardWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.BadgeID = uid.DeShortID(req.BadgeID)

	resp, total, err := b.badgeAwardService.GetBadgeAwardList(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, pager.NewPageModel(total, resp))
}

// GetAllBadgeAwardListByUsername get user badge award list
// @Summary get user badge award list
// @Description get user badge award list
// @Tags api-badge
// @Accept json
// @Produce json
// @Param username query string true "user name"
// @Success 200 {object} handler.RespBody{data=[]schema.GetUserBadgeAwardListResp}
// @Router /answer/api/v1/badge/user/awards [get]
func (b *BadgeController) GetAllBadgeAwardListByUsername(ctx *gin.Context) {
	req := &schema.GetUserBadgeAwardListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, total, err := b.badgeAwardService.GetUserBadgeAwardList(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	handler.HandleResponse(ctx, nil, pager.NewPageModel(total, resp))
}

// GetRecentBadgeAwardListByUsername get user badge award list
// @Summary get user badge award list
// @Description get user badge award list
// @Tags api-badge
// @Accept json
// @Produce json
// @Param username query string true "user name"
// @Success 200 {object} handler.RespBody{data=[]schema.GetUserBadgeAwardListResp}
// @Router /answer/api/v1/badge/user/awards/recent [get]
func (b *BadgeController) GetRecentBadgeAwardListByUsername(ctx *gin.Context) {
	req := &schema.GetUserBadgeAwardListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.Limit = 10

	resp, total, err := b.badgeAwardService.GetUserRecentBadgeAwardList(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	handler.HandleResponse(ctx, nil, pager.NewPageModel(total, resp))
}
