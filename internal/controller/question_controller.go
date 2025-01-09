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
	"net/http"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/action"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/permission"
	"github.com/apache/answer/internal/service/rank"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
)

// QuestionController question controller
type QuestionController struct {
	questionService     *content.QuestionService
	answerService       *content.AnswerService
	rankService         *rank.RankService
	siteInfoService     siteinfo_common.SiteInfoCommonService
	actionService       *action.CaptchaService
	rateLimitMiddleware *middleware.RateLimitMiddleware
}

// NewQuestionController new controller
func NewQuestionController(
	questionService *content.QuestionService,
	answerService *content.AnswerService,
	rankService *rank.RankService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	actionService *action.CaptchaService,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
) *QuestionController {
	return &QuestionController{
		questionService:     questionService,
		answerService:       answerService,
		rankService:         rankService,
		siteInfoService:     siteInfoService,
		actionService:       actionService,
		rateLimitMiddleware: rateLimitMiddleware,
	}
}

// RemoveQuestion delete question
// @Summary delete question
// @Description delete question
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.RemoveQuestionReq true "question"
// @Success 200 {object} handler.RespBody
// @Router  /answer/api/v1/question [delete]
func (qc *QuestionController) RemoveQuestion(ctx *gin.Context) {
	req := &schema.RemoveQuestionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ID = uid.DeShortID(req.ID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetIsAdminFromContext(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := qc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionDelete, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	can, err := qc.rankService.CheckOperationPermission(ctx, req.UserID, permission.QuestionDelete, req.ID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	err = qc.questionService.RemoveQuestion(ctx, req)
	if !isAdmin {
		qc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionDelete, req.UserID)
	}
	handler.HandleResponse(ctx, err, nil)
}

// OperationQuestion Operation question
// @Summary Operation question
// @Description Operation question \n operation [pin unpin hide show]
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.OperationQuestionReq true "question"
// @Success 200 {object} handler.RespBody
// @Router  /answer/api/v1/question/operation [put]
func (qc *QuestionController) OperationQuestion(ctx *gin.Context) {
	req := &schema.OperationQuestionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ID = uid.DeShortID(req.ID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := qc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.QuestionPin,
		permission.QuestionUnPin,
		permission.QuestionHide,
		permission.QuestionShow,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanPin = canList[0]
	req.CanList = canList[1]
	if (req.Operation == schema.QuestionOperationPin || req.Operation == schema.QuestionOperationUnPin) && !req.CanPin {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	if (req.Operation == schema.QuestionOperationHide || req.Operation == schema.QuestionOperationShow) && !req.CanList {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	err = qc.questionService.OperationQuestion(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// CloseQuestion Close question
// @Summary Close question
// @Description Close question
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.CloseQuestionReq true "question"
// @Success 200 {object} handler.RespBody
// @Router  /answer/api/v1/question/status [put]
func (qc *QuestionController) CloseQuestion(ctx *gin.Context) {
	req := &schema.CloseQuestionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.ID = uid.DeShortID(req.ID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := qc.rankService.CheckOperationPermission(ctx, req.UserID, permission.QuestionClose, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = qc.questionService.CloseQuestion(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// ReopenQuestion reopen question
// @Summary reopen question
// @Description reopen question
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.ReopenQuestionReq true "question"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/question/reopen [put]
func (qc *QuestionController) ReopenQuestion(ctx *gin.Context) {
	req := &schema.ReopenQuestionReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.QuestionID = uid.DeShortID(req.QuestionID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := qc.rankService.CheckOperationPermission(ctx, req.UserID, permission.QuestionReopen, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = qc.questionService.ReopenQuestion(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetQuestion get question details
// @Summary get question details
// @Description get question details
// @Tags Question
// @Accept  json
// @Produce  json
// @Param id query string true "Question TagID"  default(1)
// @Success 200 {string} string ""
// @Router /answer/api/v1/question/info [get]
func (qc *QuestionController) GetQuestion(ctx *gin.Context) {
	id := ctx.Query("id")
	id = uid.DeShortID(id)
	userID := middleware.GetLoginUserIDFromContext(ctx)
	req := schema.QuestionPermission{}
	canList, err := qc.rankService.CheckOperationPermissions(ctx, userID, []string{
		permission.QuestionEdit,
		permission.QuestionDelete,
		permission.QuestionClose,
		permission.QuestionReopen,
		permission.QuestionPin,
		permission.QuestionUnPin,
		permission.QuestionHide,
		permission.QuestionShow,
		permission.AnswerInviteSomeoneToAnswer,
		permission.QuestionUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	objectOwner := qc.rankService.CheckOperationObjectOwner(ctx, userID, id)

	req.CanEdit = canList[0] || objectOwner
	req.CanDelete = canList[1]
	req.CanClose = canList[2]
	req.CanReopen = canList[3]
	req.CanPin = canList[4]
	req.CanUnPin = canList[5]
	req.CanHide = canList[6]
	req.CanShow = canList[7]
	req.CanInviteOtherToAnswer = canList[8]
	req.CanRecover = canList[9]

	info, err := qc.questionService.GetQuestionAndAddPV(ctx, id, userID, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if handler.GetEnableShortID(ctx) {
		info.ID = uid.EnShortID(info.ID)
	}
	handler.HandleResponse(ctx, nil, info)
}

// GetQuestionInviteUserInfo get question invite user info
// @Summary get question invite user info
// @Description get question invite user info
// @Tags Question
// @Accept  json
// @Produce  json
// @Param id query string true "Question ID"  default(1)
// @Success 200 {string} string ""
// @Router /answer/api/v1/question/invite [get]
func (qc *QuestionController) GetQuestionInviteUserInfo(ctx *gin.Context) {
	questionID := uid.DeShortID(ctx.Query("id"))
	resp, err := qc.questionService.InviteUserInfo(ctx, questionID)
	handler.HandleResponse(ctx, err, resp)

}

// SimilarQuestion godoc
// @Summary Search Similar Question
// @Description Search Similar Question
// @Tags Question
// @Accept  json
// @Produce  json
// @Param question_id query string true "question_id"  default()
// @Success 200 {string} string ""
// @Router /answer/api/v1/question/similar/tag [get]
func (qc *QuestionController) SimilarQuestion(ctx *gin.Context) {
	questionID := ctx.Query("question_id")
	questionID = uid.DeShortID(questionID)
	userID := middleware.GetLoginUserIDFromContext(ctx)
	list, count, err := qc.questionService.SimilarQuestion(ctx, questionID, userID)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, gin.H{
		"list":  list,
		"count": count,
	})
}

// QuestionPage get questions by page
// @Summary get questions by page
// @Description get questions by page
// @Tags Question
// @Accept  json
// @Produce  json
// @Param data body schema.QuestionPageReq  true "QuestionPageReq"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.QuestionPageResp}}
// @Router /answer/api/v1/question/page [get]
func (qc *QuestionController) QuestionPage(ctx *gin.Context) {
	req := &schema.QuestionPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	questions, total, err := qc.questionService.GetQuestionPage(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if pager.ValPageOutOfRange(total, req.Page, req.PageSize) {
		handler.HandleResponse(ctx, errors.NotFound(reason.RequestFormatError), nil)
		return
	}
	handler.HandleResponse(ctx, nil, pager.NewPageModel(total, questions))
}

// QuestionRecommendPage get recommend questions by page
// @Summary get recommend questions by page
// @Description get recommend questions by page
// @Tags Question
// @Accept  json
// @Produce  json
// @Param data body schema.QuestionPageReq  true "QuestionPageReq"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.QuestionPageResp}}
// @Router /answer/api/v1/question/recommend/page [get]
func (qc *QuestionController) QuestionRecommendPage(ctx *gin.Context) {
	req := &schema.QuestionPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	if req.LoginUserID == "" {
		handler.HandleResponse(ctx, errors.Unauthorized(reason.UnauthorizedError), nil)
		return
	}

	questions, total, err := qc.questionService.GetRecommendQuestionPage(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, pager.NewPageModel(total, questions))
}

// AddQuestion add question
// @Summary add question
// @Description add question
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.QuestionAdd true "question"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/question [post]
func (qc *QuestionController) AddQuestion(ctx *gin.Context) {
	req := &schema.QuestionAdd{}
	errFields := handler.BindAndCheckReturnErr(ctx, req)
	if ctx.IsAborted() {
		return
	}
	reject, rejectKey := qc.rateLimitMiddleware.DuplicateRequestRejection(ctx, req)
	if reject {
		return
	}
	defer func() {
		// If status is not 200 means that the bad request has been returned, so the record should be cleared
		if ctx.Writer.Status() != http.StatusOK {
			qc.rateLimitMiddleware.DuplicateRequestClear(ctx, rejectKey)
		}
	}()

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, requireRanks, err := qc.rankService.CheckOperationPermissionsForRanks(ctx, req.UserID, []string{
		permission.QuestionAdd,
		permission.QuestionEdit,
		permission.QuestionDelete,
		permission.QuestionClose,
		permission.QuestionReopen,
		permission.TagUseReservedTag,
		permission.TagAdd,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	linkUrlLimitUser := canList[7]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := qc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionQuestion, req.UserID, req.CaptchaID, req.CaptchaCode)
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
	req.CanClose = canList[3]
	req.CanReopen = canList[4]
	req.CanUseReservedTag = canList[5]
	req.CanAddTag = canList[6]
	if !req.CanAdd {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	// can add tag
	hasNewTag, err := qc.questionService.HasNewTag(ctx, req.Tags)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !req.CanAddTag && hasNewTag {
		lang := handler.GetLang(ctx)
		msg := translator.TrWithData(lang, reason.NoEnoughRankToOperate, &schema.PermissionTrTplData{Rank: requireRanks[6]})
		handler.HandleResponse(ctx, errors.Forbidden(reason.NoEnoughRankToOperate).WithMsg(msg), nil)
		return
	}

	errList, err := qc.questionService.CheckAddQuestion(ctx, req)
	if err != nil {
		errlist, ok := errList.([]*validator.FormErrorField)
		if ok {
			errFields = append(errFields, errlist...)
		}
	}

	if len(errFields) > 0 {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), errFields)
		return
	}

	req.UserAgent = ctx.GetHeader("User-Agent")
	req.IP = ctx.ClientIP()

	resp, err := qc.questionService.AddQuestion(ctx, req)
	if err != nil {
		errlist, ok := resp.([]*validator.FormErrorField)
		if ok {
			errFields = append(errFields, errlist...)
		}
	}

	if len(errFields) > 0 {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), errFields)
		return
	}
	if !isAdmin || !linkUrlLimitUser {
		qc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionQuestion, req.UserID)
	}
	handler.HandleResponse(ctx, err, resp)
}

// AddQuestionByAnswer add question
// @Summary add question and answer
// @Description add question and answer
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.QuestionAddByAnswer true "question"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/question/answer [post]
func (qc *QuestionController) AddQuestionByAnswer(ctx *gin.Context) {
	req := &schema.QuestionAddByAnswer{}
	errFields := handler.BindAndCheckReturnErr(ctx, req)
	if ctx.IsAborted() {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := qc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.QuestionAdd,
		permission.QuestionEdit,
		permission.QuestionDelete,
		permission.QuestionClose,
		permission.QuestionReopen,
		permission.TagUseReservedTag,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	linkUrlLimitUser := canList[6]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := qc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionQuestion, req.UserID, req.CaptchaID, req.CaptchaCode)
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
	req.CanClose = canList[3]
	req.CanReopen = canList[4]
	req.CanUseReservedTag = canList[5]
	if !req.CanAdd {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	questionReq := new(schema.QuestionAdd)
	err = copier.Copy(questionReq, req)
	if err != nil {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RequestFormatError), nil)
		return
	}
	errList, err := qc.questionService.CheckAddQuestion(ctx, questionReq)
	if err != nil {
		errlist, ok := errList.([]*validator.FormErrorField)
		if ok {
			errFields = append(errFields, errlist...)
		}
	}

	if len(errFields) > 0 {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), errFields)
		return
	}

	req.UserAgent = ctx.GetHeader("User-Agent")
	req.IP = ctx.ClientIP()
	resp, err := qc.questionService.AddQuestion(ctx, questionReq)
	if err != nil {
		errlist, ok := resp.([]*validator.FormErrorField)
		if ok {
			errFields = append(errFields, errlist...)
		}
	}

	if !isAdmin || !linkUrlLimitUser {
		qc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionQuestion, req.UserID)
	}

	if len(errFields) > 0 {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), errFields)
		return
	}
	//add the question id to the answer
	questionInfo, ok := resp.(*schema.QuestionInfoResp)
	if ok {
		answerReq := &schema.AnswerAddReq{}
		answerReq.QuestionID = uid.DeShortID(questionInfo.ID)
		answerReq.UserID = middleware.GetLoginUserIDFromContext(ctx)
		answerReq.Content = req.AnswerContent
		answerReq.HTML = req.AnswerHTML
		answerID, err := qc.answerService.Insert(ctx, answerReq)
		if err != nil {
			handler.HandleResponse(ctx, err, nil)
			return
		}
		info, questionInfo, has, err := qc.answerService.Get(ctx, answerID, req.UserID)
		if err != nil {
			handler.HandleResponse(ctx, err, nil)
			return
		}
		if !has {
			handler.HandleResponse(ctx, nil, nil)
			return
		}
		handler.HandleResponse(ctx, err, gin.H{
			"info":     info,
			"question": questionInfo,
		})
		return
	}

	handler.HandleResponse(ctx, err, resp)
}

// UpdateQuestion update question
// @Summary update question
// @Description update question
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.QuestionUpdate true "question"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/question [put]
func (qc *QuestionController) UpdateQuestion(ctx *gin.Context) {
	req := &schema.QuestionUpdate{}
	errFields := handler.BindAndCheckReturnErr(ctx, req)
	if ctx.IsAborted() {
		return
	}
	req.ID = uid.DeShortID(req.ID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, requireRanks, err := qc.rankService.CheckOperationPermissionsForRanks(ctx, req.UserID, []string{
		permission.QuestionEdit,
		permission.QuestionDelete,
		permission.QuestionEditWithoutReview,
		permission.TagUseReservedTag,
		permission.TagAdd,
		permission.LinkUrlLimit,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	linkUrlLimitUser := canList[5]
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin || !linkUrlLimitUser {
		captchaPass := qc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionEdit, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	objectOwner := qc.rankService.CheckOperationObjectOwner(ctx, req.UserID, req.ID)
	req.CanEdit = canList[0] || objectOwner
	req.CanDelete = canList[1]
	req.NoNeedReview = canList[2] || objectOwner
	req.CanUseReservedTag = canList[3]
	req.CanAddTag = canList[4]
	if !req.CanEdit {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	errlist, err := qc.questionService.UpdateQuestionCheckTags(ctx, req)
	if err != nil {
		errFields = append(errFields, errlist...)
	}

	if len(errFields) > 0 {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), errFields)
		return
	}

	// can add tag
	hasNewTag, err := qc.questionService.HasNewTag(ctx, req.Tags)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !req.CanAddTag && hasNewTag {
		lang := handler.GetLang(ctx)
		msg := translator.TrWithData(lang, reason.NoEnoughRankToOperate, &schema.PermissionTrTplData{Rank: requireRanks[4]})
		handler.HandleResponse(ctx, errors.Forbidden(reason.NoEnoughRankToOperate).WithMsg(msg), nil)
		return
	}

	resp, err := qc.questionService.UpdateQuestion(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, resp)
		return
	}
	respInfo, ok := resp.(*schema.QuestionInfoResp)
	if !ok {
		handler.HandleResponse(ctx, err, resp)
		return
	}
	if !isAdmin || !linkUrlLimitUser {
		qc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionEdit, req.UserID)
	}
	handler.HandleResponse(ctx, nil, &schema.UpdateQuestionResp{UrlTitle: respInfo.UrlTitle, WaitForReview: !req.NoNeedReview})
}

// QuestionRecover recover deleted question
// @Summary recover deleted question
// @Description recover deleted question
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.QuestionRecoverReq true "question"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/question/recover [post]
func (qc *QuestionController) QuestionRecover(ctx *gin.Context) {
	req := &schema.QuestionRecoverReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.QuestionID = uid.DeShortID(req.QuestionID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := qc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.QuestionUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = qc.questionService.RecoverQuestion(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateQuestionInviteUser update question invite user
// @Summary update question invite user
// @Description update question invite user
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.QuestionUpdateInviteUser true "question"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/question/invite [put]
func (qc *QuestionController) UpdateQuestionInviteUser(ctx *gin.Context) {
	req := &schema.QuestionUpdateInviteUser{}
	errFields := handler.BindAndCheckReturnErr(ctx, req)
	if ctx.IsAborted() {
		return
	}
	if len(errFields) > 0 {
		handler.HandleResponse(ctx, errors.BadRequest(reason.RequestFormatError), errFields)
		return
	}
	req.ID = uid.DeShortID(req.ID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	isAdmin := middleware.GetUserIsAdminModerator(ctx)
	if !isAdmin {
		captchaPass := qc.actionService.ActionRecordVerifyCaptcha(ctx, entity.CaptchaActionInvitationAnswer, req.UserID, req.CaptchaID, req.CaptchaCode)
		if !captchaPass {
			errFields := append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "captcha_code",
				ErrorMsg:   translator.Tr(handler.GetLang(ctx), reason.CaptchaVerificationFailed),
			})
			handler.HandleResponse(ctx, errors.BadRequest(reason.CaptchaVerificationFailed), errFields)
			return
		}
	}

	canList, err := qc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.AnswerInviteSomeoneToAnswer,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}

	req.CanInviteOtherToAnswer = canList[0]
	if !req.CanInviteOtherToAnswer {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	err = qc.questionService.UpdateQuestionInviteUser(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !isAdmin {
		qc.actionService.ActionRecordAdd(ctx, entity.CaptchaActionInvitationAnswer, req.UserID)
	}
	handler.HandleResponse(ctx, nil, nil)
}

// GetSimilarQuestions fuzzy query similar questions based on title
// @Summary fuzzy query similar questions based on title
// @Description fuzzy query similar questions based on title
// @Tags Question
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param title query string true "title"  default(string)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/question/similar [get]
func (qc *QuestionController) GetSimilarQuestions(ctx *gin.Context) {
	title := ctx.Query("title")
	resp, err := qc.questionService.GetQuestionsByTitle(ctx, title)
	handler.HandleResponse(ctx, err, resp)
}

// UserTop godoc
// @Summary UserTop
// @Description UserTop
// @Tags Question
// @Accept json
// @Produce json
// @Param username query string true "username"  default(string)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/personal/qa/top [get]
func (qc *QuestionController) UserTop(ctx *gin.Context) {
	userName := ctx.Query("username")
	userID := middleware.GetLoginUserIDFromContext(ctx)
	questionList, answerList, err := qc.questionService.SearchUserTopList(ctx, userName, userID)
	handler.HandleResponse(ctx, err, gin.H{
		"question": questionList,
		"answer":   answerList,
	})
}

// PersonalQuestionPage list personal questions
// @Summary list personal questions
// @Description list personal questions
// @Tags Personal
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param username query string true "username"  default(string)
// @Param order query string true "order"  Enums(newest,score)
// @Param page query string true "page"  default(0)
// @Param page_size query string true "page_size" default(20)
// @Success 200 {object} handler.RespBody
// @Router /personal/question/page [get]
func (qc *QuestionController) PersonalQuestionPage(ctx *gin.Context) {
	req := &schema.PersonalQuestionPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)
	resp, err := qc.questionService.PersonalQuestionPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// PersonalAnswerPage list personal answers
// @Summary list personal answers
// @Description list personal answers
// @Tags Personal
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param username query string true "username"  default(string)
// @Param order query string true "order"  Enums(newest,score)
// @Param page query string true "page"  default(0)
// @Param page_size query string true "page_size"  default(20)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/personal/answer/page [get]
func (qc *QuestionController) PersonalAnswerPage(ctx *gin.Context) {
	req := &schema.PersonalAnswerPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)
	resp, err := qc.questionService.PersonalAnswerPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// PersonalCollectionPage list personal collections
// @Summary list personal collections
// @Description list personal collections
// @Tags Collection
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query string true "page"  default(0)
// @Param page_size query string true "page_size"  default(20)
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/personal/collection/page [get]
func (qc *QuestionController) PersonalCollectionPage(ctx *gin.Context) {
	req := &schema.PersonalCollectionPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := qc.questionService.PersonalCollectionPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// AdminQuestionPage admin question page
// @Summary AdminQuestionPage admin question page
// @Description Status:[available,closed,deleted,pending]
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Param status query string false "user status" Enums(available, closed, deleted, pending)
// @Param query query string false "question id or title"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/question/page [get]
func (qc *QuestionController) AdminQuestionPage(ctx *gin.Context) {
	req := &schema.AdminQuestionPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := qc.questionService.AdminQuestionPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// AdminAnswerPage admin answer page
// @Summary AdminAnswerPage admin answer page
// @Description Status:[available,deleted,pending]
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Param status query string false "user status" Enums(available,deleted,pending)
// @Param query query string false "answer id or question title"
// @Param question_id query string false "question id"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/answer/page [get]
func (qc *QuestionController) AdminAnswerPage(ctx *gin.Context) {
	req := &schema.AdminAnswerPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)
	resp, err := qc.questionService.AdminAnswerPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// AdminUpdateQuestionStatus update question status
// @Summary update question status
// @Description update question status
// @Tags admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.AdminUpdateQuestionStatusReq true "AdminUpdateQuestionStatusReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/question/status [put]
func (qc *QuestionController) AdminUpdateQuestionStatus(ctx *gin.Context) {
	req := &schema.AdminUpdateQuestionStatusReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.QuestionID = uid.DeShortID(req.QuestionID)
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	err := qc.questionService.AdminSetQuestionStatus(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetQuestionLink get question link
// @Summary get question link
// @Description get question link
// @Tags Question
// @Param data query schema.GetQuestionLinkReq  true "GetQuestionLinkReq"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.QuestionPageResp}}
// @Router /answer/api/v1/question/link [get]
func (qc *QuestionController) GetQuestionLink(ctx *gin.Context) {
	req := &schema.GetQuestionLinkReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)
	req.QuestionID = uid.DeShortID(req.QuestionID)
	questions, total, err := qc.questionService.GetQuestionLink(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	handler.HandleResponse(ctx, nil, pager.NewPageModel(total, questions))
}
