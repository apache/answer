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

package follow

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

type followVisibilityObjectInfoService struct{ info *schema.SimpleObjectInfo }

func (s followVisibilityObjectInfoService) GetInfo(context.Context, string) (*schema.SimpleObjectInfo, error) {
	return s.info, nil
}

type followVisibilityRepo struct{ followCalled, cancelCalled bool }

func (r *followVisibilityRepo) Follow(context.Context, string, string) error {
	r.followCalled = true
	return nil
}
func (r *followVisibilityRepo) FollowCancel(context.Context, string, string) error {
	r.cancelCalled = true
	return nil
}

type followVisibilityCommonRepo struct{}

func (followVisibilityCommonRepo) GetFollowIDs(context.Context, string, string) ([]string, error) {
	return nil, nil
}
func (followVisibilityCommonRepo) GetFollowAmount(context.Context, string) (int, error) {
	return 0, nil
}
func (followVisibilityCommonRepo) GetFollowUserIDs(context.Context, string) ([]string, error) {
	return nil, nil
}
func (followVisibilityCommonRepo) IsFollowed(context.Context, string, string) (bool, error) {
	return false, nil
}
func (followVisibilityCommonRepo) MigrateFollowers(context.Context, string, string, string) error {
	return nil
}

func TestFollowRejectsRestrictedQuestionForOtherUser(t *testing.T) {
	repo := &followVisibilityRepo{}
	service := &FollowService{
		followRepo: repo,
		objectInfoService: followVisibilityObjectInfoService{info: &schema.SimpleObjectInfo{
			ObjectType:            "question",
			QuestionCreatorUserID: "author",
			QuestionStatus:        entity.QuestionStatusAvailable,
			QuestionShow:          entity.QuestionHide,
		}},
	}

	_, err := service.Follow(context.Background(), &schema.FollowDTO{
		ObjectID: "10010000000000001", UserID: "other-user",
	})
	if err == nil {
		t.Fatal("Follow() allowed another user to follow a hidden question")
	}
	if repo.followCalled {
		t.Fatal("Follow() called the repository after visibility was denied")
	}
}

func TestFollowAllowsQuestionAuthorToFollowRestrictedQuestion(t *testing.T) {
	repo := &followVisibilityRepo{}
	service := &FollowService{
		followRepo: repo,
		objectInfoService: followVisibilityObjectInfoService{info: &schema.SimpleObjectInfo{
			ObjectType:            "question",
			QuestionCreatorUserID: "author",
			QuestionStatus:        entity.QuestionStatusPending,
			QuestionShow:          entity.QuestionShow,
		}},
	}

	if err := service.checkVisibility(context.Background(), &schema.FollowDTO{
		ObjectID: "10010000000000001", UserID: "author",
	}); err != nil {
		t.Fatalf("question author was denied: %v", err)
	}
}

func TestFollowCancelAllowsRemovalOfHistoricalRestrictedFollow(t *testing.T) {
	repo := &followVisibilityRepo{}
	service := &FollowService{
		followRepo:       repo,
		followCommonRepo: followVisibilityCommonRepo{},
	}

	_, err := service.Follow(context.Background(), &schema.FollowDTO{
		ObjectID: "10010000000000001", UserID: "ordinary-user", IsCancel: true,
	})
	if err != nil {
		t.Fatalf("Follow() cancel error: %v", err)
	}
	if !repo.cancelCalled {
		t.Fatal("Follow() did not cancel the historical follow relationship")
	}
}
