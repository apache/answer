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

import (
	"time"
)

const (
	// RevisionNormalStatus this revision is normal
	RevisionNormalStatus = 0
	// RevisionUnreviewedStatus this revision is unreviewed
	RevisionUnreviewedStatus = 1
	// RevisionReviewPassStatus this revision is reviewed and approved by operator
	RevisionReviewPassStatus = 2
	// RevisionReviewRejectStatus this revision is reviewed and rejected by operator
	RevisionReviewRejectStatus = 3
)

// Revision revision
type Revision struct {
	ID           string    `xorm:"not null pk autoincr BIGINT(20) id"`
	CreatedAt    time.Time `xorm:"created TIMESTAMP created_at"`
	UpdatedAt    time.Time `xorm:"updated TIMESTAMP updated_at"`
	UserID       string    `xorm:"not null default 0 BIGINT(20) user_id"`
	ObjectType   int       `xorm:"not null default 0 INT(11) object_type"`
	ObjectID     string    `xorm:"not null default 0 BIGINT(20) INDEX object_id"`
	Title        string    `xorm:"not null default '' VARCHAR(255) title"`
	Content      string    `xorm:"not null MEDIUMTEXT content"`
	Log          string    `xorm:"VARCHAR(255) log"`
	Status       int       `xorm:"not null default 1 INT(11) status"`
	ReviewUserID int64     `xorm:"not null default 0 BIGINT(20) review_user_id"`
}

// TableName revision table name
func (Revision) TableName() string {
	return "revision"
}
