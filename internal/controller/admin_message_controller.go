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
	"github.com/apache/answer/internal/service/admin_message"
	"github.com/gin-gonic/gin"
)

// AdminMessageController messages from admins / moderators to users (role checked in the service)
type AdminMessageController struct {
	adminMessageService *admin_message.AdminMessageService
}

// NewAdminMessageController new controller
func NewAdminMessageController(adminMessageService *admin_message.AdminMessageService) *AdminMessageController {
	return &AdminMessageController{adminMessageService: adminMessageService}
}

// Send write a message to a user
// @Summary send a message to a user (admin / moderator)
// @Tags admin-message
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.SendAdminMessageReq true "message"
// @Success 200 {object} handler.RespBody{data=schema.AdminMessageItem}
// @Router /answer/api/v1/admin-message [post]
func (c *AdminMessageController) Send(ctx *gin.Context) {
	req := &schema.SendAdminMessageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := c.adminMessageService.Send(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// Page list sent messages
// @Summary list messages (admin / moderator)
// @Tags admin-message
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.AdminMessageItem}}
// @Router /answer/api/v1/admin-message/page [get]
func (c *AdminMessageController) Page(ctx *gin.Context) {
	req := &schema.AdminMessagePageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := c.adminMessageService.Page(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
