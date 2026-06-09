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

package embedding

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/telemetry"
	"github.com/apache/answer/plugin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// EmbeddingService is a thin facade that delegates semantic search to a VectorSearch plugin.
// If no plugin is enabled, semantic search is unavailable.
type EmbeddingService struct{}

// NewEmbeddingService creates a new EmbeddingService.
func NewEmbeddingService() *EmbeddingService {
	return &EmbeddingService{}
}

// SearchSimilar delegates to the VectorSearch plugin.
// Returns an error if no plugin is enabled.
func (s *EmbeddingService) SearchSimilar(ctx context.Context, query string, topK int) ([]plugin.VectorSearchResult, error) {
	_, span := otel.Tracer(telemetry.Scope).Start(ctx, "vector_search query")
	span.SetAttributes(
		attribute.String("db.system.name", "vector_database"),
		attribute.String("db.operation.name", "search_similar"),
		attribute.Int("answer.vector_search.top_k", topK),
	)
	defer span.End()

	var results []plugin.VectorSearchResult
	var searchErr error
	found := false

	err := plugin.CallVectorSearch(func(vs plugin.VectorSearch) error {
		found = true
		results, searchErr = vs.SearchSimilar(ctx, query, topK)
		return nil
	})
	if err != nil {
		span.SetStatus(codes.Error, "")
		span.SetAttributes(attribute.String("error.type", "plugin_unavailable"))
		return nil, fmt.Errorf("call vector search plugin failed: %w", err)
	}
	if !found {
		span.SetStatus(codes.Error, "")
		span.SetAttributes(attribute.String("error.type", "plugin_unavailable"))
		return nil, fmt.Errorf("semantic search is not available: no vector search plugin is enabled")
	}
	if searchErr != nil {
		span.SetStatus(codes.Error, "")
		span.SetAttributes(attribute.String("error.type", "search_error"))
		return nil, searchErr
	}
	return results, nil
}
