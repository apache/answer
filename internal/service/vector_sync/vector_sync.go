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

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/queue"
	"github.com/apache/answer/internal/repo/vector_search_sync"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"
)

const (
	ActionUpsert = "upsert"
	ActionDelete = "delete"

	ObjectTypeQuestion = "question"
	ObjectTypeAnswer   = "answer"
)

const maxRetry = 3

type Task struct {
	Action     string
	ObjectType string
	ObjectID   string
}

type Service queue.Service[*Task]

func NewService(data *data.Data) Service {
	q := queue.New[*Task]("vector_sync", 128)
	q.RegisterHandler(func(ctx context.Context, msg *Task) error {
		return handle(ctx, data, msg)
	})
	return q
}

func handle(ctx context.Context, data *data.Data, msg *Task) error {
	if msg == nil || msg.ObjectID == "" {
		return nil
	}

	var vectorSearch plugin.VectorSearch
	_ = plugin.CallVectorSearch(func(vs plugin.VectorSearch) error {
		vectorSearch = vs
		return nil
	})
	if vectorSearch == nil {
		return nil
	}

	objectID := uid.DeShortID(msg.ObjectID)
	var lastErr error
	for attempt := 1; attempt <= maxRetry; attempt++ {
		err := handleOnce(ctx, data, vectorSearch, msg.Action, msg.ObjectType, objectID)
		if err == nil {
			return nil
		}
		lastErr = err
		log.Warnf("vector sync failed: action=%s object_type=%s object_id=%s attempt=%d err=%v",
			msg.Action, msg.ObjectType, objectID, attempt, err)
	}
	return lastErr
}

func handleOnce(ctx context.Context, data *data.Data, vectorSearch plugin.VectorSearch,
	action, objectType, objectID string) error {
	if action == ActionDelete {
		return vectorSearch.DeleteContent(ctx, objectID)
	}
	if action != ActionUpsert {
		return nil
	}

	var (
		content *plugin.VectorSearchContent
		err     error
	)
	switch objectType {
	case ObjectTypeQuestion:
		content, err = vector_search_sync.BuildQuestionContentByID(ctx, data, objectID)
	case ObjectTypeAnswer:
		content, err = vector_search_sync.BuildAnswerContentByID(ctx, data, objectID)
	default:
		return nil
	}
	if err != nil {
		return err
	}
	if content == nil {
		return vectorSearch.DeleteContent(ctx, objectID)
	}
	return vectorSearch.UpdateContent(ctx, content)
}
