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

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// AIConversationRepo
type AIConversationRepo interface {
	CreateConversation(ctx context.Context, conversation *entity.AIConversation) error
	GetConversation(ctx context.Context, conversationID string) (*entity.AIConversation, bool, error)
	UpdateConversation(ctx context.Context, conversation *entity.AIConversation) error
	GetConversationsPage(ctx context.Context, page, pageSize int, cond *entity.AIConversation) (list []*entity.AIConversation, total int64, err error)
	CreateRecord(ctx context.Context, record *entity.AIConversationRecord) error
	GetRecordsByConversationID(ctx context.Context, conversationID string) ([]*entity.AIConversationRecord, error)
	UpdateRecordVote(ctx context.Context, cond *entity.AIConversationRecord) error
	GetRecord(ctx context.Context, recordID int) (*entity.AIConversationRecord, bool, error)
	GetRecordByChatCompletionID(ctx context.Context, role, chatCompletionID string) (*entity.AIConversationRecord, bool, error)
	GetConversationsForAdmin(ctx context.Context, page, pageSize int, cond *entity.AIConversation) (list []*entity.AIConversation, total int64, err error)
	GetConversationWithVoteStats(ctx context.Context, conversationID string) (helpful, unhelpful int64, err error)
	DeleteConversation(ctx context.Context, conversationID string) error
}

type aiConversationRepo struct {
	data *data.Data
}

// NewAIConversationRepo new AIConversationRepo
func NewAIConversationRepo(data *data.Data) AIConversationRepo {
	return &aiConversationRepo{
		data: data,
	}
}

// CreateConversation creates a conversation
func (r *aiConversationRepo) CreateConversation(ctx context.Context, conversation *entity.AIConversation) error {
	_, err := r.data.DB.Context(ctx).Insert(conversation)
	if err != nil {
		log.Errorf("create ai conversation failed: %v", err)
		return err
	}
	return nil
}

// GetConversation gets a conversation
func (r *aiConversationRepo) GetConversation(ctx context.Context, conversationID string) (*entity.AIConversation, bool, error) {
	conversation := &entity.AIConversation{}
	exist, err := r.data.DB.Context(ctx).Where(builder.Eq{"conversation_id": conversationID}).Get(conversation)
	if err != nil {
		log.Errorf("get ai conversation failed: %v", err)
		return nil, false, err
	}
	return conversation, exist, nil
}

// UpdateConversation updates a conversation
func (r *aiConversationRepo) UpdateConversation(ctx context.Context, conversation *entity.AIConversation) error {
	_, err := r.data.DB.Context(ctx).ID(conversation.ID).Update(conversation)
	if err != nil {
		log.Errorf("update ai conversation failed: %v", err)
		return err
	}
	return nil
}

// GetConversationsPage get conversations by user ID
func (r *aiConversationRepo) GetConversationsPage(ctx context.Context, page, pageSize int, cond *entity.AIConversation) (list []*entity.AIConversation, total int64, err error) {
	list = make([]*entity.AIConversation, 0)
	total, err = pager.Help(page, pageSize, &list, cond, r.data.DB.Context(ctx).Desc("id"))
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return list, total, err
}

// CreateRecord creates a conversation record
func (r *aiConversationRepo) CreateRecord(ctx context.Context, record *entity.AIConversationRecord) error {
	_, err := r.data.DB.Context(ctx).Insert(record)
	if err != nil {
		log.Errorf("create ai conversation record failed: %v", err)
		return err
	}
	return nil
}

// GetRecordsByConversationID get records by conversation ID
func (r *aiConversationRepo) GetRecordsByConversationID(ctx context.Context, conversationID string) ([]*entity.AIConversationRecord, error) {
	records := make([]*entity.AIConversationRecord, 0)
	err := r.data.DB.Context(ctx).
		Where(builder.Eq{"conversation_id": conversationID}).
		OrderBy("created_at ASC").
		Find(&records)
	if err != nil {
		log.Errorf("get ai conversation records failed: %v", err)
		return nil, err
	}
	return records, nil
}

// UpdateRecordVote update record vote
func (r *aiConversationRepo) UpdateRecordVote(ctx context.Context, cond *entity.AIConversationRecord) (err error) {
	_, err = r.data.DB.Context(ctx).ID(cond.ID).MustCols("helpful", "unhelpful").Update(cond)
	if err != nil {
		log.Errorf("update ai conversation record vote failed: %v", err)
		return err
	}
	return nil
}

// GetRecord get record
func (r *aiConversationRepo) GetRecord(ctx context.Context, recordID int) (*entity.AIConversationRecord, bool, error) {
	record := &entity.AIConversationRecord{}
	exist, err := r.data.DB.Context(ctx).ID(recordID).Get(record)
	if err != nil {
		log.Errorf("get ai conversation record failed: %v", err)
		return nil, false, err
	}
	return record, exist, nil
}

// GetRecordByChatCompletionID gets record by chat completion ID
func (r *aiConversationRepo) GetRecordByChatCompletionID(ctx context.Context, role, chatCompletionID string) (*entity.AIConversationRecord, bool, error) {
	record := &entity.AIConversationRecord{}
	exist, err := r.data.DB.Context(ctx).Where(builder.Eq{"role": role}).
		Where(builder.Eq{"chat_completion_id": chatCompletionID}).Get(record)
	if err != nil {
		log.Errorf("get ai conversation record by chat completion id failed: %v", err)
		return nil, false, err
	}
	return record, exist, nil
}

// GetConversationsForAdmin gets conversation list for admin
func (r *aiConversationRepo) GetConversationsForAdmin(ctx context.Context, page, pageSize int, cond *entity.AIConversation) (list []*entity.AIConversation, total int64, err error) {
	list = make([]*entity.AIConversation, 0)
	total, err = pager.Help(page, pageSize, &list, cond, r.data.DB.Context(ctx).Desc("id"))
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return list, total, err
}

// GetConversationWithVoteStats gets conversation vote statistics
func (r *aiConversationRepo) GetConversationWithVoteStats(ctx context.Context, conversationID string) (helpful, unhelpful int64, err error) {
	res, err := r.data.DB.Context(ctx).SumsInt(&entity.AIConversationRecord{ConversationID: conversationID}, "helpful", "unhelpful")
	if err != nil {
		log.Errorf("get ai conversation vote stats failed: %v", err)
		return 0, 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if len(res) < 2 {
		log.Errorf("get ai conversation vote stats failed: invalid result length %d", len(res))
		return 0, 0, nil
	}
	return res[0], res[1], nil
}

// DeleteConversation deletes a conversation and its related records
func (r *aiConversationRepo) DeleteConversation(ctx context.Context, conversationID string) error {
	_, err := r.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		if _, err := session.Context(ctx).Where("conversation_id = ?", conversationID).Delete(&entity.AIConversationRecord{}); err != nil {
			log.Errorf("delete ai conversation records failed: %v", err)
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}

		if _, err := session.Context(ctx).Where("conversation_id = ?", conversationID).Delete(&entity.AIConversation{}); err != nil {
			log.Errorf("delete ai conversation failed: %v", err)
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}
