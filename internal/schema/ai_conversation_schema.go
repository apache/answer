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

package schema

import (
	"github.com/apache/answer/internal/base/validator"
)

// AIConversationListReq ai conversation list req
type AIConversationListReq struct {
	Page     int    `validate:"omitempty,min=1" form:"page"`
	PageSize int    `validate:"omitempty,min=1" form:"page_size"`
	UserID   string `validate:"omitempty" json:"-"`
}

// AIConversationListItem ai conversation list item
type AIConversationListItem struct {
	ConversationID string `json:"conversation_id"`
	Topic          string `json:"topic"`
	CreatedAt      int64  `json:"created_at"`
}

// AIConversationDetailReq ai conversation detail req
type AIConversationDetailReq struct {
	ConversationID string `validate:"required" form:"conversation_id" json:"conversation_id"`
	UserID         string `validate:"omitempty" json:"-"`
}

// AIConversationRecord ai conversation record
type AIConversationRecord struct {
	ChatCompletionID string `json:"chat_completion_id"`
	Role             string `json:"role"`
	Content          string `json:"content"`
	Helpful          int    `json:"helpful"`
	Unhelpful        int    `json:"unhelpful"`
	CreatedAt        int64  `json:"created_at"`
}

// AIConversationDetailResp ai conversation detail resp
type AIConversationDetailResp struct {
	ConversationID string                  `json:"conversation_id"`
	Topic          string                  `json:"topic"`
	Records        []*AIConversationRecord `json:"records"`
	CreatedAt      int64                   `json:"created_at"`
	UpdatedAt      int64                   `json:"updated_at"`
}

// AIConversationVoteReq ai conversation vote req
type AIConversationVoteReq struct {
	ChatCompletionID string `validate:"required" json:"chat_completion_id"`
	VoteType         string `validate:"required,oneof=helpful unhelpful" json:"vote_type"`
	Cancel           bool   `validate:"omitempty" json:"cancel"`
	UserID           string `validate:"omitempty" json:"-"`
}

// AIConversationAdminListReq ai conversation admin list req
type AIConversationAdminListReq struct {
	Page     int `validate:"omitempty,min=1" form:"page"`
	PageSize int `validate:"omitempty,min=1" form:"page_size"`
}

// AIConversationAdminListItem ai conversation admin list item
type AIConversationAdminListItem struct {
	ID             string                 `json:"id"`
	Topic          string                 `json:"topic"`
	UserInfo       AIConversationUserInfo `json:"user_info"`
	HelpfulCount   int64                  `json:"helpful_count"`
	UnhelpfulCount int64                  `json:"unhelpful_count"`
	CreatedAt      int64                  `json:"created_at"`
}

// AIConversationUserInfo ai conversation user info
type AIConversationUserInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Rank        int    `json:"rank"`
}

// AIConversationAdminDetailReq ai conversation admin detail req
type AIConversationAdminDetailReq struct {
	ConversationID string `validate:"required" form:"conversation_id" json:"conversation_id"`
}

// AIConversationAdminDetailResp ai conversation admin detail resp
type AIConversationAdminDetailResp struct {
	ConversationID string                 `json:"conversation_id"`
	Topic          string                 `json:"topic"`
	UserInfo       AIConversationUserInfo `json:"user_info"`
	Records        []AIConversationRecord `json:"records"`
	CreatedAt      int64                  `json:"created_at"`
}

// AIConversationAdminDeleteReq admin delete ai
type AIConversationAdminDeleteReq struct {
	ConversationID string `validate:"required" json:"conversation_id"`
}

func (req *AIConversationDetailReq) Check() (errFields []*validator.FormErrorField, err error) {
	return nil, nil
}

func (req *AIConversationVoteReq) Check() (errFields []*validator.FormErrorField, err error) {
	return nil, nil
}
