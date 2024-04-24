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

package entity

import "time"

const (
	AnswerSearchOrderByDefault = "default"
	AnswerSearchOrderByTime    = "updated"
	AnswerSearchOrderByVote    = "vote"
	AnswerSearchOrderByTimeAsc = "created"

	AnswerStatusAvailable = 1
	AnswerStatusDeleted   = 10
	AnswerStatusPending   = 11
)

var AdminAnswerSearchStatus = map[string]int{
	"available": AnswerStatusAvailable,
	"deleted":   AnswerStatusDeleted,
	"pending":   AnswerStatusPending,
}

// Answer answer
type Answer struct {
	ID             string    `xorm:"not null pk autoincr BIGINT(20) id"`
	CreatedAt      time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt      time.Time `xorm:"updated_at TIMESTAMP"`
	QuestionID     string    `xorm:"not null default 0 BIGINT(20) question_id"`
	UserID         string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	LastEditUserID string    `xorm:"not null default 0 BIGINT(20) last_edit_user_id"`
	OriginalText   string    `xorm:"not null MEDIUMTEXT original_text"`
	ParsedText     string    `xorm:"not null MEDIUMTEXT parsed_text"`
	Status         int       `xorm:"not null default 1 INT(11) status"`
	Accepted       int       `xorm:"not null default 1 INT(11) adopted"`
	CommentCount   int       `xorm:"not null default 0 INT(11) comment_count"`
	VoteCount      int       `xorm:"not null default 0 INT(11) vote_count"`
	RevisionID     string    `xorm:"not null default 0 BIGINT(20) revision_id"`
}

type AnswerSearch struct {
	Answer
	IncludeDeleted bool   `json:"include_deleted"`
	LoginUserID    string `json:"login_user_id"`
	Order          string `json:"order_by"`                   // default or updated
	Page           int    `json:"page" form:"page"`           // Query number of pages
	PageSize       int    `json:"page_size" form:"page_size"` // Search page size
}

type PersonalAnswerPageQueryCond struct {
	Page        int
	PageSize    int
	UserID      string
	Order       string
	ShowPending bool
}

// TableName answer table name
func (Answer) TableName() string {
	return "answer"
}
