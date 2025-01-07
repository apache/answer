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
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/role"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
)

type ActivityController struct {
	activityService *activity.ActivityService
}

// NewActivityController new activity controller.
func NewActivityController(
	activityService *activity.ActivityService) *ActivityController {
	return &ActivityController{activityService: activityService}
}

// GetObjectTimeline get object timeline
// @Summary get object timeline
// @Description get object timeline
// @Tags Comment
// @Produce json
// @Param object_id query string false "object id"
// @Param tag_slug_name query string false "tag slug name"
// @Param object_type query string false "object type" Enums(question, answer, tag)
// @Param show_vote query boolean false "is show vote"
// @Success 200 {object} handler.RespBody{data=schema.GetObjectTimelineResp}
// @Router /answer/api/v1/activity/timeline [get]
func (ac *ActivityController) GetObjectTimeline(ctx *gin.Context) {
	req := &schema.GetObjectTimelineReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	if userInfo := middleware.GetUserInfoFromContext(ctx); userInfo != nil {
		req.IsAdmin = userInfo.RoleID == role.RoleAdminID
	}

	resp, err := ac.activityService.GetObjectTimeline(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetObjectTimelineDetail get object timeline detail
// @Summary get object timeline detail
// @Description get object timeline detail
// @Tags Comment
// @Produce json
// @Param revision_id query string true "revision id"
// @Success 200 {object} handler.RespBody{data=schema.GetObjectTimelineResp}
// @Router /answer/api/v1/activity/timeline/detail [get]
func (ac *ActivityController) GetObjectTimelineDetail(ctx *gin.Context) {
	req := &schema.GetObjectTimelineDetailReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := ac.activityService.GetObjectTimelineDetail(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
