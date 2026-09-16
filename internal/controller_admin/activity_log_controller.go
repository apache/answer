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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity_log"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

// ActivityLogController admin community activity log
type ActivityLogController struct {
	activityLogAdminService *activity_log.ActivityLogAdminService
}

// NewActivityLogController new controller
func NewActivityLogController(activityLogAdminService *activity_log.ActivityLogAdminService) *ActivityLogController {
	return &ActivityLogController{activityLogAdminService: activityLogAdminService}
}

// GetActivityLogPage activity log page
// @Summary activity log page
// @Description activity log page (admin)
// @Tags admin
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param from query int false "from (unix seconds)"
// @Param to query int false "to (unix seconds, exclusive)"
// @Param username query string false "actor or target username"
// @Param action query string false "comma separated action keys"
// @Param q query string false "free text"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.ActivityLogItem}}
// @Router /answer/admin/api/activity-log [get]
func (c *ActivityLogController) GetActivityLogPage(ctx *gin.Context) {
	req := &schema.ActivityLogPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := c.activityLogAdminService.Page(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetActivityLogActions action keys with counts in the date range
// @Summary activity log actions
// @Tags admin
// @Produce json
// @Security ApiKeyAuth
// @Param from query int false "from (unix seconds)"
// @Param to query int false "to (unix seconds, exclusive)"
// @Success 200 {object} handler.RespBody{data=[]schema.ActivityLogActionCount}
// @Router /answer/admin/api/activity-log/actions [get]
func (c *ActivityLogController) GetActivityLogActions(ctx *gin.Context) {
	req := &schema.ActivityLogPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := c.activityLogAdminService.Actions(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetActivityLogDaily daily counts for the dashboard
// @Summary activity log daily counts
// @Tags admin
// @Produce json
// @Security ApiKeyAuth
// @Param days query int false "days back incl. today (default 14)"
// @Param tz_offset query int false "browser UTC offset in minutes"
// @Success 200 {object} handler.RespBody{data=[]schema.ActivityLogDailyRow}
// @Router /answer/admin/api/activity-log/daily [get]
func (c *ActivityLogController) GetActivityLogDaily(ctx *gin.Context) {
	req := &schema.ActivityLogDailyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := c.activityLogAdminService.Daily(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// ExportActivityLog TSV export with the same filters as the page
// @Summary activity log export (TSV)
// @Tags admin
// @Produce text/tab-separated-values
// @Security ApiKeyAuth
// @Router /answer/admin/api/activity-log/export [get]
func (c *ActivityLogController) ExportActivityLog(ctx *gin.Context) {
	req := &schema.ActivityLogPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	lang := handler.GetLangByCtx(ctx)
	labels := map[string]string{}
	for _, a := range activity_log.KnownActions {
		labels[a] = translator.Tr(lang, "ui.admin.activity_log.action."+strings.ReplaceAll(a, ".", "_"))
	}
	name := fmt.Sprintf("log-aktywnosci-%s-%s.tsv", dayOf(req.From, time.Now()), dayOf(req.To-1, time.Now()))
	ctx.Header("Content-Type", "text/tab-separated-values; charset=utf-8")
	ctx.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	ctx.Status(http.StatusOK)
	if err := c.activityLogAdminService.Export(ctx, req, labels, ctx.Writer); err != nil {
		log.Errorf("activity log export: %v", err)
	}
}

func dayOf(unix int64, fallback time.Time) string {
	if unix <= 0 {
		return fallback.Format("2006-01-02")
	}
	return time.Unix(unix, 0).Format("2006-01-02")
}
