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
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/action"
	"github.com/apache/answer/internal/service/rank"
	"github.com/apache/answer/internal/service/review"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// ReviewController review controller
type ReviewController struct {
	reviewService *review.ReviewService
	rankService   *rank.RankService
	actionService *action.CaptchaService
}

// NewReviewController new controller
func NewReviewController(
	reviewService *review.ReviewService,
	rankService *rank.RankService,
	actionService *action.CaptchaService,
) *ReviewController {
	return &ReviewController{
		reviewService: reviewService,
		rankService:   rankService,
		actionService: actionService,
	}
}

// GetUnreviewedPostPage get unreviewed post page
// @Summary get unreviewed post page
// @Description get unreviewed post page
// @Tags Review
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param page query int false "page"
// @Param object_id query string false "object_id"
// @Success 200 {object} handler.RespBody{data=pager.PageModel{list=[]schema.GetUnreviewedPostPageResp}}
// @Router /answer/api/v1/review/pending/post/page [get]
func (rc *ReviewController) GetUnreviewedPostPage(ctx *gin.Context) {
	req := &schema.GetUnreviewedPostPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)

	req.ReviewerMapping = make(map[string]string)
	_ = plugin.CallReviewer(func(base plugin.Reviewer) error {
		info := base.Info()
		req.ReviewerMapping[info.SlugName] = info.Name.Translate(ctx)
		return nil
	})

	resp, err := rc.reviewService.GetUnreviewedPostPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateReview update review
// @Summary update review
// @Description update review
// @Tags Review
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param data body schema.UpdateReviewReq true "review"
// @Success 200 {object} handler.RespBody
// @Router /answer/api/v1/review/pending/post [put]
func (rc *ReviewController) UpdateReview(ctx *gin.Context) {
	req := &schema.UpdateReviewReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.UserID = middleware.GetLoginUserIDFromContext(ctx)
	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)
	if !req.IsAdmin {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		return
	}

	err := rc.reviewService.UpdateReview(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
