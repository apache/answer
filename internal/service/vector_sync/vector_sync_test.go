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

package vector_sync

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/plugin"
)

type cleanupTestVectorSearch struct {
	deleted map[string]bool
	updated map[string]bool
}

func (*cleanupTestVectorSearch) Info() plugin.Info { return plugin.Info{} }
func (*cleanupTestVectorSearch) Description() plugin.VectorSearchDesc {
	return plugin.VectorSearchDesc{}
}
func (*cleanupTestVectorSearch) RegisterSyncer(context.Context, plugin.VectorSearchSyncer) {}
func (*cleanupTestVectorSearch) SearchSimilar(context.Context, string, int) ([]plugin.VectorSearchResult, error) {
	return nil, nil
}
func (s *cleanupTestVectorSearch) UpdateContent(_ context.Context, content *plugin.VectorSearchContent) error {
	if s.updated != nil {
		s.updated[content.ObjectID] = true
	}
	return nil
}
func (s *cleanupTestVectorSearch) DeleteContent(_ context.Context, objectID string) error {
	s.deleted[objectID] = true
	return nil
}

func TestRemoveRestrictedContentsRemovesHistoricalVectors(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question), new(entity.Answer), new(entity.Comment)); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	hiddenQuestion := &entity.Question{ID: "10010000000000001", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide}
	publicQuestion := &entity.Question{ID: "10010000000000002", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow}
	for _, question := range []*entity.Question{hiddenQuestion, publicQuestion} {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question: %v", err)
		}
	}
	hiddenAnswer := &entity.Answer{ID: "10020000000000001", QuestionID: hiddenQuestion.ID, Status: entity.AnswerStatusAvailable}
	deletedAnswer := &entity.Answer{ID: "10020000000000002", QuestionID: publicQuestion.ID, Status: entity.AnswerStatusDeleted}
	publicAnswer := &entity.Answer{ID: "10020000000000003", QuestionID: publicQuestion.ID, Status: entity.AnswerStatusAvailable}
	for _, answer := range []*entity.Answer{hiddenAnswer, deletedAnswer, publicAnswer} {
		if _, err := db.Context(ctx).Insert(answer); err != nil {
			t.Fatalf("insert answer: %v", err)
		}
	}
	deletedComment := &entity.Comment{ID: "10030000000000001", ObjectID: publicAnswer.ID, QuestionID: publicQuestion.ID, Status: entity.CommentStatusDeleted}
	if _, err := db.Context(ctx).Insert(deletedComment); err != nil {
		t.Fatalf("insert deleted comment: %v", err)
	}

	vector := &cleanupTestVectorSearch{deleted: map[string]bool{}}
	if err := RemoveRestrictedContents(ctx, &data.Data{DB: db}, vector); err != nil {
		t.Fatalf("remove restricted contents: %v", err)
	}
	for _, objectID := range []string{hiddenQuestion.ID, hiddenAnswer.ID, deletedAnswer.ID, publicQuestion.ID, publicAnswer.ID} {
		if !vector.deleted[objectID] {
			t.Fatalf("restricted vector %s was not deleted: %v", objectID, vector.deleted)
		}
	}
	if vector.deleted["not-a-vector"] {
		t.Fatalf("unexpected vector deletion: %v", vector.deleted)
	}
}

func TestHandleOnceRemovesAnswerVectorsWhenQuestionBecomesHidden(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question), new(entity.Answer), new(entity.Comment)); err != nil {
		t.Fatalf("create tables: %v", err)
	}
	question := &entity.Question{ID: "10010000000000010", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide}
	answer := &entity.Answer{ID: "10020000000000010", QuestionID: question.ID, Status: entity.AnswerStatusAvailable}
	if _, err := db.Context(ctx).Insert(question); err != nil {
		t.Fatalf("insert hidden question: %v", err)
	}
	if _, err := db.Context(ctx).Insert(answer); err != nil {
		t.Fatalf("insert answer: %v", err)
	}

	vector := &cleanupTestVectorSearch{deleted: map[string]bool{}}
	if err := handleOnce(ctx, &data.Data{DB: db}, vector, ActionUpsert, ObjectTypeQuestion, question.ID); err != nil {
		t.Fatalf("handle hidden question upsert: %v", err)
	}
	if !vector.deleted[question.ID] || !vector.deleted[answer.ID] {
		t.Fatalf("hidden question must remove question and answer vectors: %v", vector.deleted)
	}
}

func TestHandleOnceRestoresAnswerVectorsWhenQuestionIsShown(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question), new(entity.Answer), new(entity.Comment)); err != nil {
		t.Fatalf("create tables: %v", err)
	}
	question := &entity.Question{ID: "10010000000000011", OriginalText: "question", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow}
	answer := &entity.Answer{ID: "10020000000000011", QuestionID: question.ID, OriginalText: "answer", Status: entity.AnswerStatusAvailable}
	if _, err := db.Context(ctx).Insert(question); err != nil {
		t.Fatalf("insert shown question: %v", err)
	}
	if _, err := db.Context(ctx).Insert(answer); err != nil {
		t.Fatalf("insert answer: %v", err)
	}

	vector := &cleanupTestVectorSearch{deleted: map[string]bool{}, updated: map[string]bool{}}
	if err := handleOnce(ctx, &data.Data{DB: db}, vector, ActionRestoreQuestion, ObjectTypeQuestion, question.ID); err != nil {
		t.Fatalf("restore shown question: %v", err)
	}
	if !vector.updated[question.ID] || !vector.updated[answer.ID] {
		t.Fatalf("shown question must restore question and answer vectors: %v", vector.updated)
	}
}
