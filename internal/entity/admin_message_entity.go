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

// AdminMessage a message written by an admin / moderator to one user, delivered as a notification
type AdminMessage struct {
	ID             string     `xorm:"not null pk BIGINT(20) id"` // unique id (type 11) — a plain autoincrement would be mangled by uid.DeShortID in the notification path
	CreatedAt      time.Time  `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP index created_at"`
	SenderUserID   string     `xorm:"not null default 0 BIGINT(20) index sender_user_id"`
	ReceiverUserID string     `xorm:"not null default 0 BIGINT(20) index receiver_user_id"`
	Title          string     `xorm:"not null default '' VARCHAR(200) title"`
	Body           string     `xorm:"TEXT body"`
	ObjectType     string     `xorm:"not null default '' VARCHAR(32) object_type"`
	ObjectID       string     `xorm:"not null default 0 BIGINT(20) object_id"`
	QuestionID     string     `xorm:"not null default 0 BIGINT(20) question_id"`
	AnswerID       string     `xorm:"not null default 0 BIGINT(20) answer_id"`
	ReadAt         *time.Time `xorm:"TIMESTAMP read_at"`
}

// TableName admin_message table name
func (AdminMessage) TableName() string {
	return "admin_message"
}
