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
	"github.com/apache/answer/internal/service/notification"
	"github.com/apache/answer/internal/service/permission"
	"github.com/apache/answer/internal/service/rank"
	"github.com/gin-gonic/gin"
)

// NotificationController notification controller
type NotificationController struct {
	notificationService *notification.NotificationService
	rankService         *rank.RankService
}

// NewNotificationController new controller
func NewNotificationController(
	notificationService *notification.NotificationService,
	rankService *rank.RankService,
) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		rankService:         rankService,
	}
}

// GetRedDot
// @Summary GetRedDot
// @Description GetRedDot
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/status [get]
func (nc *NotificationController) GetRedDot(ctx *gin.Context) {
	req := &schema.GetRedDot{}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := nc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.QuestionAudit,
		permission.AnswerAudit,
		permission.TagAudit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanReviewQuestion = canList[0]
	req.CanReviewAnswer = canList[1]
	req.CanReviewTag = canList[2]
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)

	resp, err := nc.notificationService.GetRedDot(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// ClearRedDot
// @Summary DelRedDot
// @Description DelRedDot
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.NotificationClearRequest true "NotificationClearRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/status [put]
func (nc *NotificationController) ClearRedDot(ctx *gin.Context) {
	req := &schema.NotificationClearRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := nc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.QuestionAudit,
		permission.AnswerAudit,
		permission.TagAudit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanReviewQuestion = canList[0]
	req.CanReviewAnswer = canList[1]
	req.CanReviewTag = canList[2]

	resp, err := nc.notificationService.ClearRedDot(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// ClearUnRead
// @Summary ClearUnRead
// @Description ClearUnRead
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.NotificationClearRequest true "NotificationClearRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/read/state/all [put]
func (nc *NotificationController) ClearUnRead(ctx *gin.Context) {
	req := &schema.NotificationClearRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userID := middleware.GetLoginUserIDFromContext(ctx)
	err := nc.notificationService.ClearUnRead(ctx, userID, req.NotificationType)
	handler.HandleResponse(ctx, err, gin.H{})
}

// ClearIDUnRead
// @Summary ClearUnRead
// @Description ClearUnRead
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.NotificationClearIDRequest true "NotificationClearIDRequest"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/read/state [put]
func (nc *NotificationController) ClearIDUnRead(ctx *gin.Context) {
	req := &schema.NotificationClearIDRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	userID := middleware.GetLoginUserIDFromContext(ctx)
	err := nc.notificationService.ClearIDUnRead(ctx, userID, req.ID)
	handler.HandleResponse(ctx, err, gin.H{})
}

// GetList get notification list
// @Summary get notification list
// @Description get notification list
// @Tags Notification
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Param type query string true "type" Enums(inbox,achievement)
// @Param inbox_type query string true "inbox_type" Enums(all,posts,invites,votes)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/notification/page [get]
func (nc *NotificationController) GetList(ctx *gin.Context) {
	req := &schema.NotificationSearch{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := nc.notificationService.GetNotificationPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
