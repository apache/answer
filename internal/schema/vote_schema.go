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

type VoteReq struct {
	ObjectID         string `validate:"required" json:"object_id"`
	IsCancel         bool   `validate:"omitempty" json:"is_cancel"`
	CaptchaID        string `json:"captcha_id"`
	CaptchaCode      string `json:"captcha_code"`
	UserID           string `json:"-"`
	IsAdminModerator bool   `json:"-"`
}

type VoteResp struct {
	UpVotes    int64  `json:"up_votes"`
	DownVotes  int64  `json:"down_votes"`
	Votes      int64  `json:"votes"`
	VoteStatus string `json:"vote_status"`
}

// VoteOperationInfo vote operation info
type VoteOperationInfo struct {
	// operation object id
	ObjectID string
	// question answer comment
	ObjectType string
	// object owner user id
	ObjectCreatorUserID string
	// operation user id
	OperatingUserID string
	// vote up
	VoteUp bool
	// vote down
	VoteDown bool
	// vote activity info
	Activities []*VoteActivity
}

// VoteActivity vote activity
type VoteActivity struct {
	ActivityType   int
	ActivityUserID string
	TriggerUserID  string
	Rank           int
}

func (v *VoteActivity) HasRank() int {
	if v.Rank != 0 {
		return 1
	}
	return 0
}

type GetVoteWithPageReq struct {
	// page
	Page int `validate:"omitempty,min=1" form:"page"`
	// page size
	PageSize int `validate:"omitempty,min=1" form:"page_size"`
	// user id
	UserID string `json:"-"`
}

type GetVoteWithPageResp struct {
	// create time
	CreatedAt int64 `json:"created_at"`
	// object id
	ObjectID string `json:"object_id"`
	// question id
	QuestionID string `json:"question_id"`
	// answer id
	AnswerID string `json:"answer_id"`
	// object type
	ObjectType string `json:"object_type" enums:"question,answer,tag,comment"`
	// title
	Title string `json:"title"`
	// url title
	UrlTitle string `json:"url_title"`
	// content
	Content string `json:"content"`
	// vote type
	VoteType string `json:"vote_type"`
}
