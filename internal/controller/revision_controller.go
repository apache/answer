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
	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/internal/base/handler"
	"github.com/apache/incubator-answer/internal/base/middleware"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/content"
	"github.com/apache/incubator-answer/internal/service/permission"
	"github.com/apache/incubator-answer/internal/service/rank"
	"github.com/apache/incubator-answer/pkg/obj"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// RevisionController revision controller
type RevisionController struct {
	revisionListService *content.RevisionService
	rankService         *rank.RankService
}

// NewRevisionController new controller
func NewRevisionController(
	revisionListService *content.RevisionService,
	rankService *rank.RankService,
) *RevisionController {
	return &RevisionController{
		revisionListService: revisionListService,
		rankService:         rankService,
	}
}

// GetRevisionList godoc
// @Summary get revision list
// @Description get revision list
// @Tags Revision
// @Produce json
// @Param object_id query string true "object id"
// @Success 200 {object} handler.RespBody{data=[]schema.GetRevisionResp}
// @Router /answer/api/v1/revisions [get]
func (rc *RevisionController) GetRevisionList(ctx *gin.Context) {
	objectID := ctx.Query("object_id")
	if objectID == "0" || objectID == "" {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), nil)
		return
	}
	objectID = uid.DeShortID(objectID)
	req := &schema.GetRevisionListReq{
		ObjectID: objectID,
	}

	resp, err := rc.revisionListService.GetRevisionList(ctx, req)
	list := make([]schema.GetRevisionResp, 0)
	for _, item := range resp {
		if item.Status == entity.RevisioNnormalStatus || item.Status == entity.RevisionReviewPassStatus {
			list = append(list, item)
		}
	}
	handler.HandleResponse(ctx, err, list)
}

// GetUnreviewedRevisionList godoc
// @Summary get unreviewed revision list
// @Description get unreviewed revision list
// @Tags Revision
// @Produce json
// @Security ApiKeyAuth
// @Param page query string true "page id"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetUnreviewedRevisionResp}}
// @Router /answer/api/v1/revisions/unreviewed [get]
func (rc *RevisionController) GetUnreviewedRevisionList(ctx *gin.Context) {
	req := &schema.RevisionSearch{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := rc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
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

	resp, err := rc.revisionListService.GetUnreviewedRevisionPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// RevisionAudit godoc
// @Summary revision audit
// @Description revision audit operation:approve or reject
// @Tags Revision
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.RevisionAuditReq true "audit"
// @Success 200 {object} handler.RespBody{}
// @Router /answer/api/v1/revisions/audit [put]
func (rc *RevisionController) RevisionAudit(ctx *gin.Context) {
	req := &schema.RevisionAuditReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := rc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
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

	err = rc.revisionListService.RevisionAudit(ctx, req)
	handler.HandleResponse(ctx, err, gin.H{})
}

// CheckCanUpdateRevision check can update revision
// @Summary check can update revision
// @Description check can update revision
// @Tags Revision
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id query string true "id" default(string)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/revisions/edit/check [get]
func (rc *RevisionController) CheckCanUpdateRevision(ctx *gin.Context) {
	req := &schema.CheckCanQuestionUpdate{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	action := ""
	req.ID = uid.DeShortID(req.ID)
	objectTypeStr, _ := obj.GetObjectTypeStrByObjectID(req.ID)
	switch objectTypeStr {
	case constant.QuestionObjectType:
		action = permission.QuestionEdit
	case constant.AnswerObjectType:
		action = permission.AnswerEdit
	case constant.TagObjectType:
		action = permission.TagEdit
	default:
		handler.HandleResponse(ctx, errors.BadRequest(reason.ObjectNotFound), nil)
		return
	}

	can, err := rc.rankService.CheckOperationPermission(ctx, req.UserID, action, req.ID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	resp, err := rc.revisionListService.CheckCanUpdateRevision(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetReviewingType get reviewing type
// @Summary get reviewing type
// @Description get reviewing type
// @Tags Revision
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} handler.RespBody{data=[]schema.GetReviewingTypeResp}
// @Router /answer/api/v1/reviewing/type [get]
func (rc *RevisionController) GetReviewingType(ctx *gin.Context) {
	req := &schema.GetReviewingTypeReq{}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := rc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
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

	resp, err := rc.revisionListService.GetReviewingType(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
