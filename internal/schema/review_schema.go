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

package schema

import (
	"github.com/apache/incubator-answer/internal/base/validator"
	"github.com/apache/incubator-answer/pkg/uid"
)

// UpdateReviewReq update review request
type UpdateReviewReq struct {
	ReviewID int    `validate:"required" json:"review_id"`
	Status   string `validate:"required,oneof=approve reject" json:"status"`
	UserID   string `json:"-"`
	IsAdmin  bool   `json:"-"`
}

func (r *UpdateReviewReq) IsApprove() bool {
	return r.Status == "approve"
}

func (r *UpdateReviewReq) IsReject() bool {
	return r.Status == "reject"
}

// GetUnreviewedPostPageReq get review page request
type GetUnreviewedPostPageReq struct {
	ObjectID        string            `validate:"omitempty" form:"object_id"`
	Page            int               `validate:"omitempty" form:"page"`
	ReviewerMapping map[string]string `json:"-"`
	UserID          string            `json:"-"`
	IsAdmin         bool              `json:"-"`
}

func (r *GetUnreviewedPostPageReq) Check() (errField []*validator.FormErrorField, err error) {
	if len(r.ObjectID) > 0 {
		r.Page = 1
		r.ObjectID = uid.DeShortID(r.ObjectID)
	}
	return
}

// GetUnreviewedPostPageResp get review page response
type GetUnreviewedPostPageResp struct {
	ReviewID             int           `json:"review_id"`
	CreatedAt            int64         `json:"created_at"`
	ObjectID             string        `json:"object_id"`
	QuestionID           string        `json:"question_id"`
	AnswerID             string        `json:"answer_id"`
	CommentID            string        `json:"comment_id"`
	ObjectType           string        `json:"object_type" enums:"question,answer,comment"`
	Title                string        `json:"title"`
	UrlTitle             string        `json:"url_title"`
	OriginalText         string        `json:"original_text"`
	ParsedText           string        `json:"parsed_text"`
	Tags                 []*TagResp    `json:"tags"`
	ObjectStatus         int           `json:"object_status"`
	ObjectShowStatus     int           `json:"object_show_status"`
	AuthorUserInfo       UserBasicInfo `json:"author_user_info"`
	SubmitAt             int64         `json:"submit_at"`
	SubmitterDisplayName string        `json:"submitter_display_name"`
	Reason               string        `json:"reason"`
}
