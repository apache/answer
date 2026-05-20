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
	"strings"
	"time"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/ai_conversation"
	"github.com/apache/answer/internal/schema"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// AIConversationService
type AIConversationService interface {
	CreateConversation(ctx context.Context, userID, conversationID, topic string) error
	SaveConversationRecords(ctx context.Context, conversationID, chatcmplID string, records []*ConversationMessage) error
	GetConversationList(ctx context.Context, req *schema.AIConversationListReq) (*pager.PageModel, error)
	GetConversationDetail(ctx context.Context, req *schema.AIConversationDetailReq) (resp *schema.AIConversationDetailResp, exist bool, err error)
	VoteRecord(ctx context.Context, req *schema.AIConversationVoteReq) error
	GetConversationListForAdmin(ctx context.Context, req *schema.AIConversationAdminListReq) (*pager.PageModel, error)
	GetConversationDetailForAdmin(ctx context.Context, req *schema.AIConversationAdminDetailReq) (*schema.AIConversationAdminDetailResp, error)
	DeleteConversationForAdmin(ctx context.Context, req *schema.AIConversationAdminDeleteReq) error
}

// ConversationMessage
type ConversationMessage struct {
	ChatCompletionID string `json:"chat_completion_id"`
	Role             string `json:"role"`
	Content          string `json:"content"`
}

// aiConversationService
type aiConversationService struct {
	aiConversationRepo ai_conversation.AIConversationRepo
	userCommon         *usercommon.UserCommon
}

// NewAIConversationService
func NewAIConversationService(
	aiConversationRepo ai_conversation.AIConversationRepo,
	userCommon *usercommon.UserCommon,
) AIConversationService {
	return &aiConversationService{
		aiConversationRepo: aiConversationRepo,
		userCommon:         userCommon,
	}
}

// CreateConversation
func (s *aiConversationService) CreateConversation(ctx context.Context, userID, conversationID, topic string) error {
	conversation := &entity.AIConversation{
		ConversationID: conversationID,
		Topic:          topic,
		UserID:         userID,
	}
	err := s.aiConversationRepo.CreateConversation(ctx, conversation)
	if err != nil {
		log.Errorf("create conversation failed: %v", err)
		return err
	}

	return nil
}

// SaveConversationRecords
func (s *aiConversationService) SaveConversationRecords(ctx context.Context, conversationID, chatcmplID string, records []*ConversationMessage) error {
	conversation, exist, err := s.aiConversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}

	content := strings.Builder{}

	for _, record := range records {
		if len(record.ChatCompletionID) > 0 {
			continue
		}
		if record.Role == "user" {
			aiRecord := &entity.AIConversationRecord{
				ConversationID:   conversationID,
				ChatCompletionID: chatcmplID,
				Role:             "user",
				Content:          record.Content,
			}

			err = s.aiConversationRepo.CreateRecord(ctx, aiRecord)
			if err != nil {
				log.Errorf("create conversation record failed: %v", err)
				return errors.InternalServer(reason.DatabaseError).WithError(err)
			}
			continue
		}

		content.WriteString(record.Content)
		content.WriteString("\n")
	}
	aiRecord := &entity.AIConversationRecord{
		ConversationID:   conversationID,
		ChatCompletionID: chatcmplID,
		Role:             "assistant",
		Content:          content.String(),
		Helpful:          0,
		Unhelpful:        0,
	}

	err = s.aiConversationRepo.CreateRecord(ctx, aiRecord)
	if err != nil {
		log.Errorf("create conversation record failed: %v", err)
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	conversation.UpdatedAt = time.Now()
	err = s.aiConversationRepo.UpdateConversation(ctx, conversation)
	if err != nil {
		log.Errorf("update conversation failed: %v", err)
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	return nil
}

// GetConversationList
func (s *aiConversationService) GetConversationList(ctx context.Context, req *schema.AIConversationListReq) (*pager.PageModel, error) {
	conversations, total, err := s.aiConversationRepo.GetConversationsPage(ctx, req.Page, req.PageSize, &entity.AIConversation{UserID: req.UserID})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	list := make([]schema.AIConversationListItem, 0, len(conversations))
	for _, conversation := range conversations {
		list = append(list, schema.AIConversationListItem{
			ConversationID: conversation.ConversationID,
			CreatedAt:      conversation.CreatedAt.Unix(),
			Topic:          conversation.Topic,
		})
	}

	return pager.NewPageModel(total, list), nil
}

// GetConversationDetail
func (s *aiConversationService) GetConversationDetail(ctx context.Context, req *schema.AIConversationDetailReq) (
	resp *schema.AIConversationDetailResp, exist bool, err error) {
	conversation, exist, err := s.aiConversationRepo.GetConversation(ctx, req.ConversationID)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist || conversation.UserID != req.UserID {
		return nil, false, nil
	}

	records, err := s.aiConversationRepo.GetRecordsByConversationID(ctx, req.ConversationID)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	recordList := make([]*schema.AIConversationRecord, 0, len(records))
	for i, record := range records {
		if i == 0 {
			record.Content = conversation.Topic
		}
		recordList = append(recordList, &schema.AIConversationRecord{
			ChatCompletionID: record.ChatCompletionID,
			Role:             record.Role,
			Content:          record.Content,
			Helpful:          record.Helpful,
			Unhelpful:        record.Unhelpful,
			CreatedAt:        record.CreatedAt.Unix(),
		})
	}

	return &schema.AIConversationDetailResp{
		ConversationID: conversation.ConversationID,
		Topic:          conversation.Topic,
		Records:        recordList,
		CreatedAt:      conversation.CreatedAt.Unix(),
		UpdatedAt:      conversation.UpdatedAt.Unix(),
	}, true, nil
}

// VoteRecord
func (s *aiConversationService) VoteRecord(ctx context.Context, req *schema.AIConversationVoteReq) error {
	record, exist, err := s.aiConversationRepo.GetRecordByChatCompletionID(ctx, "assistant", req.ChatCompletionID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}

	conversation, exist, err := s.aiConversationRepo.GetConversation(ctx, record.ConversationID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}

	if conversation.UserID != req.UserID {
		return errors.Forbidden(reason.UnauthorizedError)
	}

	if record.Role != "assistant" {
		return errors.BadRequest("Only AI responses can be voted")
	}

	if req.VoteType == "helpful" {
		if req.Cancel {
			record.Helpful = 0
		} else {
			record.Helpful = 1
			record.Unhelpful = 0
		}
	} else {
		if req.Cancel {
			record.Unhelpful = 0
		} else {
			record.Unhelpful = 1
			record.Helpful = 0
		}
	}

	err = s.aiConversationRepo.UpdateRecordVote(ctx, record)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	return nil
}

// GetConversationListForAdmin
func (s *aiConversationService) GetConversationListForAdmin(
	ctx context.Context, req *schema.AIConversationAdminListReq) (*pager.PageModel, error) {
	conversations, total, err := s.aiConversationRepo.GetConversationsForAdmin(ctx, req.Page, req.PageSize, &entity.AIConversation{})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	list := make([]*schema.AIConversationAdminListItem, 0, len(conversations))
	for _, conversation := range conversations {
		userInfo, err := s.getUserInfo(ctx, conversation.UserID)
		if err != nil {
			log.Errorf("get user info failed for user %s: %v", conversation.UserID, err)
			continue
		}

		helpful, unhelpful, err := s.aiConversationRepo.GetConversationWithVoteStats(ctx, conversation.ConversationID)
		if err != nil {
			log.Errorf("get conversation vote stats failed for conversation %s: %v", conversation.ConversationID, err)
			continue
		}

		list = append(list, &schema.AIConversationAdminListItem{
			ID:             conversation.ConversationID,
			Topic:          conversation.Topic,
			UserInfo:       userInfo,
			HelpfulCount:   helpful,
			UnhelpfulCount: unhelpful,
			CreatedAt:      conversation.CreatedAt.Unix(),
		})
	}

	return pager.NewPageModel(total, list), nil
}

// GetConversationDetailForAdmin
func (s *aiConversationService) GetConversationDetailForAdmin(ctx context.Context, req *schema.AIConversationAdminDetailReq) (*schema.AIConversationAdminDetailResp, error) {
	conversation, exist, err := s.aiConversationRepo.GetConversation(ctx, req.ConversationID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return nil, errors.BadRequest(reason.ObjectNotFound)
	}

	userInfo, err := s.getUserInfo(ctx, conversation.UserID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	records, err := s.aiConversationRepo.GetRecordsByConversationID(ctx, req.ConversationID)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err)
	}

	recordList := make([]schema.AIConversationRecord, 0, len(records))
	for i, record := range records {
		if i == 0 {
			record.Content = conversation.Topic
		}
		recordList = append(recordList, schema.AIConversationRecord{
			ChatCompletionID: record.ChatCompletionID,
			Role:             record.Role,
			Content:          record.Content,
			Helpful:          record.Helpful,
			Unhelpful:        record.Unhelpful,
			CreatedAt:        record.CreatedAt.Unix(),
		})
	}

	return &schema.AIConversationAdminDetailResp{
		ConversationID: conversation.ConversationID,
		Topic:          conversation.Topic,
		UserInfo:       userInfo,
		Records:        recordList,
		CreatedAt:      conversation.CreatedAt.Unix(),
	}, nil
}

// getUserInfo
func (s *aiConversationService) getUserInfo(ctx context.Context, userID string) (schema.AIConversationUserInfo, error) {
	userInfo := schema.AIConversationUserInfo{}

	user, exist, err := s.userCommon.GetUserBasicInfoByID(ctx, userID)
	if err != nil {
		return userInfo, err
	}
	if !exist {
		return userInfo, errors.BadRequest(reason.ObjectNotFound)
	}

	userInfo.ID = user.ID
	userInfo.Username = user.Username
	userInfo.DisplayName = user.DisplayName
	userInfo.Avatar = user.Avatar
	userInfo.Rank = user.Rank
	return userInfo, nil
}

// DeleteConversationForAdmin
func (s *aiConversationService) DeleteConversationForAdmin(ctx context.Context, req *schema.AIConversationAdminDeleteReq) error {
	_, exist, err := s.aiConversationRepo.GetConversation(ctx, req.ConversationID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err)
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}

	if err := s.aiConversationRepo.DeleteConversation(ctx, req.ConversationID); err != nil {
		return err
	}

	return nil
}
