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

package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/segmentfault/pacman/log"
)

// VectorSearchResult holds a single similarity search result returned by a VectorSearch plugin.
type VectorSearchResult struct {
	// ObjectID is the unique identifier of the matched object (question ID or answer ID).
	ObjectID string `json:"object_id"`
	// ObjectType is "question" or "answer".
	ObjectType string `json:"object_type"`
	// Metadata is a JSON string containing VectorSearchMetadata for link composition and content retrieval.
	Metadata string `json:"metadata"`
	// Score is the cosine similarity score (0-1).
	Score float64 `json:"score"`
}

// VectorSearchContent is the document structure passed to plugins for indexing.
type VectorSearchContent struct {
	// ObjectID is the unique identifier (question ID or answer ID).
	ObjectID string `json:"objectID"`
	// ObjectType is "question" or "answer".
	ObjectType string `json:"objectType"`
	// Title is the question title.
	Title string `json:"title"`
	// Content is the aggregated text to be embedded (question body + answers + comments).
	Content string `json:"content"`
	// Metadata is a JSON string containing VectorSearchMetadata.
	Metadata string `json:"metadata"`
}

// VectorSearchDesc describes the vector search engine for display purposes.
type VectorSearchDesc struct {
	// Icon is an SVG icon for display. Optional.
	Icon string `json:"icon"`
	// Link is the URL of the vector search engine. Optional.
	Link string `json:"link"`
}

// VectorSearchMetadata holds IDs for URI composition and content retrieval at query time.
// Shared between plugins and the core MCP controller.
type VectorSearchMetadata struct {
	QuestionID string                        `json:"question_id"`
	AnswerID   string                        `json:"answer_id,omitempty"`
	Answers    []VectorSearchMetadataAnswer  `json:"answers,omitempty"`
	Comments   []VectorSearchMetadataComment `json:"comments,omitempty"`
}

// VectorSearchMetadataAnswer stores answer ID and its comment IDs in metadata.
type VectorSearchMetadataAnswer struct {
	AnswerID string                        `json:"answer_id"`
	Comments []VectorSearchMetadataComment `json:"comments,omitempty"`
}

// VectorSearchMetadataComment stores a comment ID in metadata.
type VectorSearchMetadataComment struct {
	CommentID string `json:"comment_id"`
}

// VectorSearch is the plugin interface for vector/semantic search engines.
// Plugins implementing this interface manage their own vector storage, embedding computation,
// data synchronization schedule, and similarity search.
type VectorSearch interface {
	Base

	// Description returns metadata about the vector search engine.
	Description() VectorSearchDesc

	// RegisterSyncer is called by the core to provide a data syncer.
	// The plugin should store the syncer and use it to bulk-sync content
	// (typically in a background goroutine).
	RegisterSyncer(ctx context.Context, syncer VectorSearchSyncer)

	// SearchSimilar performs a semantic similarity search.
	// The plugin is responsible for embedding the query text and searching its vector store.
	// Returns up to topK results sorted by similarity score (descending).
	SearchSimilar(ctx context.Context, query string, topK int) ([]VectorSearchResult, error)

	// UpdateContent upserts a single document in the vector store.
	// Called by the core on incremental content changes.
	UpdateContent(ctx context.Context, content *VectorSearchContent) error

	// DeleteContent removes a document from the vector store by object ID.
	DeleteContent(ctx context.Context, objectID string) error
}

// VectorSearchSyncer is implemented by the core and provided to plugins via RegisterSyncer.
// Plugins call these methods to pull all content for bulk indexing.
type VectorSearchSyncer interface {
	// GetQuestionsPage returns a page of questions with aggregated text (title + body + answers + comments).
	GetQuestionsPage(ctx context.Context, page, pageSize int) ([]*VectorSearchContent, error)
	// GetAnswersPage returns a page of answers with aggregated text (answer body + parent question title + comments).
	GetAnswersPage(ctx context.Context, page, pageSize int) ([]*VectorSearchContent, error)
}

var (
	// CallVectorSearch is a function that calls all registered VectorSearch plugins.
	CallVectorSearch,
	registerVectorSearch = MakePlugin[VectorSearch](false)
)

// GenerateEmbedding is a base utility function that generates an embedding vector
// using an OpenAI-compatible API. Plugins that don't have a built-in vectorizer
// (most vector databases) can call this function with their own credentials.
// Plugins with built-in vectorizers (e.g., Weaviate) can skip this and use their own.
//
// Parameters:
//   - ctx: context for cancellation
//   - apiHost: the API base URL (e.g. "https://api.openai.com"); "/v1" is appended if missing
//   - apiKey: the API key for authentication
//   - model: the embedding model name (e.g. "text-embedding-3-small")
//   - text: the text to embed
//
// Returns the embedding vector as []float32, or an error.
func GenerateEmbedding(ctx context.Context, apiHost, apiKey, model, text string) ([]float32, error) {
	if model == "" {
		return nil, fmt.Errorf("embedding model is not configured")
	}
	if text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = apiHost
	if !strings.HasSuffix(config.BaseURL, "/v1") {
		config.BaseURL += "/v1"
	}

	log.Debugf("embedding: requesting model=%s baseURL=%s textLen=%d", model, config.BaseURL, len(text))

	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input: []string{text},
		Model: openai.EmbeddingModel(model),
	})
	if err != nil {
		log.Errorf("embedding: request failed model=%s baseURL=%s err=%v", model, config.BaseURL, err)
		return nil, fmt.Errorf("create embeddings failed: %w", err)
	}
	if len(resp.Data) == 0 {
		log.Errorf("embedding: no data returned model=%s baseURL=%s", model, config.BaseURL)
		return nil, fmt.Errorf("no embedding returned")
	}

	log.Debugf("embedding: success model=%s dimensions=%d usage={prompt=%d,total=%d}",
		model, len(resp.Data[0].Embedding), resp.Usage.PromptTokens, resp.Usage.TotalTokens)
	return resp.Data[0].Embedding, nil
}
