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
	FileRecordStatusAvailable = 1
	FileRecordStatusDeleted   = 10
)

// FileRecord file record
type FileRecord struct {
	ID        int       `xorm:"not null pk autoincr INT(10) id"`
	CreatedAt time.Time `xorm:"not null default CURRENT_TIMESTAMP created TIMESTAMP created_at"`
	UpdatedAt time.Time `xorm:"not null default CURRENT_TIMESTAMP updated TIMESTAMP updated_at"`
	UserID    string    `xorm:"not null default 0 BIGINT(20) user_id"`
	FilePath  string    `xorm:"not null VARCHAR(256) file_path"`
	FileURL   string    `xorm:"not null VARCHAR(1024) file_url"`
	ObjectID  string    `xorm:"not null default 0 INDEX BIGINT(20) object_id"`
	Source    string    `xorm:"not null VARCHAR(128) source"`
	Status    int       `xorm:"not null default 0 TINYINT(4) status"`
}

// TableName file record table name
func (FileRecord) TableName() string {
	return "file_record"
}
