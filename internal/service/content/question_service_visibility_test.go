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

package content

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

func TestCanViewSimilarQuestionWithSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question)); err != nil {
		t.Fatalf("create question table: %v", err)
	}

	questions := []*entity.Question{
		{ID: "1", UserID: "author", Title: "public match", OriginalText: "public", ParsedText: "public", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow},
		{ID: "2", UserID: "author", Title: "pending match", OriginalText: "pending", ParsedText: "pending", Status: entity.QuestionStatusPending, Show: entity.QuestionShow},
		{ID: "3", UserID: "author", Title: "hidden match", OriginalText: "hidden", ParsedText: "hidden", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		{ID: "4", UserID: "author", Title: "deleted match", OriginalText: "deleted", ParsedText: "deleted", Status: entity.QuestionStatusDeleted, Show: entity.QuestionShow},
		{ID: "5", UserID: "viewer", Title: "own pending match", OriginalText: "own pending", ParsedText: "own pending", Status: entity.QuestionStatusPending, Show: entity.QuestionShow},
	}
	for _, question := range questions {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question %s: %v", question.ID, err)
		}
	}

	var candidates []*entity.Question
	if err := db.Context(ctx).Where("title like ?", "%match%").Find(&candidates); err != nil {
		t.Fatalf("load similar-question candidates: %v", err)
	}
	if len(candidates) != len(questions) {
		t.Fatalf("loaded %d candidates, want %d", len(candidates), len(questions))
	}

	visibleIDs := make(map[string]bool)
	for _, question := range candidates {
		if canViewSimilarQuestion(question, "viewer", schema.QuestionPermission{}) {
			visibleIDs[question.ID] = true
		}
	}
	if len(visibleIDs) != 2 || !visibleIDs["1"] || !visibleIDs["5"] {
		t.Fatalf("visible similar questions = %v, want [1 5]", visibleIDs)
	}
}

func TestCanViewSimilarQuestion(t *testing.T) {
	testCases := []struct {
		name    string
		status  int
		show    int
		author  string
		viewer  string
		per     schema.QuestionPermission
		allowed bool
	}{
		{"public question", entity.QuestionStatusAvailable, entity.QuestionShow, "author", "viewer", schema.QuestionPermission{}, true},
		{"other user's pending question", entity.QuestionStatusPending, entity.QuestionShow, "author", "viewer", schema.QuestionPermission{}, false},
		{"author's pending question", entity.QuestionStatusPending, entity.QuestionShow, "author", "author", schema.QuestionPermission{}, true},
		{"reviewer's pending question", entity.QuestionStatusPending, entity.QuestionShow, "author", "reviewer", schema.QuestionPermission{CanReopen: true}, true},
		{"other user's hidden question", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "viewer", schema.QuestionPermission{}, false},
		{"author's hidden question", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "author", schema.QuestionPermission{}, true},
		{"moderator's hidden question", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "moderator", schema.QuestionPermission{IsAdminModerator: true}, true},
		{"deleted question", entity.QuestionStatusDeleted, entity.QuestionShow, "author", "author", schema.QuestionPermission{}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			question := &entity.Question{UserID: testCase.author, Status: testCase.status, Show: testCase.show}
			if got := canViewSimilarQuestion(question, testCase.viewer, testCase.per); got != testCase.allowed {
				t.Fatalf("canViewSimilarQuestion() = %t, want %t", got, testCase.allowed)
			}
		})
	}
}

func TestCheckQuestionVisibility(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		show   int
		userID string
		viewer string
		per    schema.QuestionPermission
		allow  bool
	}{
		{"public question", entity.QuestionStatusAvailable, entity.QuestionShow, "author", "", schema.QuestionPermission{}, true},
		{"pending question anonymous", entity.QuestionStatusPending, entity.QuestionShow, "author", "", schema.QuestionPermission{}, false},
		{"pending question author", entity.QuestionStatusPending, entity.QuestionShow, "author", "author", schema.QuestionPermission{}, true},
		{"pending question reviewer", entity.QuestionStatusPending, entity.QuestionShow, "author", "reviewer", schema.QuestionPermission{CanReopen: true}, true},
		{"deleted question anonymous", entity.QuestionStatusDeleted, entity.QuestionShow, "author", "", schema.QuestionPermission{}, false},
		{"hidden question anonymous", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "", schema.QuestionPermission{}, false},
		{"hidden question author", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "author", schema.QuestionPermission{}, true},
		{"hidden question moderator", entity.QuestionStatusAvailable, entity.QuestionHide, "author", "moderator", schema.QuestionPermission{IsAdminModerator: true}, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			question := &schema.QuestionInfoResp{
				Status: testCase.status,
				Show:   testCase.show,
				UserID: testCase.userID,
			}
			err := checkQuestionVisibility(question, testCase.viewer, testCase.per)
			if testCase.allow && err != nil {
				t.Fatalf("visibility unexpectedly denied: %v", err)
			}
			if !testCase.allow && err == nil {
				t.Fatal("visibility unexpectedly allowed")
			}
		})
	}
}
