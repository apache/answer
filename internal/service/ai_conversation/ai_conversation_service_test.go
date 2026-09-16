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

package ai_conversation

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/ai_conversation"
	"github.com/apache/answer/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRepo serves one conversation and its records in whatever order the
// test hands them over, standing in for the database.
type fakeRepo struct {
	ai_conversation.AIConversationRepo
	conversation *entity.AIConversation
	records      []*entity.AIConversationRecord
}

func (f *fakeRepo) GetConversation(_ context.Context, conversationID string) (*entity.AIConversation, bool, error) {
	if f.conversation == nil || f.conversation.ConversationID != conversationID {
		return nil, false, nil
	}
	return f.conversation, true, nil
}

func (f *fakeRepo) GetRecordsByConversationID(_ context.Context, _ string) ([]*entity.AIConversationRecord, error) {
	return f.records, nil
}

func TestGetConversationDetail_TopicReplacesOpeningQuestionNotFirstRecord(t *testing.T) {
	const prompt = "You are an assistant. User question: what is the topic?"
	repo := &fakeRepo{
		conversation: &entity.AIConversation{ConversationID: "c1", Topic: "what is the topic?", UserID: "u1"},
		records: []*entity.AIConversationRecord{
			{ID: 2, Role: "assistant", Content: "the answer"},
			{ID: 1, Role: "user", Content: prompt},
			{ID: 3, Role: "user", Content: "a follow-up"},
			{ID: 4, Role: "assistant", Content: "another answer"},
		},
	}
	service := NewAIConversationService(repo, nil)

	resp, exist, err := service.GetConversationDetail(context.TODO(), &schema.AIConversationDetailReq{
		ConversationID: "c1",
		UserID:         "u1",
	})
	require.NoError(t, err)
	require.True(t, exist)
	require.Len(t, resp.Records, 4)

	assert.Equal(t, "the answer", resp.Records[0].Content, "an answer that sorts first must keep its content")
	assert.Equal(t, "what is the topic?", resp.Records[1].Content, "the opening question shows the topic, not the prompt")
	assert.Equal(t, "a follow-up", resp.Records[2].Content)
	assert.Equal(t, "another answer", resp.Records[3].Content)
}

func TestFirstUserRecordID(t *testing.T) {
	assert.Equal(t, 0, firstUserRecordID(nil))
	assert.Equal(t, 0, firstUserRecordID([]*entity.AIConversationRecord{{ID: 1, Role: "assistant"}}))
	assert.Equal(t, 5, firstUserRecordID([]*entity.AIConversationRecord{
		{ID: 4, Role: "assistant"}, {ID: 5, Role: "user"}, {ID: 6, Role: "user"},
	}))
}
