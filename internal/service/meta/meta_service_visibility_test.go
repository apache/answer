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

package meta

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

func TestReactionVisibility(t *testing.T) {
	tests := []struct {
		name             string
		objectInfo       *schema.SimpleObjectInfo
		userID           string
		isAdminModerator bool
		allowed          bool
	}{
		{
			name:       "anonymous cannot access hidden question",
			objectInfo: &schema.SimpleObjectInfo{ObjectType: constant.QuestionObjectType, QuestionCreatorUserID: "question-owner", QuestionShow: entity.QuestionHide},
		},
		{
			name:       "non-owner cannot access pending question",
			objectInfo: &schema.SimpleObjectInfo{ObjectType: constant.QuestionObjectType, QuestionCreatorUserID: "question-owner", QuestionStatus: entity.QuestionStatusPending, QuestionShow: entity.QuestionShow},
			userID:     "other-user",
		},
		{
			name:       "non-owner cannot access deleted answer",
			objectInfo: &schema.SimpleObjectInfo{ObjectType: constant.AnswerObjectType, ObjectCreatorUserID: "answer-owner", QuestionCreatorUserID: "question-owner", AnswerStatus: entity.AnswerStatusDeleted, QuestionShow: entity.QuestionShow},
			userID:     "other-user",
		},
		{
			name:       "non-owner cannot access answer on hidden question",
			objectInfo: &schema.SimpleObjectInfo{ObjectType: constant.AnswerObjectType, ObjectCreatorUserID: "answer-owner", QuestionID: "question-id", QuestionCreatorUserID: "question-owner", AnswerStatus: entity.AnswerStatusAvailable, QuestionShow: entity.QuestionHide},
			userID:     "other-user",
		},
		{
			name:       "question owner can access hidden question",
			objectInfo: &schema.SimpleObjectInfo{ObjectType: constant.QuestionObjectType, QuestionCreatorUserID: "question-owner", QuestionShow: entity.QuestionHide},
			userID:     "question-owner",
			allowed:    true,
		},
		{
			name:             "moderator can access deleted answer",
			objectInfo:       &schema.SimpleObjectInfo{ObjectType: constant.AnswerObjectType, ObjectCreatorUserID: "answer-owner", QuestionCreatorUserID: "question-owner", AnswerStatus: entity.AnswerStatusDeleted, QuestionShow: entity.QuestionShow},
			isAdminModerator: true,
			allowed:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &MetaService{objectInfoService: metaTestObjectInfoService{objectInfo: test.objectInfo}}
			err := service.checkReactionVisibility(context.Background(), "object-id", test.userID, test.isAdminModerator)
			if (err == nil) != test.allowed {
				t.Fatalf("visibility error = %v, allowed = %v", err, test.allowed)
			}
		})
	}
}

func TestReactionOperationsRejectRestrictedObject(t *testing.T) {
	service := &MetaService{objectInfoService: metaTestObjectInfoService{
		objectInfo: &schema.SimpleObjectInfo{ObjectType: constant.QuestionObjectType, QuestionCreatorUserID: "question-owner", QuestionShow: entity.QuestionHide},
	}}

	if _, err := service.GetReactionByObjectId(context.Background(), &schema.GetReactionReq{ObjectID: "question-id"}); err == nil {
		t.Fatal("anonymous read of a hidden question reaction was allowed")
	}
	if _, err := service.AddOrUpdateReaction(context.Background(), &schema.UpdateReactionReq{ObjectID: "question-id", UserID: "other-user"}); err == nil {
		t.Fatal("non-owner write to a hidden question reaction was allowed")
	}
}

type metaTestObjectInfoService struct {
	objectInfo *schema.SimpleObjectInfo
	err        error
}

func (s metaTestObjectInfoService) GetInfo(context.Context, string) (*schema.SimpleObjectInfo, error) {
	return s.objectInfo, s.err
}
