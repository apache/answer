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
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/permission"
	"github.com/apache/answer/internal/service/rank"
	"github.com/apache/answer/internal/service/tag"
	"github.com/apache/answer/internal/service/tag_common"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// TagController tag controller
type TagController struct {
	tagService       *tag.TagService
	tagCommonService *tag_common.TagCommonService
	rankService      *rank.RankService
}

// NewTagController new controller
func NewTagController(
	tagService *tag.TagService,
	tagCommonService *tag_common.TagCommonService,
	rankService *rank.RankService,
) *TagController {
	return &TagController{tagService: tagService, tagCommonService: tagCommonService, rankService: rankService}
}

// SearchTagLike get tag list
// @Summary get tag list
// @Description get tag list
// @Tags Tag
// @Produce json
// @Security ApiKeyAuth
// @Param tag query string false "tag"
// @Success 200 {object} handler.RespBody{data=[]schema.GetTagBasicResp}
// @Router /answer/api/v1/question/tags [get]
func (tc *TagController) SearchTagLike(ctx *gin.Context) {
	req := &schema.SearchTagLikeReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := tc.tagCommonService.SearchTagLike(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetTagsBySlugName get tags list
// @Summary get tags list
// @Description get tags list by slug name
// @Tags Tag
// @Produce json
// @Param tags query []string false "string collection" collectionFormat(csv)
// @Success 200 {object} handler.RespBody{data=[]schema.GetTagBasicResp}
// @Router /answer/api/v1/tags [get]
func (tc *TagController) GetTagsBySlugName(ctx *gin.Context) {
	req := &schema.SearchTagsBySlugName{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := tc.tagService.GetTagsBySlugName(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// RemoveTag delete tag
// @Summary delete tag
// @Description delete tag
// @Security ApiKeyAuth
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.RemoveTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag [delete]
func (tc *TagController) RemoveTag(ctx *gin.Context) {
	req := &schema.RemoveTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := tc.rankService.CheckOperationPermission(ctx, req.UserID, permission.TagDelete, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	err = tc.tagService.RemoveTag(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AddTag add tag
// @Summary add tag
// @Description add tag
// @Security ApiKeyAuth
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.AddTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag [post]
func (tc *TagController) AddTag(ctx *gin.Context) {
	req := &schema.AddTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := tc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagAdd,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	resp, err := tc.tagCommonService.AddTag(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateTag update tag
// @Summary update tag
// @Description update tag
// @Security ApiKeyAuth
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.UpdateTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag [put]
func (tc *TagController) UpdateTag(ctx *gin.Context) {
	req := &schema.UpdateTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := tc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagEdit,
		permission.TagEditWithoutReview,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}
	req.NoNeedReview = canList[1]

	err = tc.tagService.UpdateTag(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
	} else {
		handler.HandleResponse(ctx, err, &schema.UpdateTagResp{WaitForReview: !req.NoNeedReview})
	}
}

// RecoverTag recover delete tag
// @Summary recover delete tag
// @Description recover delete tag
// @Security ApiKeyAuth
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.RecoverTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag/recover [post]
func (tc *TagController) RecoverTag(ctx *gin.Context) {
	req := &schema.RecoverTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	canList, err := tc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !canList[0] {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = tc.tagService.RecoverTag(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// GetTagInfo get tag one
// @Summary get tag one
// @Description get tag one
// @Tags Tag
// @Accept json
// @Produce json
// @Param tag_id query string true "tag id"
// @Param tag_name query string true "tag name"
// @Success 200 {object} handler.RespBody{data=schema.GetTagResp}
// @Router /answer/api/v1/tag [get]
func (tc *TagController) GetTagInfo(ctx *gin.Context) {
	req := &schema.GetTagInfoReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	canList, err := tc.rankService.CheckOperationPermissions(ctx, req.UserID, []string{
		permission.TagEdit,
		permission.TagDelete,
		permission.TagUnDelete,
	})
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = canList[0]
	req.CanDelete = canList[1]
	req.CanRecover = canList[2]
	req.CanMerge = middleware.GetUserIsAdminModerator(ctx)

	resp, err := tc.tagService.GetTagInfo(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetTagWithPage get tag page
// @Summary get tag page
// @Description get tag page
// @Tags Tag
// @Produce json
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Param slug_name query string false "slug_name"
// @Param query_cond query string false "query condition" Enums(popular, name, newest)
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetTagPageResp}}
// @Router /answer/api/v1/tags/page [get]
func (tc *TagController) GetTagWithPage(ctx *gin.Context) {
	req := &schema.GetTagWithPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	resp, err := tc.tagService.GetTagWithPage(ctx, req)
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if pager.ValPageOutOfRange(resp.Count, req.Page, req.PageSize) {
		handler.HandleResponse(ctx, errors.NotFound(reason.RequestFormatError), nil)
		return
	}
	handler.HandleResponse(ctx, err, resp)
}

// GetFollowingTags get following tag list
// @Summary get following tag list
// @Description get following tag list
// @Security ApiKeyAuth
// @Tags Tag
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]schema.GetFollowingTagsResp}
// @Router /answer/api/v1/tags/following [get]
func (tc *TagController) GetFollowingTags(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	resp, err := tc.tagService.GetFollowingTags(ctx, userID)
	handler.HandleResponse(ctx, err, resp)
}

// GetTagSynonyms get tag synonyms
// @Summary get tag synonyms
// @Description get tag synonyms
// @Tags Tag
// @Produce json
// @Param tag_id query int true "tag id"
// @Success 200 {object} handler.RespBody{data=schema.GetTagSynonymsResp}
// @Router /answer/api/v1/tag/synonyms [get]
func (tc *TagController) GetTagSynonyms(ctx *gin.Context) {
	req := &schema.GetTagSynonymsReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := tc.rankService.CheckOperationPermission(ctx, req.UserID, permission.TagSynonym, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	req.CanEdit = can

	resp, err := tc.tagService.GetTagSynonyms(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateTagSynonym update tag
// @Summary update tag
// @Description update tag
// @Security ApiKeyAuth
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.UpdateTagSynonymReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag/synonym [put]
func (tc *TagController) UpdateTagSynonym(ctx *gin.Context) {
	req := &schema.UpdateTagSynonymReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	can, err := tc.rankService.CheckOperationPermission(ctx, req.UserID, permission.TagSynonym, "")
	if err != nil {
		handler.HandleResponse(ctx, err, nil)
		return
	}
	if !can {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	err = tc.tagService.UpdateTagSynonym(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// MergeTag merge tag
// @Summary merge tag
// @Description merge tag
// @Security ApiKeyAuth
// @Tags Tag
// @Accept json
// @Produce json
// @Param data body schema.AddTagReq true "tag"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/tag/merge [post]
func (tc *TagController) MergeTag(ctx *gin.Context) {
	req := &schema.MergeTagReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	isAdminModerator := middleware.GetUserIsAdminModerator(ctx)
	if !isAdminModerator {
		handler.HandleResponse(ctx, errors.Forbidden(reason.RankFailToMeetTheCondition), nil)
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	err := tc.tagService.MergeTag(ctx, req)

	handler.HandleResponse(ctx, err, nil)
}
