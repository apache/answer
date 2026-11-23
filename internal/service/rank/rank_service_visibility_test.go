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

package rank_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	answerrepo "github.com/apache/answer/internal/repo/answer"
	questionrepo "github.com/apache/answer/internal/repo/question"
	rankrepo "github.com/apache/answer/internal/repo/rank"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/object_info"
	rankservice "github.com/apache/answer/internal/service/rank"
)

type rankVisibilityConfigRepo struct{}

func (*rankVisibilityConfigRepo) GetConfigByID(_ context.Context, id int) (*entity.Config, error) {
	return &entity.Config{ID: id}, nil
}

func (*rankVisibilityConfigRepo) GetConfigByKey(context.Context, string) (*entity.Config, error) {
	return nil, nil
}

func (*rankVisibilityConfigRepo) GetConfigByKeyFromDB(context.Context, string) (*entity.Config, error) {
	return nil, nil
}

func (*rankVisibilityConfigRepo) UpdateConfig(context.Context, string, string) error { return nil }

func TestRankPersonalPageFiltersRestrictedContentWithSQLite(t *testing.T) {
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
		rankUserID      = "10000000000000001"
		questionOwnerID = "10000000000000002"
		answerOwnerID   = "10000000000000003"
		publicQuestion  = "10010000000000001"
		hiddenQuestion  = "10010000000000002"
		parentQuestion  = "10010000000000003"
		deletedAnswer   = "10020000000000001"
		hiddenAnswer    = "10020000000000002"
	)
	questions := []*entity.Question{
		{ID: publicQuestion, UserID: questionOwnerID, Title: "public title", ParsedText: "public body", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow},
		{ID: hiddenQuestion, UserID: rankUserID, Title: "hidden title", ParsedText: "hidden body", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		{ID: parentQuestion, UserID: questionOwnerID, Title: "parent hidden title", ParsedText: "parent hidden body", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
	}
	for _, question := range questions {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question %s: %v", question.ID, err)
		}
	}
	answers := []*entity.Answer{
		{ID: deletedAnswer, QuestionID: publicQuestion, UserID: answerOwnerID, ParsedText: "deleted answer body", Status: entity.AnswerStatusDeleted},
		{ID: hiddenAnswer, QuestionID: parentQuestion, UserID: answerOwnerID, ParsedText: "answer under hidden question", Status: entity.AnswerStatusAvailable},
	}
	for _, answer := range answers {
		if _, err := db.Context(ctx).Insert(answer); err != nil {
			t.Fatalf("insert answer %s: %v", answer.ID, err)
		}
	}
	for index, objectID := range []string{publicQuestion, hiddenQuestion, deletedAnswer, hiddenAnswer} {
		if _, err := db.Context(ctx).Insert(&entity.Activity{
			ID:           "1000000000000001" + string(rune('0'+index)),
			UserID:       rankUserID,
			ObjectID:     objectID,
			ActivityType: 1,
			Rank:         1,
			HasRank:      1,
			Cancelled:    entity.ActivityAvailable,
		}); err != nil {
			t.Fatalf("insert activity for %s: %v", objectID, err)
		}
	}

	dbData := &data.Data{DB: db}
	questionRepo := questionrepo.NewQuestionRepo(dbData, nil)
	answerRepo := answerrepo.NewAnswerRepo(dbData, nil, nil, nil)
	objectInfoService := object_info.NewObjService(answerRepo, questionRepo, nil, nil, nil)
	service := rankservice.NewRankService(
		nil,
		rankrepo.NewUserRankRepo(dbData, nil),
		objectInfoService,
		nil,
		nil,
		config.NewConfigService(&rankVisibilityConfigRepo{}),
	)

	assertPage := func(name, loginUserID string, isAdminModerator bool, wantCount int64, wantIDs ...string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			page, err := service.GetRankPersonalPage(ctx, &schema.GetRankPersonalWithPageReq{
				Page:             1,
				PageSize:         20,
				UserID:           rankUserID,
				LoginUserID:      loginUserID,
				IsAdminModerator: isAdminModerator,
			})
			if err != nil {
				t.Fatalf("get rank page: %v", err)
			}
			if page.Count != wantCount {
				t.Fatalf("page count = %d, want %d", page.Count, wantCount)
			}
			items, ok := page.List.([]*schema.GetRankPersonalPageResp)
			if !ok {
				t.Fatalf("rank page list type = %T", page.List)
			}
			if len(items) != len(wantIDs) {
				t.Fatalf("rank page length = %d, want %d", len(items), len(wantIDs))
			}
			gotIDs := make(map[string]bool, len(items))
			for _, item := range items {
				gotIDs[item.ObjectID] = true
			}
			for _, wantID := range wantIDs {
				if !gotIDs[wantID] {
					t.Fatalf("rank page IDs = %v, missing %s", gotIDs, wantID)
				}
			}
		})
	}

	assertPage("anonymous only sees public content", "", false, 1, publicQuestion)
	assertPage("question owner sees own hidden content", rankUserID, false, 2, publicQuestion, hiddenQuestion)
	assertPage("administrator sees restricted content", "", true, 4, publicQuestion, hiddenQuestion, deletedAnswer, hiddenAnswer)
}
