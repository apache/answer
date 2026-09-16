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

// ActivityLog one row per user action for the admin "community activity log"
// (logins, page views, posts, votes, badges, review queue, reputation changes, ...).
type ActivityLog struct {
	ID           int64     `xorm:"not null pk autoincr BIGINT(20) id"`
	CreatedAt    time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP index created_at"`
	UserID       string    `xorm:"not null default 0 BIGINT(20) index user_id"`
	Action       string    `xorm:"not null default '' VARCHAR(64) index action"`
	ObjectType   string    `xorm:"not null default '' VARCHAR(32) object_type"`
	ObjectID     string    `xorm:"not null default 0 BIGINT(20) index object_id"`
	QuestionID   string    `xorm:"not null default 0 BIGINT(20) question_id"`
	AnswerID     string    `xorm:"not null default 0 BIGINT(20) answer_id"`
	TargetUserID string    `xorm:"not null default 0 BIGINT(20) index target_user_id"`
	RankDelta    int       `xorm:"not null default 0 INT(11) rank_delta"`
	Detail       string    `xorm:"TEXT detail"`
	IP           string    `xorm:"not null default '' VARCHAR(64) ip"`
	UserAgent    string    `xorm:"not null default '' VARCHAR(512) user_agent"`
	SourceRef    string    `xorm:"not null default '' VARCHAR(96) index source_ref"`
}

// TableName activity_log table name
func (ActivityLog) TableName() string {
	return "activity_log"
}
