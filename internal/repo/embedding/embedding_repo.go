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
	"encoding/json"
	"math"
	"sort"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
)

// EmbeddingRepo defines the interface for embedding data access.
type EmbeddingRepo interface {
	Upsert(ctx context.Context, emb *entity.Embedding) error
	GetByObjectID(ctx context.Context, objectID, objectType string) (*entity.Embedding, bool, error)
	GetAll(ctx context.Context) ([]*entity.Embedding, error)
	SearchSimilar(ctx context.Context, queryVector []float32, topK int) ([]SimilarResult, error)
	DeleteByObjectID(ctx context.Context, objectID, objectType string) error
	Count(ctx context.Context) (int64, error)
}

// SimilarResult holds a similarity search result.
type SimilarResult struct {
	ObjectID   string  `json:"object_id"`
	ObjectType string  `json:"object_type"`
	Metadata   string  `json:"metadata"`
	Score      float64 `json:"score"`
}

type embeddingRepo struct {
	data *data.Data
}

// NewEmbeddingRepo creates a new EmbeddingRepo.
func NewEmbeddingRepo(data *data.Data) EmbeddingRepo {
	return &embeddingRepo{data: data}
}

// Upsert inserts or updates an embedding by (object_id, object_type).
func (r *embeddingRepo) Upsert(ctx context.Context, emb *entity.Embedding) error {
	existing := &entity.Embedding{}
	exist, err := r.data.DB.Context(ctx).
		Where(builder.Eq{"object_id": emb.ObjectID, "object_type": emb.ObjectType}).
		Get(existing)
	if err != nil {
		log.Errorf("check embedding existence failed: %v", err)
		return err
	}

	if exist {
		emb.ID = existing.ID
		_, err = r.data.DB.Context(ctx).ID(existing.ID).
			Cols("content_hash", "metadata", "embedding", "dimensions", "updated_at").
			Update(emb)
		if err != nil {
			log.Errorf("update embedding failed: %v", err)
			return err
		}
		return nil
	}

	_, err = r.data.DB.Context(ctx).Insert(emb)
	if err != nil {
		log.Errorf("insert embedding failed: %v", err)
		return err
	}
	return nil
}

// GetByObjectID returns an embedding by object ID and type.
func (r *embeddingRepo) GetByObjectID(ctx context.Context, objectID, objectType string) (*entity.Embedding, bool, error) {
	emb := &entity.Embedding{}
	exist, err := r.data.DB.Context(ctx).
		Where(builder.Eq{"object_id": objectID, "object_type": objectType}).
		Get(emb)
	if err != nil {
		log.Errorf("get embedding failed: %v", err)
		return nil, false, err
	}
	return emb, exist, nil
}

// GetAll returns all embeddings.
func (r *embeddingRepo) GetAll(ctx context.Context) ([]*entity.Embedding, error) {
	var list []*entity.Embedding
	err := r.data.DB.Context(ctx).Find(&list)
	if err != nil {
		log.Errorf("get all embeddings failed: %v", err)
		return nil, err
	}
	return list, nil
}

// SearchSimilar performs brute-force cosine similarity search in Go.
func (r *embeddingRepo) SearchSimilar(ctx context.Context, queryVector []float32, topK int) ([]SimilarResult, error) {
	allEmbeddings, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	type scored struct {
		emb   *entity.Embedding
		score float64
	}
	results := make([]scored, 0, len(allEmbeddings))

	for _, emb := range allEmbeddings {
		var vec []float32
		if err := json.Unmarshal([]byte(emb.Embedding), &vec); err != nil {
			log.Warnf("skip embedding id=%d, unmarshal failed: %v", emb.ID, err)
			continue
		}
		score := cosineSimilarity(queryVector, vec)
		results = append(results, scored{emb: emb, score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if topK > len(results) {
		topK = len(results)
	}

	out := make([]SimilarResult, 0, topK)
	for i := 0; i < topK; i++ {
		out = append(out, SimilarResult{
			ObjectID:   results[i].emb.ObjectID,
			ObjectType: results[i].emb.ObjectType,
			Metadata:   results[i].emb.Metadata,
			Score:      results[i].score,
		})
	}
	return out, nil
}

// DeleteByObjectID deletes an embedding by object ID and type.
func (r *embeddingRepo) DeleteByObjectID(ctx context.Context, objectID, objectType string) error {
	_, err := r.data.DB.Context(ctx).
		Where(builder.Eq{"object_id": objectID, "object_type": objectType}).
		Delete(&entity.Embedding{})
	if err != nil {
		log.Errorf("delete embedding failed: %v", err)
		return err
	}
	return nil
}

// Count returns the total number of embeddings.
func (r *embeddingRepo) Count(ctx context.Context) (int64, error) {
	count, err := r.data.DB.Context(ctx).Count(&entity.Embedding{})
	if err != nil {
		log.Errorf("count embeddings failed: %v", err)
		return 0, err
	}
	return count, nil
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dotProduct / denom
}
