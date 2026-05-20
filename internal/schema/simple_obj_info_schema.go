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
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/errors"
)

// SimpleObjectInfo simple object info
type SimpleObjectInfo struct {
	ObjectID              string `json:"object_id"`
	ObjectCreatorUserID   string `json:"object_creator_user_id"`
	QuestionID            string `json:"question_id"`
	QuestionCreatorUserID string `json:"question_creator_user_id"`
	QuestionStatus        int    `json:"question_status"`
	QuestionShow          int    `json:"question_show"`
	AnswerID              string `json:"answer_id"`
	AnswerStatus          int    `json:"answer_status"`
	CommentID             string `json:"comment_id"`
	CommentStatus         int    `json:"comment_status"`
	TagID                 string `json:"tag_id"`
	TagStatus             int    `json:"tag_status"`
	ObjectType            string `json:"object_type"`
	Title                 string `json:"title"`
	Content               string `json:"content"`
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
	case constant.TagObjectType:
		return s.TagStatus == entity.TagStatusDeleted
	}
	return false
}

func (s *SimpleObjectInfo) CheckVisibility(userID string, isAdminModerator bool) error {
	if s == nil {
		return errors.NotFound(reason.ObjectNotFound)
	}
	if s.isObjectRestricted() && !s.canViewObject(userID, isAdminModerator) {
		return errors.NotFound(s.objectNotFoundReason())
	}
	if s.hasParentQuestion() && s.isParentQuestionRestricted() &&
		!s.canViewParentQuestion(userID, isAdminModerator) {
		return errors.NotFound(reason.QuestionNotFound)
	}
	return nil
}

func (s *SimpleObjectInfo) canViewObject(userID string, isAdminModerator bool) bool {
	if isAdminModerator {
		return true
	}
	switch s.ObjectType {
	case constant.QuestionObjectType:
		return s.QuestionCreatorUserID == userID
	case constant.AnswerObjectType, constant.CommentObjectType, constant.TagObjectType:
		return s.ObjectCreatorUserID == userID
	default:
		return false
	}
}

func (s *SimpleObjectInfo) canViewParentQuestion(userID string, isAdminModerator bool) bool {
	if isAdminModerator {
		return true
	}
	return s.QuestionCreatorUserID == userID
}

func (s *SimpleObjectInfo) hasParentQuestion() bool {
	switch s.ObjectType {
	case constant.AnswerObjectType, constant.CommentObjectType:
		return len(s.QuestionID) > 0 && s.QuestionID != "0"
	default:
		return false
	}
}

func (s *SimpleObjectInfo) isObjectRestricted() bool {
	switch s.ObjectType {
	case constant.QuestionObjectType:
		return s.QuestionStatus == entity.QuestionStatusDeleted ||
			s.QuestionStatus == entity.QuestionStatusPending ||
			s.QuestionShow == entity.QuestionHide
	case constant.AnswerObjectType:
		return s.AnswerStatus == entity.AnswerStatusDeleted || s.AnswerStatus == entity.AnswerStatusPending
	case constant.CommentObjectType:
		return s.CommentStatus == entity.CommentStatusDeleted || s.CommentStatus == entity.CommentStatusPending
	case constant.TagObjectType:
		return s.TagStatus == entity.TagStatusDeleted
	default:
		return false
	}
}

func (s *SimpleObjectInfo) isParentQuestionRestricted() bool {
	return s.QuestionStatus == entity.QuestionStatusDeleted ||
		s.QuestionStatus == entity.QuestionStatusPending ||
		s.QuestionShow == entity.QuestionHide
}

func (s *SimpleObjectInfo) objectNotFoundReason() string {
	switch s.ObjectType {
	case constant.QuestionObjectType:
		return reason.QuestionNotFound
	case constant.AnswerObjectType:
		return reason.AnswerNotFound
	case constant.CommentObjectType:
		return reason.CommentNotFound
	case constant.TagObjectType:
		return reason.TagNotFound
	default:
		return reason.ObjectNotFound
	}
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
