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

package comment

import (
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

func TestCanViewPersonalComment(t *testing.T) {
	testCases := []struct {
		name     string
		status   int
		show     int
		viewer   string
		isAdmin  bool
		expected bool
	}{
		{"anonymous can view public question", entity.QuestionStatusAvailable, entity.QuestionShow, "", false, true},
		{"anonymous cannot view hidden question", entity.QuestionStatusAvailable, entity.QuestionHide, "", false, false},
		{"anonymous cannot view pending question", entity.QuestionStatusPending, entity.QuestionShow, "", false, false},
		{"anonymous cannot view deleted question", entity.QuestionStatusDeleted, entity.QuestionShow, "", false, false},
		{"question owner can view hidden question", entity.QuestionStatusAvailable, entity.QuestionHide, "question-owner", false, true},
		{"moderator can view deleted question", entity.QuestionStatusDeleted, entity.QuestionShow, "", true, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			objInfo := &schema.SimpleObjectInfo{
				ObjectType:            constant.QuestionObjectType,
				QuestionCreatorUserID: "question-owner",
				QuestionStatus:        testCase.status,
				QuestionShow:          testCase.show,
			}
			if got := canViewPersonalComment(objInfo, testCase.viewer, testCase.isAdmin); got != testCase.expected {
				t.Fatalf("canViewPersonalComment() = %v, want %v", got, testCase.expected)
			}
		})
	}
}

func TestCanViewPersonalCommentOnAnswer(t *testing.T) {
	testCases := []struct {
		name     string
		viewer   string
		isAdmin  bool
		info     *schema.SimpleObjectInfo
		expected bool
	}{
		{
			name:     "anonymous cannot view answer on hidden question",
			info:     &schema.SimpleObjectInfo{ObjectType: constant.AnswerObjectType, ObjectCreatorUserID: "answer-owner", QuestionID: "question-id", QuestionCreatorUserID: "question-owner", AnswerStatus: entity.AnswerStatusAvailable, QuestionStatus: entity.QuestionStatusAvailable, QuestionShow: entity.QuestionHide},
			expected: false,
		},
		{
			name:     "question owner can view answer on hidden question",
			viewer:   "question-owner",
			info:     &schema.SimpleObjectInfo{ObjectType: constant.AnswerObjectType, ObjectCreatorUserID: "answer-owner", QuestionID: "question-id", QuestionCreatorUserID: "question-owner", AnswerStatus: entity.AnswerStatusAvailable, QuestionStatus: entity.QuestionStatusAvailable, QuestionShow: entity.QuestionHide},
			expected: true,
		},
		{
			name:     "anonymous cannot view pending answer",
			info:     &schema.SimpleObjectInfo{ObjectType: constant.AnswerObjectType, ObjectCreatorUserID: "answer-owner", QuestionID: "question-id", QuestionCreatorUserID: "question-owner", AnswerStatus: entity.AnswerStatusPending, QuestionStatus: entity.QuestionStatusAvailable, QuestionShow: entity.QuestionShow},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := canViewPersonalComment(testCase.info, testCase.viewer, testCase.isAdmin); got != testCase.expected {
				t.Fatalf("canViewPersonalComment() = %v, want %v", got, testCase.expected)
			}
		})
	}
}
