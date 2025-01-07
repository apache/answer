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
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/action"
	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/permission"
	"github.com/apache/answer/internal/service/rank"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"net/http"
)

// CommentController comment controller
type CommentController struct {
	commentService      *comment.CommentService
	rankService         *rank.RankService
	actionService       *action.CaptchaService
	rateLimitMiddleware *middleware.RateLimitMiddleware
}

// NewCommentController new controller
func NewCommentController(
	commentService *comment.CommentService,
	rankService *rank.RankService,
	actionService *action.CaptchaService,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
) *CommentController {
	return &CommentController{
		commentService:      commentService,
		rankService:         rankService,
		actionService:       actionService,
		rateLimitMiddleware: rateLimitMiddleware,
	}
}

// AddComment add comment
// @Summary add comment
// @Description add comment
// @Tags Comment
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AddCommentReq true "comment"
// @Success 200 {object} handler.RespBody{data=schema.GetCommentResp}
// @Router /answer/api/v1/comment [post]
func (cc *CommentController) AddComment(ctx *gin.Context) {
	req := &schema.AddCommentReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	reject, rejectKey := cc.rateLimitMiddleware.DuplicateRequestRejection(ctx, req)
	if reject {
		return
	}
	defer func() {
		// If status is not 200 means that the bad request has been returned, so the record should be cleared
		if ctx.Writer.Status() != http.StatusOK {
			cc.rateLimitMiddleware.DuplicateRequestClear(ctx, rejectKey)
		}
	}()
	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := cc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.CommentAdd,
		permission.CommentEdit,
		permission.CommentDelete,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	linkUrlLimitUser := canList[3]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := cc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionComment, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	req.CanAdd = canList[0]
	req.CanEdit = canList[1]
	req.CanDelete = canList[2]
	if !req.CanAdd {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	resp, err := cc.commentService.AddComment(ctx, req)
	if !isAdmin || !linkUrlLimitUser {
		cc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionComment, req.UserID)
	}
	handler.HandleResponse(ctx, err, resp)
}

// RemoveComment remove comment
// @Summary remove comment
// @Description remove comment
// @Tags Comment
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.RemoveCommentReq true "comment"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/comment [delete]
func (cc *CommentController) RemoveComment(ctx *gin.Context) {
	req := &schema.RemoveCommentReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := cc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionDelete, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}
	can, err := cc.rankService.CheckOperationPermission(ctx, req.UserID, permission.CommentDelete, req.CommentID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = cc.commentService.RemoveComment(ctx, req)
	if !isAdmin {
		cc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionDelete, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// UpdateComment update comment
// @Summary update comment
// @Description update comment
// @Tags Comment
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateCommentReq true "comment"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/comment [put]
func (cc *CommentController) UpdateComment(ctx *gin.Context) {
	req := &schema.UpdateCommentReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetIsAdminFromContext(ctx)
	canList, err := cc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.CommentEdit,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = canList[0] || cc.rankService.CheckOperationObjectOwner(ctx, req.UserID, req.CommentID)
	linkUrlLimitUser := canList[1]
	if !req.CanEdit {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	if !req.IsAdmin || !linkUrlLimitUser {
		captchaPass := cc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEdit, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	resp, err := cc.commentService.UpdateComment(ctx, req)
	if !req.IsAdmin || !linkUrlLimitUser {
		cc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionEdit, req.UserID)
	}
	handler.HandleResponse(ctx, err, resp)
}

// GetCommentWithPage get comment page
// @Summary get comment page
// @Description get comment page
// @Tags Comment
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param object_id query string true "object id"
// @Param query_cond query string false "query condition" Enums(vote)
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetCommentResp}}
// @Router /answer/api/v1/comment/page [get]
func (cc *CommentController) GetCommentWithPage(ctx *gin.Context) {
	req := &schema.GetCommentWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ObjectID = uid.DeShortID(req.ObjectID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := cc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.CommentEdit,
		permission.CommentDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = canList[0]
	req.CanDelete = canList[1]

	resp, err := cc.commentService.GetCommentWithPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetCommentPersonalWithPage user personal comment list
// @Summary user personal comment list
// @Description user personal comment list
// @Tags Comment
// @Produce json
// @Param page query int false "page"
// @Param page_size query int false "page size"
// @Param username query string false "username"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetCommentPersonalWithPageResp}}
// @Router /answer/api/v1/personal/comment/page [get]
func (cc *CommentController) GetCommentPersonalWithPage(ctx *gin.Context) {
	req := &schema.GetCommentPersonalWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := cc.commentService.GetCommentPersonalWithPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetComment godoc
// @Summary get comment by id
// @Description get comment by id
// @Tags Comment
// @Produce json
// @Param id query string true "id"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetCommentResp}}
// @Router /answer/api/v1/comment [get]
func (cc *CommentController) GetComment(ctx *gin.Context) {
	req := &schema.GetCommentReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := cc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.CommentEdit,
		permission.CommentDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = canList[0]
	req.CanDelete = canList[1]

	resp, err := cc.commentService.GetComment(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}
