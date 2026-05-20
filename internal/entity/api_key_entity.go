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

// APIKey entity
type APIKey struct {
	ID          int       `xorm:"not null pk autoincr INT(11) id"`
	CreatedAt   time.Time `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt   time.Time `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
	LastUsedAt  time.Time `xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP last_used_at"`
	Description string    `xorm:"not null MEDIUMTEXT description"`
	AccessKey   string    `xorm:"not null unique VARCHAR(255) access_key"`
	Scope       string    `xorm:"not null VARCHAR(255) scope"`
	UserID      string    `xorm:"not null default 0 BIGINT(20) user_id"`
	Hidden      int       `xorm:"not null default 0 INT(11) hidden"`
}

// TableName category table name
func (c *APIKey) TableName() string {
	return "api_key"
}
