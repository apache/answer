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

// AIConversationRecord AI Conversation Record
type AIConversationRecord struct {
	ID               int       `xorm:"not null pk autoincr INT(11) id"`
	CreatedAt        time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt        time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
	ConversationID   string    `xorm:"not null VARCHAR(255) conversation_id"`
	ChatCompletionID string    `xorm:"not null VARCHAR(255) chat_completion_id"`
	Role             string    `xorm:"not null default '' VARCHAR(128) role"`
	Content          string    `xorm:"not null MEDIUMTEXT content"`
	Helpful          int       `xorm:"not null default 0 INT(11) helpful"`
	Unhelpful        int       `xorm:"not null default 0 INT(11) unhelpful"`
}

// TableName returns the table name
func (AIConversationRecord) TableName() string {
	return "ai_conversation_record"
}
