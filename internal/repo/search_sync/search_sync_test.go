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

package search_sync

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/plugin"
)

type cleanupTestSearch struct {
	deleted map[string]bool
}

func (*cleanupTestSearch) Info() plugin.Info { return plugin.Info{} }
func (*cleanupTestSearch) Description() plugin.SearchDesc {
	return plugin.SearchDesc{}
}
func (*cleanupTestSearch) RegisterSyncer(context.Context, plugin.SearchSyncer) {}
func (*cleanupTestSearch) SearchContents(context.Context, *plugin.SearchBasicCond) ([]plugin.SearchResult, int64, error) {
	return nil, 0, nil
}
func (*cleanupTestSearch) SearchQuestions(context.Context, *plugin.SearchBasicCond) ([]plugin.SearchResult, int64, error) {
	return nil, 0, nil
}
func (*cleanupTestSearch) SearchAnswers(context.Context, *plugin.SearchBasicCond) ([]plugin.SearchResult, int64, error) {
	return nil, 0, nil
}
func (*cleanupTestSearch) UpdateContent(context.Context, *plugin.SearchContent) error { return nil }
func (s *cleanupTestSearch) DeleteContent(_ context.Context, objectID string) error {
	s.deleted[objectID] = true
	return nil
}

func TestPluginSyncerIndexesOnlyPublicContent(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question), new(entity.Answer), new(entity.TagRel)); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	publicQuestion := &entity.Question{ID: "10010000000000001", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow}
	hiddenQuestion := &entity.Question{ID: "10010000000000002", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide}
	pendingQuestion := &entity.Question{ID: "10010000000000003", Status: entity.QuestionStatusPending, Show: entity.QuestionShow}
	for _, question := range []*entity.Question{publicQuestion, hiddenQuestion, pendingQuestion} {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question: %v", err)
		}
	}
	publicAnswer := &entity.Answer{ID: "10020000000000001", QuestionID: publicQuestion.ID, Status: entity.AnswerStatusAvailable}
	hiddenAnswer := &entity.Answer{ID: "10020000000000002", QuestionID: hiddenQuestion.ID, Status: entity.AnswerStatusAvailable}
	deletedAnswer := &entity.Answer{ID: "10020000000000003", QuestionID: publicQuestion.ID, Status: entity.AnswerStatusDeleted}
	for _, answer := range []*entity.Answer{publicAnswer, hiddenAnswer, deletedAnswer} {
		if _, err := db.Context(ctx).Insert(answer); err != nil {
			t.Fatalf("insert answer: %v", err)
		}
	}

	syncer := NewPluginSyncer(&data.Data{DB: db})
	questions, err := syncer.GetQuestionsPage(ctx, 1, 10)
	if err != nil {
		t.Fatalf("get questions: %v", err)
	}
	if len(questions) != 1 || questions[0].ObjectID != publicQuestion.ID {
		t.Fatalf("question documents = %#v, want only the public question", questions)
	}
	answers, err := syncer.GetAnswersPage(ctx, 1, 10)
	if err != nil {
		t.Fatalf("get answers: %v", err)
	}
	if len(answers) != 1 || answers[0].ObjectID != publicAnswer.ID {
		t.Fatalf("answer documents = %#v, want only the public answer", answers)
	}
}

func TestRemoveRestrictedContentsRemovesHistoricalSearchDocuments(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question), new(entity.Answer)); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	hiddenQuestion := &entity.Question{ID: "10010000000000011", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide}
	publicQuestion := &entity.Question{ID: "10010000000000012", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow}
	for _, question := range []*entity.Question{hiddenQuestion, publicQuestion} {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question: %v", err)
		}
	}
	hiddenAnswer := &entity.Answer{ID: "10020000000000011", QuestionID: hiddenQuestion.ID, Status: entity.AnswerStatusAvailable}
	deletedAnswer := &entity.Answer{ID: "10020000000000012", QuestionID: publicQuestion.ID, Status: entity.AnswerStatusDeleted}
	publicAnswer := &entity.Answer{ID: "10020000000000013", QuestionID: publicQuestion.ID, Status: entity.AnswerStatusAvailable}
	for _, answer := range []*entity.Answer{hiddenAnswer, deletedAnswer, publicAnswer} {
		if _, err := db.Context(ctx).Insert(answer); err != nil {
			t.Fatalf("insert answer: %v", err)
		}
	}

	search := &cleanupTestSearch{deleted: map[string]bool{}}
	if err := RemoveRestrictedContents(ctx, &data.Data{DB: db}, search); err != nil {
		t.Fatalf("remove restricted contents: %v", err)
	}
	for _, objectID := range []string{hiddenQuestion.ID, hiddenAnswer.ID, deletedAnswer.ID} {
		if !search.deleted[objectID] {
			t.Fatalf("restricted search document %s was not deleted: %v", objectID, search.deleted)
		}
	}
	for _, objectID := range []string{publicQuestion.ID, publicAnswer.ID} {
		if search.deleted[objectID] {
			t.Fatalf("public search document %s was deleted: %v", objectID, search.deleted)
		}
	}
}
