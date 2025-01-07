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
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/badge"
	"github.com/gin-gonic/gin"
)

type BadgeController struct {
	badgeService *badge.BadgeService
}

func NewBadgeController(badgeService *badge.BadgeService) *BadgeController {
	return &BadgeController{
		badgeService: badgeService,
	}
}

// GetBadgeList list all badges by page
// @Summary list all badges by page
// @Description list all badges by page
// @Tags AdminBadge
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param status query string false "badge status" Enums(, active, inactive)
// @Param q query string false "search param"
// @Success 200 {object} handler.RespBody{data=[]schema.GetBadgeListPagedResp}
// @Router /answer/admin/api/badges [get]
func (b *BadgeController) GetBadgeList(ctx *gin.Context) {
	req := &schema.GetBadgeListPagedReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, total, err := b.badgeService.ListPaged(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	handler.HandleResponse(ctx, nil, pager.NewPageModel(total, resp))
}

// UpdateBadgeStatus update badge status
// @Summary update badge status
// @Description update badge status
// @Tags AdminBadge
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateBadgeStatusReq true "UpdateBadgeStatusReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/badge/status [put]
func (b *BadgeController) UpdateBadgeStatus(ctx *gin.Context) {
	req := &schema.UpdateBadgeStatusReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	err := b.badgeService.UpdateStatus(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
