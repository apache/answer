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

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
)

func TestSearchListExcludesRestrictedQuestionsForOtherUsers(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question), new(entity.Collection)); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	questions := []*entity.Question{
		{ID: "1001", UserID: "author", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow},
		{ID: "1002", UserID: "author", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		{ID: "1003", UserID: "author", Status: entity.QuestionStatusPending, Show: entity.QuestionShow},
		{ID: "1004", UserID: "viewer", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		{ID: "1005", UserID: "author", Status: entity.QuestionStatusClosed, Show: entity.QuestionShow},
	}
	for _, question := range questions {
		if _, err := db.Context(ctx).Insert(question); err != nil {
			t.Fatalf("insert question %s: %v", question.ID, err)
		}
		if _, err := db.Context(ctx).Insert(&entity.Collection{ID: "200" + question.ID[3:], UserID: "viewer", ObjectID: question.ID}); err != nil {
			t.Fatalf("insert collection for %s: %v", question.ID, err)
		}
	}

	repo := &collectionRepo{data: &data.Data{DB: db}}
	collections, total, err := repo.SearchList(ctx, &entity.CollectionSearch{
		Collection: entity.Collection{UserID: "viewer"},
		Page:       1,
		PageSize:   10,
	})
	if err != nil {
		t.Fatalf("search collections: %v", err)
	}
	if total != 3 || len(collections) != 3 {
		t.Fatalf("visible collections = %d (total %d), want 3", len(collections), total)
	}
	visible := map[string]bool{}
	for _, collection := range collections {
		visible[collection.ObjectID] = true
	}
	if !visible["1001"] || !visible["1004"] || !visible["1005"] || visible["1002"] || visible["1003"] {
		t.Fatalf("visible collections = %v, want 1001, 1004, and 1005", visible)
	}
}
