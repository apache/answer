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

type PermissionController struct {
	rankService *rank.RankService
}

// NewPermissionController new language controller.
func NewPermissionController(rankService *rank.RankService) *PermissionController {
	return &PermissionController{rankService: rankService}
}

// GetPermission check user permission
// @Summary check user permission
// @Description check user permission
// @Tags Permission
// @Security ApiKeyAuth
// @Param Authorization header string true "access-token"
// @Produce json
// @Param action query string true "permission key" Enums(question.add, question.edit, question.edit_without_review, question.delete, question.close, question.reopen, question.vote_up, question.vote_down, question.pin, question.unpin, question.hide, question.show, answer.add, answer.edit, answer.edit_without_review, answer.delete, answer.accept, answer.vote_up, answer.vote_down, answer.invite_someone_to_answer, comment.add, comment.edit, comment.delete, comment.vote_up, comment.vote_down, report.add, tag.add, tag.edit, tag.edit_slug_name, tag.edit_without_review, tag.delete, tag.synonym, link.url_limit, vote.detail, answer.audit, question.audit, tag.audit, tag.use_reserved_tag)
// @Success 200 {object} handler.RespBody{data=map[string]bool}
// @Router /answer/api/v1/permission [get]
func (u *PermissionController) GetPermission(ctx *gin.Context) {
	req := &schema.GetPermissionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	userID := middleware.GetLoginUserIDFromContext(ctx)
	ops, requireRanks, err := u.rankService.CheckOperationPermissionsForRanks(ctx, userID, req.Actions)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	lang := handler.GetLangByCtx(ctx)
	mapping := make(map[string]*schema.GetPermissionResp, len(ops))
	for i, action := range req.Actions {
		t := &schema.GetPermissionResp{HasPermission: ops[i]}
		t.TrTip(lang, requireRanks[i])
		mapping[action] = t
	}
	handler.HandleResponse(ctx, err, mapping)
}
