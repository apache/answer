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
	"github.com/apache/answer/internal/service/ai_conversation"
	"github.com/apache/answer/internal/service/feature_toggle"
	"github.com/gin-gonic/gin"
)

// AIConversationController ai conversation controller
type AIConversationController struct {
	aiConversationService ai_conversation.AIConversationService
	featureToggleSvc      *feature_toggle.FeatureToggleService
}

// NewAIConversationController creates a new AI conversation controller
func NewAIConversationController(
	aiConversationService ai_conversation.AIConversationService,
	featureToggleSvc *feature_toggle.FeatureToggleService,
) *AIConversationController {
	return &AIConversationController{
		aiConversationService: aiConversationService,
		featureToggleSvc:      featureToggleSvc,
	}
}

func (ctrl *AIConversationController) ensureEnabled(ctx *gin.Context) bool {
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
// @Summary get conversation list
// @Description get conversation list
// @Tags ai-conversation
// @Accept json
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.AIConversationListItem}}
// @Router /answer/api/v1/ai/conversation/page [get]
func (ctrl *AIConversationController) GetConversationList(ctx *gin.Context) {
	if !ctrl.ensureEnabled(ctx) {
		return
	}
	req := &schema.AIConversationListReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := ctrl.aiConversationService.GetConversationList(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetConversationDetail gets conversation detail
// @Summary get conversation detail
// @Description get conversation detail
// @Tags ai-conversation
// @Accept json
// @Produce json
// @Param conversation_id query string true "conversation id"
// @Success 200 {object} handler.RespBody{data=schema.AIConversationDetailResp}
// @Router /answer/api/v1/ai/conversation [get]
func (ctrl *AIConversationController) GetConversationDetail(ctx *gin.Context) {
	if !ctrl.ensureEnabled(ctx) {
		return
	}
	req := &schema.AIConversationDetailReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, _, err := ctrl.aiConversationService.GetConversationDetail(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// VoteRecord vote record
// @Summary vote record
// @Description vote record
// @Tags ai-conversation
// @Accept json
// @Produce json
// @Param data body schema.AIConversationVoteReq true "vote request"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/ai/conversation/vote [post]
func (ctrl *AIConversationController) VoteRecord(ctx *gin.Context) {
	if !ctrl.ensureEnabled(ctx) {
		return
	}
	req := &schema.AIConversationVoteReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	err := ctrl.aiConversationService.VoteRecord(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
