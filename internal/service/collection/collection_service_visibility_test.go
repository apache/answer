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

package collection

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/errors"
)

type collectionTestObjectInfoService struct {
	objectInfo *schema.SimpleObjectInfo
	err        error
}

func (s collectionTestObjectInfoService) GetInfo(context.Context, string) (*schema.SimpleObjectInfo, error) {
	return s.objectInfo, s.err
}

func TestCollectionSwitchRejectsHiddenQuestionForOtherUsers(t *testing.T) {
	service := &CollectionService{objectInfoService: collectionTestObjectInfoService{
		objectInfo: &schema.SimpleObjectInfo{
			ObjectType:            constant.QuestionObjectType,
			QuestionCreatorUserID: "question-owner",
			QuestionStatus:        entity.QuestionStatusAvailable,
			QuestionShow:          entity.QuestionHide,
		},
	}}

	_, err := service.CollectionSwitch(context.Background(), &schema.CollectionSwitchReq{
		ObjectID: "question-id",
		UserID:   "other-user",
		Bookmark: true,
	})
	if err == nil {
		t.Fatal("CollectionSwitch() allowed another user to bookmark a hidden question")
	}
}

func TestCollectionSwitchReturnsNotFoundForMissingQuestion(t *testing.T) {
	service := &CollectionService{objectInfoService: collectionTestObjectInfoService{
		err: errors.BadRequest(reason.ObjectNotFound),
	}}
	_, err := service.CollectionSwitch(context.Background(), &schema.CollectionSwitchReq{ObjectID: "question-id", UserID: "other-user"})
	if err == nil {
		t.Fatal("CollectionSwitch() allowed a missing question")
	}
	serviceErr, ok := err.(*errors.Error)
	if !ok || serviceErr.Code != 404 || serviceErr.Reason != reason.QuestionNotFound {
		t.Fatalf("CollectionSwitch() error = %#v, want QuestionNotFound", err)
	}
}
