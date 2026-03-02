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

package entity

import "time"

// Embedding stores vector embeddings for questions or answers.
type Embedding struct {
	ID          int       `xorm:"not null pk autoincr INT(11) id"`
	CreatedAt   time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt   time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
	ObjectID    string    `xorm:"not null BIGINT(20) INDEX object_id unique(object_embedding)"`
	ObjectType  string    `xorm:"not null default '' VARCHAR(20) object_type unique(object_embedding)"`
	ContentHash string    `xorm:"not null default '' VARCHAR(64) content_hash"`
	Metadata    string    `xorm:"not null MEDIUMTEXT metadata"`
	Embedding   string    `xorm:"not null MEDIUMTEXT embedding"`
	Dimensions  int       `xorm:"not null default 0 INT(11) dimensions"`
}

// TableName returns the table name
func (Embedding) TableName() string {
	return "embedding"
}

// EmbeddingMetadata holds IDs for URI composition and content retrieval at query time.
type EmbeddingMetadata struct {
	QuestionID string                     `json:"question_id"`
	AnswerID   string                     `json:"answer_id,omitempty"`
	Answers    []EmbeddingMetadataAnswer  `json:"answers,omitempty"`
	Comments   []EmbeddingMetadataComment `json:"comments,omitempty"`
}

// EmbeddingMetadataAnswer stores answer ID and comment IDs in metadata.
type EmbeddingMetadataAnswer struct {
	AnswerID string                     `json:"answer_id"`
	Comments []EmbeddingMetadataComment `json:"comments,omitempty"`
}

// EmbeddingMetadataComment stores comment ID in metadata.
type EmbeddingMetadataComment struct {
	CommentID string `json:"comment_id"`
}
