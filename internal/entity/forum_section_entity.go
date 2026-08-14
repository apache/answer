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

const ForumSectionStatusAvailable = 1

// ForumSection is a fixed campus forum category. Only leaf sections accept posts.
type ForumSection struct {
	ID        int64     `xorm:"not null pk BIGINT(20) id"`
	CreatedAt time.Time `xorm:"created TIMESTAMP created_at"`
	UpdatedAt time.Time `xorm:"updated TIMESTAMP updated_at"`
	ParentID  int64     `xorm:"not null default 0 BIGINT(20) INDEX parent_id"`
	Slug      string    `xorm:"not null unique VARCHAR(50) slug"`
	Name      string    `xorm:"not null VARCHAR(50) name"`
	Sort      int       `xorm:"not null default 0 INT(11) sort"`
	AdminOnly bool      `xorm:"not null default false BOOL admin_only"`
	Status    int       `xorm:"not null default 1 INT(11) status"`
}

func (ForumSection) TableName() string {
	return "forum_section"
}
