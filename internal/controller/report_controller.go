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
	"github.com/apache/incubator-answer/internal/base/handler"
	"github.com/apache/incubator-answer/internal/base/middleware"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/base/translator"
	"github.com/apache/incubator-answer/internal/base/validator"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/action"
	"github.com/apache/incubator-answer/internal/service/permission"
	"github.com/apache/incubator-answer/internal/service/rank"
	"github.com/apache/incubator-answer/internal/service/report"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// ReportController report controller
type ReportController struct {
	reportService *report.ReportService
	rankService   *rank.RankService
	actionService *action.CaptchaService
}

// NewReportController new controller
func NewReportController(
	reportService *report.ReportService,
	rankService *rank.RankService,
	actionService *action.CaptchaService,
) *ReportController {
	return &ReportController{
		reportService: reportService,
		rankService:   rankService,
		actionService: actionService,
	}
}

// AddReport add report
// @Summary add report
// @Description add report <br> source (question, answer, comment, user)
// @Security ApiKeyAuth
// @Tags Report
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AddReportReq true "report"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/report [post]
func (rc *ReportController) AddReport(ctx *gin.Context) {
	req := &schema.AddReportReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := rc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionReport, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	can, err := rc.rankService.CheckOperationPermission(ctx, req.UserID, permission.ReportAdd, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = rc.reportService.AddReport(ctx, req)
	if !isAdmin {
		rc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionReport, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// GetUnreviewedReportPostPage get unreviewed report post page
// @Summary get unreviewed report post page
// @Description get unreviewed report post page
// @Tags Report
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetReportListPageResp}}
// @Router /answer/api/v1/report/unreviewed/post [get]
func (rc *ReportController) GetUnreviewedReportPostPage(ctx *gin.Context) {
	req := &schema.GetUnreviewedReportPostPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)

	resp, err := rc.reportService.GetUnreviewedReportPostPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// ReviewReport review report
// @Summary review report
// @Description review report
// @Security ApiKeyAuth
// @Tags Report
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.ReviewReportReq true "flag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/report/review [put]
func (rc *ReportController) ReviewReport(ctx *gin.Context) {
	req := &schema.ReviewReportReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)
	if !req.IsAdmin {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		return
	}

	err := rc.reportService.ReviewReport(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
