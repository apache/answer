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

// AddReportReq add report request
type AddReportReq struct {
	// object id
	ObjectID string `validate:"required,gt=0,lte=20" json:"object_id"`
	// report type
	ReportType int `validate:"required" json:"report_type"`
	// report content
	Content string `validate:"omitempty,gt=0,lte=500" json:"content"`
	// user id
	UserID      string `json:"-"`
	CaptchaID   string `json:"captcha_id"` // captcha_id
	CaptchaCode string `json:"captcha_code"`
}

// GetReportListReq get report list all request
type GetReportListReq struct {
	// report source
	Source string `validate:"required,oneof=question answer comment" form:"source"`
}

// GetReportTypeResp get report response
type GetReportTypeResp struct {
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

// ReportHandleReq request handle request
type ReportHandleReq struct {
	ID             string `validate:"required" comment:"report id" form:"id" json:"id"`
	FlaggedType    int    `validate:"required" comment:"flagged type" form:"flagged_type" json:"flagged_type"`
	FlaggedContent string `validate:"omitempty" comment:"flagged content" form:"flagged_content" json:"flagged_content"`
}

// GetReportListPageDTO report list data transfer object
type GetReportListPageDTO struct {
	Page     int
	PageSize int
	Status   int
}

// GetReportListPageResp get report list
type GetReportListPageResp struct {
	FlagID           string        `json:"flag_id"`
	CreatedAt        int64         `json:"created_at"`
	ObjectID         string        `json:"object_id"`
	QuestionID       string        `json:"question_id"`
	AnswerID         string        `json:"answer_id"`
	CommentID        string        `json:"comment_id"`
	ObjectType       string        `json:"object_type" enums:"question,answer,comment"`
	Title            string        `json:"title"`
	UrlTitle         string        `json:"url_title"`
	OriginalText     string        `json:"original_text"`
	ParsedText       string        `json:"parsed_text"`
	AnswerCount      int           `json:"answer_count"`
	AnswerAccepted   bool          `json:"answer_accepted"`
	Tags             []*TagResp    `json:"tags"`
	ObjectStatus     int           `json:"object_status"`
	ObjectShowStatus int           `json:"object_show_status"`
	AuthorUserInfo   UserBasicInfo `json:"author_user_info"`
	SubmitAt         int64         `json:"submit_at"`
	SubmitterUser    UserBasicInfo `json:"submitter_user"`
	Reason           *ReasonItem   `json:"reason"`
	ReasonContent    string        `json:"reason_content"`
}

// GetUnreviewedReportPostPageReq get unreviewed report post page request
type GetUnreviewedReportPostPageReq struct {
	Page    int    `json:"page" form:"page"`
	UserID  string `json:"-"`
	IsAdmin bool   `json:"-"`
}

// ReviewReportReq review report request
type ReviewReportReq struct {
	FlagID        string     `validate:"required" json:"flag_id"`
	OperationType string     `validate:"required,oneof=edit_post close_post delete_post unlist_post ignore_report" json:"operation_type"`
	CloseType     int        `validate:"omitempty" json:"close_type"`
	CloseMsg      string     `validate:"omitempty" json:"close_msg"`
	Title         string     `validate:"omitempty,notblank,gte=6,lte=150" json:"title"`
	Content       string     `validate:"omitempty,notblank,gte=6,lte=65535" json:"content"`
	Tags          []*TagItem `validate:"omitempty,dive" json:"tags"`
	UserID        string     `json:"-"`
	IsAdmin       bool       `json:"-"`
}
