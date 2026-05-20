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
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/ai_conversation"
	"github.com/apache/answer/internal/service/feature_toggle"
	"github.com/gin-gonic/gin"
)

// AIConversationAdminController ai conversation admin controller
type AIConversationAdminController struct {
	aiConversationService ai_conversation.AIConversationService
	featureToggleSvc      *feature_toggle.FeatureToggleService
}

// NewAIConversationAdminController new AI conversation admin controller
func NewAIConversationAdminController(
	aiConversationService ai_conversation.AIConversationService,
	featureToggleSvc *feature_toggle.FeatureToggleService,
) *AIConversationAdminController {
	return &AIConversationAdminController{
		aiConversationService: aiConversationService,
		featureToggleSvc:      featureToggleSvc,
	}
}

func (ctrl *AIConversationAdminController) ensureEnabled(ctx *gin.Context) bool {
	if ctrl.featureToggleSvc == nil {
		return true
	}
	if err := ctrl.featureToggleSvc.EnsureEnabled(ctx, feature_toggle.FeatureAIChatbot); err != nil {
		handler.HandleResponse(ctx, err, nil)
		return false
	}
	return true
}

// GetConversationList gets conversation list
// @Summary get conversation list for admin
// @Description get conversation list for admin
// @Tags ai-conversation-admin
// @Accept json
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.AIConversationAdminListItem}}
// @Router /answer/admin/api/ai/conversation/page [get]
func (ctrl *AIConversationAdminController) GetConversationList(ctx *gin.Context) {
	if !ctrl.ensureEnabled(ctx) {
		return
	}
	req := &schema.AIConversationAdminListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := ctrl.aiConversationService.GetConversationListForAdmin(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetConversationDetail get conversation detail
// @Summary get conversation detail for admin
// @Description get conversation detail for admin
// @Tags ai-conversation-admin
// @Accept json
// @Produce json
// @Param conversation_id query string true "conversation id"
// @Success 200 {object} handler.RespBody{data=schema.AIConversationAdminDetailResp}
// @Router /answer/admin/api/ai/conversation [get]
func (ctrl *AIConversationAdminController) GetConversationDetail(ctx *gin.Context) {
	if !ctrl.ensureEnabled(ctx) {
		return
	}
	req := &schema.AIConversationAdminDetailReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := ctrl.aiConversationService.GetConversationDetailForAdmin(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// DeleteConversation delete conversation
// @Summary delete conversation for admin
// @Description delete conversation and its related records for admin
// @Tags ai-conversation-admin
// @Accept json
// @Produce json
// @Param data body schema.AIConversationAdminDeleteReq true "apikey"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/ai/conversation [delete]
func (ctrl *AIConversationAdminController) DeleteConversation(ctx *gin.Context) {
	if !ctrl.ensureEnabled(ctx) {
		return
	}
	req := &schema.AIConversationAdminDeleteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	err := ctrl.aiConversationService.DeleteConversationForAdmin(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
