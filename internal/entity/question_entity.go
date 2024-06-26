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

import (
	"time"
)

const (
	QuestionStatusAvailable = 1
	QuestionStatusClosed    = 2
	QuestionStatusDeleted   = 10
	QuestionStatusPending   = 11
	QuestionUnPin           = 1
	QuestionPin             = 2
	QuestionShow            = 1
	QuestionHide            = 2
)

var AdminQuestionSearchStatus = map[string]int{
	"available": QuestionStatusAvailable,
	"closed":    QuestionStatusClosed,
	"deleted":   QuestionStatusDeleted,
	"pending":   QuestionStatusPending,
}

var AdminQuestionSearchStatusIntToString = map[int]string{
	QuestionStatusAvailable: "available",
	QuestionStatusClosed:    "closed",
	QuestionStatusDeleted:   "deleted",
	QuestionStatusPending:   "pending",
}

// Question question
type Question struct {
	ID               string    `xorm:"not null pk BIGINT(20) id"`
	CreatedAt        time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt        time.Time `xorm:"updated_at TIMESTAMP"`
	UserID           string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
	InviteUserID     string    `xorm:"TEXT invite_user_id"`
	LastEditUserID   string    `xorm:"not null default 0 BIGINT(20) last_edit_user_id"`
	Title            string    `xorm:"not null default '' VARCHAR(150) title"`
	OriginalText     string    `xorm:"not null MEDIUMTEXT original_text"`
	ParsedText       string    `xorm:"not null MEDIUMTEXT parsed_text"`
	Pin              int       `xorm:"not null default 1 INT(11) pin"`
	Show             int       `xorm:"not null default 1 INT(11) show"`
	Status           int       `xorm:"not null default 1 INT(11) status"`
	ViewCount        int       `xorm:"not null default 0 INT(11) view_count"`
	UniqueViewCount  int       `xorm:"not null default 0 INT(11) unique_view_count"`
	VoteCount        int       `xorm:"not null default 0 INT(11) vote_count"`
	AnswerCount      int       `xorm:"not null default 0 INT(11) answer_count"`
	CollectionCount  int       `xorm:"not null default 0 INT(11) collection_count"`
	FollowCount      int       `xorm:"not null default 0 INT(11) follow_count"`
	AcceptedAnswerID string    `xorm:"not null default 0 BIGINT(20) accepted_answer_id"`
	LastAnswerID     string    `xorm:"not null default 0 BIGINT(20) last_answer_id"`
	PostUpdateTime   time.Time `xorm:"post_update_time TIMESTAMP"`
	RevisionID       string    `xorm:"not null default 0 BIGINT(20) revision_id"`
}

// TableName question table name
func (Question) TableName() string {
	return "question"
}

// QuestionWithTagsRevision question
type QuestionWithTagsRevision struct {
	Question
	Tags []*TagSimpleInfoForRevision `json:"tags"`
}

// TagSimpleInfoForRevision tag simple info for revision
type TagSimpleInfoForRevision struct {
	ID              string `xorm:"not null pk comment('tag_id') BIGINT(20) id"`
	MainTagID       int64  `xorm:"not null default 0 BIGINT(20) main_tag_id"`
	MainTagSlugName string `xorm:"not null default '' VARCHAR(35) main_tag_slug_name"`
	SlugName        string `xorm:"not null default '' unique VARCHAR(35) slug_name"`
	DisplayName     string `xorm:"not null default '' VARCHAR(35) display_name"`
	Recommend       bool   `xorm:"not null default false BOOL recommend"`
	Reserved        bool   `xorm:"not null default false BOOL reserved"`
	RevisionID      string `xorm:"not null default 0 BIGINT(20) revision_id"`
}
