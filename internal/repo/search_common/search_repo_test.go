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

package search_common

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/plugin"
)

func TestParseSearchPluginResultExcludesHiddenQuestions(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.Question)); err != nil {
		t.Fatalf("create question table: %v", err)
	}

	hidden := &entity.Question{
		ID:     "10010000000000001",
		Status: entity.QuestionStatusAvailable,
		Show:   entity.QuestionHide,
	}
	if _, err := db.Context(ctx).Insert(hidden); err != nil {
		t.Fatalf("insert hidden question: %v", err)
	}

	repo := &searchRepo{data: &data.Data{DB: db}}
	result, err := repo.ParseSearchPluginResult(ctx, []plugin.SearchResult{{
		ID: hidden.ID, Type: "question",
	}}, []string{"confidential"})
	if err != nil {
		t.Fatalf("parse search result: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("hidden question appeared in search results: %#v", result)
	}
}
