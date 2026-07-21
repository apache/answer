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

const (
	TagStatusAvailable = 1
	TagStatusDeleted   = 10
)

var TagStatusDisplayMapping = map[int]string{
	TagStatusAvailable: "available",
	TagStatusDeleted:   "deleted",
}

// Tag tag
type Tag struct {
	ID              string    `xorm:"not null pk comment('tag_id') BIGINT(20) id"`
	CreatedAt       time.Time `xorm:"created TIMESTAMP created_at"`
	UpdatedAt       time.Time `xorm:"updated TIMESTAMP updated_at"`
	MainTagID       int64     `xorm:"not null default 0 BIGINT(20) main_tag_id"`
	MainTagSlugName string    `xorm:"not null default '' VARCHAR(35) main_tag_slug_name"`
	SlugName        string    `xorm:"not null default '' unique VARCHAR(35) slug_name"`
	Domain          string    `xorm:"not null default '' VARCHAR(253) domain index"`
	DisplayName     string    `xorm:"not null default '' VARCHAR(35) display_name"`
	OriginalText    string    `xorm:"not null MEDIUMTEXT original_text"`
	ParsedText      string    `xorm:"not null MEDIUMTEXT parsed_text"`
	FollowCount     int       `xorm:"not null default 0 INT(11) follow_count"`
	QuestionCount   int       `xorm:"not null default 0 INT(11) question_count"`
	Status          int       `xorm:"not null default 1 INT(11) status"`
	Recommend       bool      `xorm:"not null default false BOOL recommend"`
	Reserved        bool      `xorm:"not null default false BOOL reserved"`
	RevisionID      string    `xorm:"not null default 0 BIGINT(20) revision_id"`
	UserID          string    `xorm:"not null default 0 BIGINT(20) user_id"`
}

// TableName tag table name
func (Tag) TableName() string {
	return "tag"
}
