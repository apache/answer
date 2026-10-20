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

package vector_search_sync

import (
	"context"
	"strings"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
)

func TestPluginSyncerIndexesOnlyPublicContent(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question), new(entity.Answer), new(entity.Comment)); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	questions := []*entity.Question{
		{ID: "10010000000000001", Title: "public", OriginalText: "PUBLIC-QUESTION", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow},
		{ID: "10010000000000002", Title: "hidden", OriginalText: "HIDDEN-QUESTION", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		{ID: "10010000000000003", Title: "pending", OriginalText: "PENDING-QUESTION", Status: entity.QuestionStatusPending, Show: entity.QuestionShow},
	}
	for _, question := range questions {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question %s: %v", question.ID, err)
		}
	}
	answers := []*entity.Answer{
		{ID: "10020000000000001", QuestionID: questions[0].ID, OriginalText: "PUBLIC-ANSWER", Status: entity.AnswerStatusAvailable},
		{ID: "10020000000000002", QuestionID: questions[0].ID, OriginalText: "DELETED-ANSWER", Status: entity.AnswerStatusDeleted},
		{ID: "10020000000000003", QuestionID: questions[1].ID, OriginalText: "HIDDEN-ANSWER", Status: entity.AnswerStatusAvailable},
	}
	for _, answer := range answers {
		if _, err := db.Context(ctx).Insert(answer); err != nil {
			t.Fatalf("insert answer %s: %v", answer.ID, err)
		}
	}
	comments := []*entity.Comment{
		{ID: "10030000000000001", ObjectID: questions[0].ID, QuestionID: questions[0].ID, OriginalText: "PUBLIC-COMMENT", Status: entity.CommentStatusAvailable},
		{ID: "10030000000000002", ObjectID: questions[0].ID, QuestionID: questions[0].ID, OriginalText: "DELETED-COMMENT", Status: entity.CommentStatusDeleted},
	}
	for _, comment := range comments {
		if _, err := db.Context(ctx).Insert(comment); err != nil {
			t.Fatalf("insert comment %s: %v", comment.ID, err)
		}
	}

	syncer := NewPluginSyncer(&data.Data{DB: db})
	questionContents, err := syncer.GetQuestionsPage(ctx, 1, 10)
	if err != nil {
		t.Fatalf("get question page: %v", err)
	}
	if len(questionContents) != 1 {
		t.Fatalf("indexed question count = %d, want 1", len(questionContents))
	}
	content := questionContents[0].Content
	for _, restricted := range []string{"HIDDEN-QUESTION", "PENDING-QUESTION", "DELETED-ANSWER", "DELETED-COMMENT"} {
		if strings.Contains(content, restricted) {
			t.Fatalf("question vector content contains restricted marker %q: %s", restricted, content)
		}
	}

	answerContents, err := syncer.GetAnswersPage(ctx, 1, 10)
	if err != nil {
		t.Fatalf("get answer page: %v", err)
	}
	if len(answerContents) != 1 || !strings.Contains(answerContents[0].Content, "PUBLIC-ANSWER") {
		t.Fatalf("answer vectors = %#v, want only the public answer", answerContents)
	}
}

func TestBuildQuestionContentByIDDeletesRestrictedQuestion(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question)); err != nil {
		t.Fatalf("create question table: %v", err)
	}
	question := &entity.Question{ID: "10010000000000010", Title: "hidden", OriginalText: "secret", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide}
	if _, err := db.Context(ctx).Insert(question); err != nil {
		t.Fatalf("insert hidden question: %v", err)
	}

	content, err := BuildQuestionContentByID(ctx, &data.Data{DB: db}, question.ID)
	if err != nil {
		t.Fatalf("build content: %v", err)
	}
	if content != nil {
		t.Fatalf("restricted question produced vector content: %#v", content)
	}
}
