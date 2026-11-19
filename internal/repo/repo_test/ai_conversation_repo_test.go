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

package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/ai_conversation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_aiConversationRepo_GetRecordsByConversationID_KeepsInsertionOrder(t *testing.T) {
	repo := ai_conversation.NewAIConversationRepo(testDataSource)
	conversationID := "conversation-record-order"

	err := repo.CreateConversation(context.TODO(), &entity.AIConversation{
		ConversationID: conversationID,
		Topic:          "question one",
		UserID:         "1",
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, repo.DeleteConversation(context.TODO(), conversationID))
	}()

	// A turn's two records are written within the same second, so every
	// record here shares one timestamp and only the primary key separates them.
	sameSecond := time.Now().Truncate(time.Second)
	roles := []string{"user", "assistant", "user", "assistant"}
	for i, role := range roles {
		_, err = testDataSource.DB.Context(context.TODO()).NoAutoTime().Insert(&entity.AIConversationRecord{
			CreatedAt:        sameSecond,
			UpdatedAt:        sameSecond,
			ConversationID:   conversationID,
			ChatCompletionID: "chatcmpl-" + string(rune('a'+i/2)),
			Role:             role,
			Content:          role,
		})
		require.NoError(t, err)
	}

	records, err := repo.GetRecordsByConversationID(context.TODO(), conversationID)
	require.NoError(t, err)
	require.Len(t, records, len(roles))
	for i, record := range records {
		assert.Equal(t, roles[i], record.Role, "record %d", i)
		if i > 0 {
			assert.Greater(t, record.ID, records[i-1].ID, "record %d", i)
		}
	}
}
