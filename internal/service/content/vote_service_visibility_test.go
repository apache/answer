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

package content_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	activityrepo "github.com/apache/answer/internal/repo/activity"
	answerrepo "github.com/apache/answer/internal/repo/answer"
	questionrepo "github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity_type"
	"github.com/apache/answer/internal/service/config"
	contentservice "github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/object_info"
)

type voteVisibilityConfigRepo struct{}

func (*voteVisibilityConfigRepo) GetConfigByID(_ context.Context, id int) (*entity.Config, error) {
	return &entity.Config{ID: id}, nil
}

func (*voteVisibilityConfigRepo) GetConfigByKey(_ context.Context, key string) (*entity.Config, error) {
	activityTypes := map[string]int{
		activity_type.QuestionVoteUp:   1,
		activity_type.QuestionVoteDown: 2,
		activity_type.AnswerVoteUp:     3,
		activity_type.AnswerVoteDown:   4,
	}
	return &entity.Config{ID: activityTypes[key], Key: key}, nil
}

func (*voteVisibilityConfigRepo) GetConfigByKeyFromDB(ctx context.Context, key string) (*entity.Config, error) {
	return (&voteVisibilityConfigRepo{}).GetConfigByKey(ctx, key)
}

func (*voteVisibilityConfigRepo) UpdateConfig(context.Context, string, string) error { return nil }

func TestListUserVotesFiltersRestrictedContentWithSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Activity), new(entity.Question), new(entity.Answer)); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	const (
		voterID         = "10000000000000001"
		ownerID         = "10000000000000002"
		publicQuestion  = "10010000000000001"
		hiddenQuestion  = "10010000000000002"
		deletedQuestion = "10010000000000003"
		parentQuestion  = "10010000000000004"
		pendingQuestion = "10010000000000005"
		hiddenAnswer    = "10020000000000001"
	)
	questions := []*entity.Question{
		{ID: publicQuestion, UserID: ownerID, Title: "public title", ParsedText: "public body", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow},
		{ID: hiddenQuestion, UserID: ownerID, Title: "hidden title", ParsedText: "hidden body", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		{ID: deletedQuestion, UserID: ownerID, Title: "deleted title", ParsedText: "deleted body", Status: entity.QuestionStatusDeleted, Show: entity.QuestionShow},
		{ID: parentQuestion, UserID: ownerID, Title: "hidden parent title", ParsedText: "hidden parent body", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		{ID: pendingQuestion, UserID: ownerID, Title: "pending title", ParsedText: "pending body", Status: entity.QuestionStatusPending, Show: entity.QuestionShow},
	}
	for _, question := range questions {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question %s: %v", question.ID, err)
		}
	}
	if _, err := db.Context(ctx).Insert(&entity.Answer{
		ID: hiddenAnswer, QuestionID: parentQuestion, UserID: ownerID, ParsedText: "hidden answer body", Status: entity.AnswerStatusAvailable,
	}); err != nil {
		t.Fatalf("insert answer: %v", err)
	}
	for index, objectID := range []string{publicQuestion, hiddenQuestion, deletedQuestion, hiddenAnswer, pendingQuestion} {
		if _, err := db.Context(ctx).Insert(&entity.Activity{
			ID:           "1000000000000001" + string(rune('0'+index)),
			UserID:       voterID,
			ObjectID:     objectID,
			ActivityType: 1,
			Cancelled:    entity.ActivityAvailable,
		}); err != nil {
			t.Fatalf("insert vote activity for %s: %v", objectID, err)
		}
	}

	dbData := &data.Data{DB: db}
	questionRepo := questionrepo.NewQuestionRepo(dbData, nil)
	answerRepo := answerrepo.NewAnswerRepo(dbData, nil, nil, nil)
	objectService := object_info.NewObjService(answerRepo, questionRepo, nil, nil, nil)
	service := contentservice.NewVoteService(
		activityrepo.NewVoteRepo(dbData, nil, nil, nil),
		config.NewConfigService(&voteVisibilityConfigRepo{}),
		nil, nil, nil, objectService, nil,
	)

	assertPage := func(name, loginUserID string, isAdminModerator bool, wantCount int64, wantIDs ...string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			page, err := service.ListUserVotes(ctx, schema.GetVoteWithPageReq{
				Page: 1, PageSize: 20, UserID: voterID, LoginUserID: loginUserID, IsAdminModerator: isAdminModerator,
			})
			if err != nil {
				t.Fatalf("list user votes: %v", err)
			}
			if page.Count != wantCount {
				t.Fatalf("page count = %d, want %d", page.Count, wantCount)
			}
			items, ok := page.List.([]*schema.GetVoteWithPageResp)
			if !ok {
				t.Fatalf("vote page list type = %T", page.List)
			}
			if len(items) != len(wantIDs) {
				t.Fatalf("vote page length = %d, want %d", len(items), len(wantIDs))
			}
			gotIDs := make(map[string]bool, len(items))
			for _, item := range items {
				gotIDs[item.ObjectID] = true
				if !isAdminModerator && loginUserID == voterID &&
					(item.Content == "hidden body" || item.Content == "deleted body" || item.Content == "hidden answer body" || item.Content == "pending body") {
					t.Fatalf("restricted content was returned: %+v", item)
				}
			}
			for _, wantID := range wantIDs {
				if !gotIDs[wantID] {
					t.Fatalf("vote page IDs = %v, missing %s", gotIDs, wantID)
				}
			}
		})
	}

	assertPage("ordinary voter only sees public content", voterID, false, 1, publicQuestion)
	assertPage("question owner sees own restricted content", ownerID, false, 5, publicQuestion, hiddenQuestion, deletedQuestion, hiddenAnswer, pendingQuestion)
	assertPage("administrator sees all content", "", true, 5, publicQuestion, hiddenQuestion, deletedQuestion, hiddenAnswer, pendingQuestion)
}
