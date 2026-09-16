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
	"github.com/apache/answer/internal/service/activity_log"
	"github.com/gin-gonic/gin"
)

// ActivityLogController page-view beacon from the UI
type ActivityLogController struct {
	activityLogAdminService *activity_log.ActivityLogAdminService
}

// NewActivityLogController new controller
func NewActivityLogController(activityLogAdminService *activity_log.ActivityLogAdminService) *ActivityLogController {
	return &ActivityLogController{activityLogAdminService: activityLogAdminService}
}

// PageView record that the logged-in user opened a page
// @Summary record a page view
// @Tags activity-log
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.ActivityLogPageViewReq true "page"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/activity-log/view [post]
func (c *ActivityLogController) PageView(ctx *gin.Context) {
	req := &schema.ActivityLogPageViewReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	c.activityLogAdminService.PageView(ctx, req)
	handler.HandleResponse(ctx, nil, nil)
}
