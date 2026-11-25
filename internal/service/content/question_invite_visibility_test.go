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
	questionrepo "github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/content"
)

func TestQuestionInviteUpdateRejectsHiddenQuestionWithSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question)); err != nil {
		t.Fatalf("create question table: %v", err)
	}

	question := &entity.Question{
		ID:           "10010000000000001",
		UserID:       "10000000000000001",
		Title:        "hidden question",
		InviteUserID: "[\"10000000000000002\"]",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionHide,
	}
	if _, err := db.Context(ctx).Insert(question); err != nil {
		t.Fatalf("insert hidden question: %v", err)
	}

	service := content.NewQuestionService(
		nil,
		questionrepo.NewQuestionRepo(&data.Data{DB: db}, nil),
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	err = service.UpdateQuestionInviteUser(ctx, &schema.QuestionUpdateInviteUser{
		ID:         question.ID,
		InviteUser: []string{"new-invitee"},
		UserID:     "10000000000000003",
	})
	if err == nil {
		t.Fatal("hidden question invite update unexpectedly succeeded")
	}

	stored := &entity.Question{}
	exists, err := db.Context(ctx).ID(question.ID).Get(stored)
	if err != nil || !exists {
		t.Fatalf("load hidden question: exists=%t err=%v", exists, err)
	}
	if stored.InviteUserID != question.InviteUserID {
		t.Fatalf("invite list changed to %q, want %q", stored.InviteUserID, question.InviteUserID)
	}
}
