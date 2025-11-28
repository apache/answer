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
	"strings"
	"time"

	"github.com/apache/answer/internal/base/reason"
	"github.com/segmentfault/pacman/errors"

	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/uid"
)

const (
	QuestionOperationPin   = "pin"
	QuestionOperationUnPin = "unpin"
	QuestionOperationHide  = "hide"
	QuestionOperationShow  = "show"
)

// RemoveQuestionReq delete question request
type RemoveQuestionReq struct {
	// question id
	ID          string `validate:"required" json:"id"`
	UserID      string `json:"-" ` // user_id
	IsAdmin     bool   `json:"-"`
	CaptchaID   string `json:"captcha_id"` // captcha_id
	CaptchaCode string `json:"captcha_code"`
}

type CloseQuestionReq struct {
	ID        string `validate:"required" json:"id"`
	CloseType int    `json:"close_type"` // close_type
	CloseMsg  string `json:"close_msg"`  // close_type
	UserID    string `json:"-"`          // user_id
}

type OperationQuestionReq struct {
	ID        string `validate:"required" json:"id"`
	Operation string `json:"operation"` // operation [pin unpin hide show]
	UserID    string `json:"-"`         // user_id
	CanPin    bool   `json:"-"`
	CanList   bool   `json:"-"`
}

type CloseQuestionMeta struct {
	CloseType int    `json:"close_type"`
	CloseMsg  string `json:"close_msg"`
}

// ReopenQuestionReq reopen question request
type ReopenQuestionReq struct {
	QuestionID string `json:"question_id"`
	UserID     string `json:"-"`
}

type QuestionAdd struct {
	// question title
	Title string `validate:"required,notblank,gte=6,lte=150" json:"title"`
	// content
	Content string `validate:"gte=0,lte=65535" json:"content"`
	// html
	HTML string `json:"-"`
	// tags
	Tags []*TagItem `validate:"dive" json:"tags"`
	// user id
	UserID string `json:"-"`
	QuestionPermission
	CaptchaID   string `json:"captcha_id"` // captcha_id
	CaptchaCode string `json:"captcha_code"`
	IP          string `json:"-"`
	UserAgent   string `json:"-"`
}

func (req *QuestionAdd) Check() (errFields []*validator.FormErrorField, err error) {
	req.HTML = converter.Markdown2HTML(req.Content)
	for _, tag := range req.Tags {
		if len(tag.OriginalText) > 0 {
			tag.ParsedText = converter.Markdown2HTML(tag.OriginalText)
		}
	}
	return nil, nil
}

type QuestionAddByAnswer struct {
	// question title
	Title string `validate:"required,notblank,gte=6,lte=150" json:"title"`
	// content
	Content string `validate:"gte=0,lte=65535" json:"content"`
	// html
	HTML          string `json:"-"`
	AnswerContent string `validate:"required,notblank,gte=6,lte=65535" json:"answer_content"`
	AnswerHTML    string `json:"-"`
	// tags
	Tags []*TagItem `validate:"dive" json:"tags"`
	// user id
	UserID              string   `json:"-"`
	MentionUsernameList []string `validate:"omitempty" json:"mention_username_list"`
	QuestionPermission
	CaptchaID   string `json:"captcha_id"` // captcha_id
	CaptchaCode string `json:"captcha_code"`
	IP          string `json:"-"`
	UserAgent   string `json:"-"`
}

func (req *QuestionAddByAnswer) Check() (errFields []*validator.FormErrorField, err error) {
	req.HTML = converter.Markdown2HTML(req.Content)
	req.AnswerHTML = converter.Markdown2HTML(req.AnswerContent)
	for _, tag := range req.Tags {
		if len(tag.OriginalText) > 0 {
			tag.ParsedText = converter.Markdown2HTML(tag.OriginalText)
		}
	}
	if req.AnswerHTML == "" {
		errFields = append(errFields, &validator.FormErrorField{
			ErrorField: "answer_content",
			ErrorMsg:   reason.AnswerContentCannotEmpty,
		})
		return errFields, errors.BadRequest(reason.QuestionContentCannotEmpty)
	}
	return nil, nil
}

type QuestionPermission struct {
	// whether user can add it
	CanAdd bool `json:"-"`
	// whether user can edit it
	CanEdit bool `json:"-"`
	// whether user can delete it
	CanDelete bool `json:"-"`
	// whether user can close it
	CanClose bool `json:"-"`
	// whether user can reopen it
	CanReopen bool `json:"-"`
	// whether user can pin it
	CanPin   bool `json:"-"`
	CanUnPin bool `json:"-"`
	// whether user can hide it
	CanHide bool `json:"-"`
	CanShow bool `json:"-"`
	// whether user can use reserved it
	CanUseReservedTag bool `json:"-"`
	// whether user can invite other user to answer this question
	CanInviteOtherToAnswer bool `json:"-"`
	CanAddTag              bool `json:"-"`
	CanRecover             bool `json:"-"`
}

type CheckCanQuestionUpdate struct {
	// question id
	ID string `validate:"required" form:"id"`
	// user id
	UserID  string `json:"-"`
	IsAdmin bool   `json:"-"`
}

type QuestionUpdate struct {
	// question id
	ID string `validate:"required" json:"id"`
	// question title
	Title string `validate:"required,notblank,gte=6,lte=150" json:"title"`
	// content
	Content string `validate:"gte=0,lte=65535" json:"content"`
	// html
	HTML       string   `json:"-"`
	InviteUser []string `validate:"omitempty"  json:"invite_user"`
	// tags
	Tags []*TagItem `validate:"dive" json:"tags"`
	// edit summary
	EditSummary string `validate:"omitempty" json:"edit_summary"`
	// user id
	UserID       string `json:"-"`
	NoNeedReview bool   `json:"-"`
	QuestionPermission
	CaptchaID   string `json:"captcha_id"` // captcha_id
	CaptchaCode string `json:"captcha_code"`
}

type QuestionRecoverReq struct {
	QuestionID string `validate:"required" json:"question_id"`
	UserID     string `json:"-"`
}

type QuestionUpdateInviteUser struct {
	ID         string   `validate:"required" json:"id"`
	InviteUser []string `validate:"omitempty"  json:"invite_user"`
	UserID     string   `json:"-"`
	QuestionPermission
	CaptchaID   string `json:"captcha_id"` // captcha_id
	CaptchaCode string `json:"captcha_code"`
}

func (req *QuestionUpdate) Check() (errFields []*validator.FormErrorField, err error) {
	req.HTML = converter.Markdown2HTML(req.Content)
	return nil, nil
}

type QuestionBaseInfo struct {
	ID              string `json:"id" `
	Title           string `json:"title"`
	UrlTitle        string `json:"url_title"`
	ViewCount       int    `json:"view_count"`
	AnswerCount     int    `json:"answer_count"`
	CollectionCount int    `json:"collection_count"`
	FollowCount     int    `json:"follow_count"`
	Status          string `json:"status"`
	AcceptedAnswer  bool   `json:"accepted_answer"`
}

type QuestionInfoResp struct {
	ID                   string         `json:"id" `
	Title                string         `json:"title"`
	UrlTitle             string         `json:"url_title"`
	Content              string         `json:"content"`
	HTML                 string         `json:"html"`
	Description          string         `json:"description"`
	Tags                 []*TagResp     `json:"tags"`
	ViewCount            int            `json:"view_count"`
	UniqueViewCount      int            `json:"unique_view_count"`
	VoteCount            int            `json:"vote_count"`
	AnswerCount          int            `json:"answer_count"`
	CollectionCount      int            `json:"collection_count"`
	FollowCount          int            `json:"follow_count"`
	AcceptedAnswerID     string         `json:"accepted_answer_id"`
	LastAnswerID         string         `json:"last_answer_id"`
	CreateTime           int64          `json:"create_time"`
	UpdateTime           int64          `json:"-"`
	PostUpdateTime       int64          `json:"update_time"`
	QuestionUpdateTime   int64          `json:"edit_time"`
	Pin                  int            `json:"pin"`
	Show                 int            `json:"show"`
	Status               int            `json:"status"`
	Operation            *Operation     `json:"operation,omitempty"`
	UserID               string         `json:"-"`
	LastEditUserID       string         `json:"-"`
	LastAnsweredUserID   string         `json:"-"`
	UserInfo             *UserBasicInfo `json:"user_info"`
	UpdateUserInfo       *UserBasicInfo `json:"update_user_info,omitempty"`
	LastAnsweredUserInfo *UserBasicInfo `json:"last_answered_user_info,omitempty"`
	Answered             bool           `json:"answered"`
	FirstAnswerId        string         `json:"first_answer_id"`
	Collected            bool           `json:"collected"`
	VoteStatus           string         `json:"vote_status"`
	IsFollowed           bool           `json:"is_followed"`

	// MemberActions
	MemberActions  []*PermissionMemberAction `json:"member_actions"`
	ExtendsActions []*PermissionMemberAction `json:"extends_actions"`
}

// UpdateQuestionResp update question resp
type UpdateQuestionResp struct {
	UrlTitle      string `json:"url_title"`
	WaitForReview bool   `json:"wait_for_review"`
}

type AdminQuestionInfo struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	VoteCount        int            `json:"vote_count"`
	Show             int            `json:"show"`
	Pin              int            `json:"pin"`
	AnswerCount      int            `json:"answer_count"`
	AcceptedAnswerID string         `json:"accepted_answer_id"`
	CreateTime       int64          `json:"create_time"`
	UpdateTime       int64          `json:"update_time"`
	EditTime         int64          `json:"edit_time"`
	UserID           string         `json:"-" `
	UserInfo         *UserBasicInfo `json:"user_info"`
}

type OperationLevel string

const (
	OperationLevelInfo      OperationLevel = "info"
	OperationLevelDanger    OperationLevel = "danger"
	OperationLevelWarning   OperationLevel = "warning"
	OperationLevelSecondary OperationLevel = "secondary"
)

type Operation struct {
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Msg         string         `json:"msg"`
	Time        int64          `json:"time"`
	Level       OperationLevel `json:"level"`
}

type GetCloseTypeResp struct {
	// report name
	Name string `json:"name"`
	// report description
	Description string `json:"description"`
	// report source
	Source string `json:"source"`
	// report type
	Type int `json:"type"`
	// is have content
	HaveContent bool `json:"have_content"`
	// content type
	ContentType string `json:"content_type"`
}

type UserAnswerInfo struct {
	AnswerID     string `json:"answer_id"`
	QuestionID   string `json:"question_id"`
	Accepted     int    `json:"accepted"`
	VoteCount    int    `json:"vote_count"`
	CreateTime   int    `json:"create_time"`
	UpdateTime   int    `json:"update_time"`
	QuestionInfo struct {
		Title    string `json:"title"`
		UrlTitle string `json:"url_title"`
		Tags     []any  `json:"tags"`
	} `json:"question_info"`
}

type UserQuestionInfo struct {
	ID               string `json:"question_id"`
	Title            string `json:"title"`
	UrlTitle         string `json:"url_title"`
	VoteCount        int    `json:"vote_count"`
	Tags             []any  `json:"tags"`
	ViewCount        int    `json:"view_count"`
	AnswerCount      int    `json:"answer_count"`
	CollectionCount  int    `json:"collection_count"`
	CreatedAt        int64  `json:"created_at"`
	AcceptedAnswerID string `json:"accepted_answer_id"`
	Status           string `json:"status"`
}

const (
	QuestionOrderCondNewest     = "newest"
	QuestionOrderCondActive     = "active"
	QuestionOrderCondHot        = "hot"
	QuestionOrderCondScore      = "score"
	QuestionOrderCondUnanswered = "unanswered"
	QuestionOrderCondRecommend  = "recommend"
	QuestionOrderCondFrequent   = "frequent"

	// HotInDays limit max days of the hottest question
	HotInDays = 90
)

// QuestionPageReq query questions page
type QuestionPageReq struct {
	Page      int    `validate:"omitempty,min=1" form:"page"`
	PageSize  int    `validate:"omitempty,min=1" form:"page_size"`
	OrderCond string `validate:"omitempty,oneof=newest active hot score unanswered recommend frequent" form:"order"`
	Tag       string `validate:"omitempty,gt=0,lte=100" form:"tag"`
	Username  string `validate:"omitempty,gt=0,lte=100" form:"username"`
	InDays    int    `validate:"omitempty,min=1" form:"in_days"`

	LoginUserID      string `json:"-"`
	UserIDBeSearched string `json:"-"`
	TagID            string `json:"-"`
	ShowPending      bool   `json:"-"`
}

const (
	QuestionPageRespOperationTypeAsked    = "asked"
	QuestionPageRespOperationTypeAnswered = "answered"
	QuestionPageRespOperationTypeModified = "modified"
)

type QuestionPageResp struct {
	ID          string     `json:"id" `
	CreatedAt   int64      `json:"created_at"`
	Title       string     `json:"title"`
	UrlTitle    string     `json:"url_title"`
	Description string     `json:"description"`
	Pin         int        `json:"pin"`  // 1: unpin, 2: pin
	Show        int        `json:"show"` // 0: show, 1: hide
	Status      int        `json:"status"`
	Tags        []*TagResp `json:"tags"`

	// question statistical information
	ViewCount       int `json:"view_count"`
	UniqueViewCount int `json:"unique_view_count"`
	VoteCount       int `json:"vote_count"`
	AnswerCount     int `json:"answer_count"`
	CollectionCount int `json:"collection_count"`
	FollowCount     int `json:"follow_count"`

	// answer information
	AcceptedAnswerID   string    `json:"accepted_answer_id"`
	LastAnswerID       string    `json:"last_answer_id"`
	LastAnsweredUserID string    `json:"-"`
	LastAnsweredAt     time.Time `json:"-"`

	// operator information
	OperatedAt    int64                     `json:"operated_at"`
	Operator      *QuestionPageRespOperator `json:"operator"`
	OperationType string                    `json:"operation_type"`
}

type QuestionPageRespOperator struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Rank        int    `json:"rank"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	Avatar      string `json:"avatar"`
}

type AdminQuestionPageReq struct {
	Page        int    `validate:"omitempty,min=1" form:"page"`
	PageSize    int    `validate:"omitempty,min=1" form:"page_size"`
	StatusCond  string `validate:"omitempty,oneof=normal closed deleted pending" form:"status"`
	Query       string `validate:"omitempty,gt=0,lte=100" json:"query" form:"query" `
	Status      int    `json:"-"`
	LoginUserID string `json:"-"`
}

func (req *AdminQuestionPageReq) Check() (errField []*validator.FormErrorField, err error) {
	status, ok := entity.AdminQuestionSearchStatus[req.StatusCond]
	if ok {
		req.Status = status
	}
	if req.Status == 0 {
		req.Status = 1
	}
	return nil, nil
}

// AdminAnswerPageReq admin answer page req
type AdminAnswerPageReq struct {
	Page          int    `validate:"omitempty,min=1" form:"page"`
	PageSize      int    `validate:"omitempty,min=1" form:"page_size"`
	StatusCond    string `validate:"omitempty,oneof=normal deleted pending" form:"status"`
	Query         string `validate:"omitempty,gt=0,lte=100" form:"query"`
	QuestionID    string `validate:"omitempty,gt=0,lte=24" form:"question_id"`
	QuestionTitle string `json:"-"`
	AnswerID      string `json:"-"`
	Status        int    `json:"-"`
	LoginUserID   string `json:"-"`
}

func (req *AdminAnswerPageReq) Check() (errField []*validator.FormErrorField, err error) {
	req.QuestionID = uid.DeShortID(req.QuestionID)
	if req.QuestionID == "0" {
		req.QuestionID = ""
	}

	if status, ok := entity.AdminAnswerSearchStatus[req.StatusCond]; ok {
		req.Status = status
	}
	if req.Status == 0 {
		req.Status = 1
	}

	// parse query condition
	if len(req.Query) > 0 {
		prefix := "answer:"
		if strings.Contains(req.Query, prefix) {
			req.AnswerID = uid.DeShortID(strings.TrimSpace(strings.TrimPrefix(req.Query, prefix)))
		} else {
			req.QuestionTitle = strings.TrimSpace(req.Query)
		}
	}
	return nil, nil
}

type AdminUpdateQuestionStatusReq struct {
	QuestionID string `validate:"required" json:"question_id"`
	Status     string `validate:"required,oneof=available closed deleted" json:"status"`
	UserID     string `json:"-"`
}

type PersonalQuestionPageReq struct {
	Page        int    `validate:"omitempty,min=1" form:"page"`
	PageSize    int    `validate:"omitempty,min=1" form:"page_size"`
	OrderCond   string `validate:"omitempty,oneof=newest active hot score unanswered" form:"order"`
	Username    string `validate:"omitempty,gt=0,lte=100" form:"username"`
	LoginUserID string `json:"-"`
	IsAdmin     bool   `json:"-"`
}

type PersonalAnswerPageReq struct {
	Page        int    `validate:"omitempty,min=1" form:"page"`
	PageSize    int    `validate:"omitempty,min=1" form:"page_size"`
	OrderCond   string `validate:"omitempty,oneof=newest active hot score unanswered" form:"order"`
	Username    string `validate:"omitempty,gt=0,lte=100" form:"username"`
	LoginUserID string `json:"-"`
	IsAdmin     bool   `json:"-"`
}

type PersonalCollectionPageReq struct {
	Page     int    `validate:"omitempty,min=1" form:"page"`
	PageSize int    `validate:"omitempty,min=1" form:"page_size"`
	UserID   string `json:"-"`
}

type GetQuestionLinkReq struct {
	Page       int    `validate:"omitempty,min=1" form:"page"`
	PageSize   int    `validate:"omitempty,min=1,max=100" form:"page_size"`
	QuestionID string `validate:"required" form:"question_id"`
	OrderCond  string `validate:"omitempty,oneof=newest active hot score unanswered recommend frequent" form:"order"`
	InDays     int    `validate:"omitempty,min=1" form:"in_days"`

	LoginUserID string `json:"-"`
}

type GetQuestionLinkResp struct {
	QuestionPageResp
}
