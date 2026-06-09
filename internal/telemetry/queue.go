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

package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// WrapQueueHandler wraps a queue handler function to create a consumer span linked to the producer trace.
func WrapQueueHandler[T any](spanName, destination string, handler func(context.Context, T) error) func(context.Context, T) error {
	return func(ctx context.Context, msg T) error {
		link := trace.Link{SpanContext: trace.SpanContextFromContext(ctx)}
		ctx, span := otel.Tracer(Scope).Start(context.Background(), spanName,
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithLinks(link),
		)
		span.SetAttributes(
			attribute.String("messaging.system", "inproc"),
			attribute.String("messaging.destination.name", destination),
			attribute.String("messaging.operation.type", "process"),
		)
		defer span.End()

		err := handler(ctx, msg)
		if err != nil {
			span.SetStatus(codes.Error, "")
			span.SetAttributes(attribute.String("error.type", classifyQueueErr(err)))
		}
		return err
	}
}

func classifyQueueErr(err error) string {
	if err == nil {
		return ""
	}
	return "handler_error"
}
