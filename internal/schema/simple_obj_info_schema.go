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
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
)

// SimpleObjectInfo simple object info
type SimpleObjectInfo struct {
	ObjectID            string `json:"object_id"`
	ObjectCreatorUserID string `json:"object_creator_user_id"`
	QuestionID          string `json:"question_id"`
	QuestionStatus      int    `json:"question_status"`
	AnswerID            string `json:"answer_id"`
	AnswerStatus        int    `json:"answer_status"`
	CommentID           string `json:"comment_id"`
	CommentStatus       int    `json:"comment_status"`
	TagID               string `json:"tag_id"`
	ObjectType          string `json:"object_type"`
	Title               string `json:"title"`
	Content             string `json:"content"`
}

// IsDeleted is deleted
func (s *SimpleObjectInfo) IsDeleted() bool {
	switch s.ObjectType {
	case constant.QuestionObjectType:
		return s.QuestionStatus == entity.QuestionStatusDeleted
	case constant.AnswerObjectType:
		return s.AnswerStatus == entity.AnswerStatusDeleted
	case constant.CommentObjectType:
		return s.CommentStatus == entity.CommentStatusDeleted
	}
	return false
}

type UnreviewedRevisionInfoInfo struct {
	CreatedAt           int64      `json:"created_at"`
	ObjectID            string     `json:"object_id"`
	QuestionID          string     `json:"question_id"`
	AnswerID            string     `json:"answer_id"`
	CommentID           string     `json:"comment_id"`
	ObjectType          string     `json:"object_type"`
	ObjectCreatorUserID string     `json:"object_creator_user_id"`
	Title               string     `json:"title"`
	UrlTitle            string     `json:"url_title"`
	Content             string     `json:"content"`
	Html                string     `json:"html"`
	AnswerCount         int        `json:"answer_count"`
	AnswerAccepted      bool       `json:"answer_accepted"`
	Tags                []*TagResp `json:"tags"`
	Status              int        `json:"status"`
	ShowStatus          int        `json:"show_status"`
}
