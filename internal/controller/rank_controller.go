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
	"github.com/apache/answer/internal/service/rank"
	"github.com/gin-gonic/gin"
)

// RankController rank controller
type RankController struct {
	rankService *rank.RankService
}

// NewRankController new controller
func NewRankController(
	rankService *rank.RankService) *RankController {
	return &RankController{rankService: rankService}
}

// GetRankPersonalWithPage user personal rank list
// @Summary user personal rank list
// @Description user personal rank list
// @Tags Rank
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param username query string false "username"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetRankPersonalPageResp}}
// @Router /answer/api/v1/personal/rank/page [get]
func (cc *RankController) GetRankPersonalWithPage(ctx *gin.Context) {
	req := &schema.GetRankPersonalWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := cc.rankService.GetRankPersonalPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
