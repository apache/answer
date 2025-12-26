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
	UserStatusAvailable = 1
	UserStatusSuspended = 9
	UserStatusDeleted   = 10
)

const (
	EmailStatusAvailable    = 1
	EmailStatusToBeVerified = 2
)

const (
	UserAdminFlag = 1
)

// PermanentSuspensionTime is a fixed time representing permanent suspension (2099-12-31 23:59:59)
var PermanentSuspensionTime = time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)

// User user
type User struct {
	ID             string    `xorm:"not null pk autoincr BIGINT(20) id"`
	CreatedAt      time.Time `xorm:"created TIMESTAMP created_at"`
	UpdatedAt      time.Time `xorm:"updated TIMESTAMP updated_at"`
	SuspendedAt    time.Time `xorm:"TIMESTAMP suspended_at"`
	SuspendedUntil time.Time `xorm:"DATETIME suspended_until"`
	DeletedAt      time.Time `xorm:"TIMESTAMP deleted_at"`
	LastLoginDate  time.Time `xorm:"TIMESTAMP last_login_date"`
	Username       string    `xorm:"not null default '' VARCHAR(50) UNIQUE username"`
	Pass           string    `xorm:"not null default '' VARCHAR(255) pass"`
	EMail          string    `xorm:"not null VARCHAR(100) e_mail"`
	MailStatus     int       `xorm:"not null default 2 TINYINT(4) mail_status"`
	NoticeStatus   int       `xorm:"not null default 2 INT(11) notice_status"`
	FollowCount    int       `xorm:"not null default 0 INT(11) follow_count"`
	AnswerCount    int       `xorm:"not null default 0 INT(11) answer_count"`
	QuestionCount  int       `xorm:"not null default 0 INT(11) question_count"`
	Rank           int       `xorm:"not null default 0 INT(11) rank"`
	Status         int       `xorm:"not null default 1 INT(11) status"`
	AuthorityGroup int       `xorm:"not null default 1 INT(11) authority_group"`
	DisplayName    string    `xorm:"not null default '' VARCHAR(30) display_name"`
	Avatar         string    `xorm:"not null default '' VARCHAR(2048) avatar"`
	Mobile         string    `xorm:"not null VARCHAR(20) mobile"`
	Bio            string    `xorm:"not null TEXT bio"`
	BioHTML        string    `xorm:"not null TEXT bio_html"`
	Website        string    `xorm:"not null default '' VARCHAR(255) website"`
	Location       string    `xorm:"not null default '' VARCHAR(100) location"`
	IPInfo         string    `xorm:"not null default '' VARCHAR(255) ip_info"`
	IsAdmin        bool      `xorm:"not null default false BOOL is_admin"`
	Language       string    `xorm:"not null default '' VARCHAR(100) language"`
	ColorScheme    string    `xorm:"not null default '' VARCHAR(100) color_scheme"`
}

// TableName user table name
func (User) TableName() string {
	return "user"
}

type UserSearch struct {
	User
	Page     int `json:"page" form:"page"`           // Query number of pages
	PageSize int `json:"page_size" form:"page_size"` // Search page size
}
