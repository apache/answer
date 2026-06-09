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
	"github.com/apache/answer/internal/telemetry"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	ActionUpsert = "upsert"
	ActionDelete = "delete"

	ObjectTypeQuestion = "question"
	ObjectTypeAnswer   = "answer"
)

const maxRetry = 3

type Task struct {
	Action             string
	ObjectType         string
	ObjectID           string
	PropagationCarrier map[string]string
}

type Service queue.Service[*Task]

func NewService(data *data.Data) Service {
	q := queue.New[*Task]("vector_sync", 128)
	q.RegisterHandler(func(ctx context.Context, msg *Task) error {
		return handleWithSpan(ctx, data, msg)
	})
	return q
}

func handleWithSpan(ctx context.Context, data *data.Data, msg *Task) error {
	link := trace.Link{SpanContext: trace.SpanContextFromContext(ctx)}
	ctx, span := otel.Tracer(telemetry.Scope).Start(context.Background(), "vector_sync process",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithLinks(link),
	)
	span.SetAttributes(
		attribute.String("messaging.system", "inproc"),
		attribute.String("messaging.destination.name", "vector_sync"),
		attribute.String("messaging.operation.type", "process"),
	)
	defer span.End()

	err := handle(ctx, data, msg, span)
	if err != nil {
		span.SetStatus(codes.Error, "")
		span.SetAttributes(attribute.String("error.type", "handler_error"))
	}
	return err
}

func handle(ctx context.Context, data *data.Data, msg *Task, span trace.Span) error {
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
		span.SetAttributes(attribute.Int("answer.queue.retry_count", attempt-1))
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

func (t *Task) getPropagationCarrier() map[string]string { return t.PropagationCarrier }
func (t *Task) initPropagationCarrier() map[string]string {
	if t.PropagationCarrier == nil {
		t.PropagationCarrier = make(map[string]string)
	}
	return t.PropagationCarrier
}
