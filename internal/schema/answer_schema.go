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
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
)

// RemoveAnswerReq delete answer request
type RemoveAnswerReq struct {
	ID          string `validate:"required" json:"id"`
	UserID      string `json:"-"`
	CanDelete   bool   `json:"-"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

// RecoverAnswerReq recover answer request
type RecoverAnswerReq struct {
	AnswerID string `validate:"required" json:"answer_id"`
	UserID   string `json:"-"`
}

const (
	AnswerAcceptedFailed = 1
	AnswerAcceptedEnable = 2
)

type AnswerAddReq struct {
	QuestionID  string `json:"question_id"`
	Content     string `validate:"required,notblank,gte=6,lte=65535" json:"content"`
	HTML        string `json:"-"`
	UserID      string `json:"-"`
	CanEdit     bool   `json:"-"`
	CanDelete   bool   `json:"-"`
	CanRecover  bool   `json:"-"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
	IP          string `json:"-"`
	UserAgent   string `json:"-"`
}

func (req *AnswerAddReq) Check() (errFields []*validator.FormErrorField, err error) {
	req.HTML = converter.Markdown2HTML(req.Content)
	if req.HTML == "" {
		return append(errFields, &validator.FormErrorField{
			ErrorField: "content",
			ErrorMsg:   reason.AnswerContentCannotEmpty,
		}), errors.BadRequest(reason.AnswerContentCannotEmpty)
	}
	return nil, nil
}

type AnswerUpdateReq struct {
	ID           string `json:"id"`
	QuestionID   string `json:"question_id"`
	Title        string `json:"title"`
	Content      string `validate:"required,notblank,gte=6,lte=65535" json:"content"`
	EditSummary  string `validate:"omitempty" json:"edit_summary"`
	HTML         string `json:"-"`
	UserID       string `json:"-"`
	NoNeedReview bool   `json:"-"`
	CanEdit      bool   `json:"-"`
	CaptchaID    string `json:"captcha_id"`
	CaptchaCode  string `json:"captcha_code"`
}

func (req *AnswerUpdateReq) Check() (errFields []*validator.FormErrorField, err error) {
	req.HTML = converter.Markdown2HTML(req.Content)
	if req.HTML == "" {
		return append(errFields, &validator.FormErrorField{
			ErrorField: "content",
			ErrorMsg:   reason.AnswerContentCannotEmpty,
		}), errors.BadRequest(reason.AnswerContentCannotEmpty)
	}
	return nil, nil
}

// AnswerUpdateResp answer update resp
type AnswerUpdateResp struct {
	WaitForReview bool `json:"wait_for_review"`
}

type AnswerListReq struct {
	QuestionID string `json:"question_id" form:"question_id"`
	Order      string `json:"order" form:"order"`
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	UserID     string `json:"-"`
	IsAdmin    bool   `json:"-"`
	CanEdit    bool   `json:"-"`
	CanDelete  bool   `json:"-"`
	CanRecover bool   `json:"-"`
}

type AnswerInfo struct {
	ID             string            `json:"id"`
	QuestionID     string            `json:"question_id"`
	Content        string            `json:"content"`
	HTML           string            `json:"html"`
	CreateTime     int64             `json:"create_time"`
	UpdateTime     int64             `json:"update_time"`
	Accepted       int               `json:"accepted"`
	UserID         string            `json:"-"`
	UpdateUserID   string            `json:"-"`
	UserInfo       *UserBasicInfo    `json:"user_info,omitempty"`
	UpdateUserInfo *UserBasicInfo    `json:"update_user_info,omitempty"`
	Collected      bool              `json:"collected"`
	VoteStatus     string            `json:"vote_status"`
	VoteCount      int               `json:"vote_count"`
	QuestionInfo   *QuestionInfoResp `json:"question_info,omitempty"`
	Status         int               `json:"status"`

	// MemberActions
	MemberActions []*PermissionMemberAction `json:"member_actions"`
}

type AdminAnswerInfo struct {
	ID           string         `json:"id"`
	QuestionID   string         `json:"question_id"`
	Description  string         `json:"description"`
	CreateTime   int64          `json:"create_time"`
	UpdateTime   int64          `json:"update_time"`
	Accepted     int            `json:"accepted"`
	UserID       string         `json:"-"`
	UpdateUserID string         `json:"-"`
	UserInfo     *UserBasicInfo `json:"user_info"`
	VoteCount    int            `json:"vote_count"`
	QuestionInfo struct {
		Title string `json:"title"`
	} `json:"question_info"`
}

type AcceptAnswerReq struct {
	QuestionID string `validate:"required,gt=0,lte=30" json:"question_id"`
	AnswerID   string `validate:"omitempty" json:"answer_id"`
	UserID     string `json:"-"`
}

func (req *AcceptAnswerReq) Check() (errFields []*validator.FormErrorField, err error) {
	if len(req.AnswerID) == 0 {
		req.AnswerID = "0"
	}
	return nil, nil
}

type AdminUpdateAnswerStatusReq struct {
	AnswerID string `validate:"required" json:"answer_id"`
	Status   string `validate:"required,oneof=available deleted" json:"status"`
	UserID   string `json:"-"`
}
