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

	"github.com/tidwall/gjson"
)

type BadgeLevel int

const (
	BadgeStatusActive   = 1
	BadgeStatusDeleted  = 10
	BadgeStatusInactive = 11

	BadgeLevelBronze BadgeLevel = 1
	BadgeLevelSilver BadgeLevel = 2
	BadgeLevelGold   BadgeLevel = 3

	BadgeSingleAward = 1
	BadgeMultiAward  = 2
)

// Badge badge
type Badge struct {
	ID           string     `xorm:"not null pk BIGINT(20) id"`
	CreatedAt    time.Time  `xorm:"created not null default CURRENT_TIMESTAMP TIMESTAMP created_at"`
	UpdatedAt    time.Time  `xorm:"updated not null default CURRENT_TIMESTAMP TIMESTAMP updated_at"`
	Name         string     `xorm:"not null default '' VARCHAR(256) name"`
	Icon         string     `xorm:"not null default '' VARCHAR(1024) icon"`
	AwardCount   int        `xorm:"not null default 0 INT(11) award_count"`
	Description  string     `xorm:"not null MEDIUMTEXT description"`
	Status       int8       `xorm:"not null default 1 INT(11) status"`
	BadgeGroupID int64      `xorm:"not null default 0 BIGINT(20) badge_group_id"`
	Level        BadgeLevel `xorm:"not null default 1 TINYINT(4) level"`
	Single       int8       `xorm:"not null default 1 TINYINT(4) single"`
	Collect      string     `xorm:"not null default '' VARCHAR(128) collect"`
	Handler      string     `xorm:"not null default '' VARCHAR(128) handler"`
	Param        string     `xorm:"not null TEXT param"`
}

// TableName badge table name
func (b *Badge) TableName() string {
	return "badge"
}

func (b *Badge) GetIntParam(key string) int64 {
	return gjson.Get(b.Param, key).Int()
}

func (b *Badge) GetStringParam(key string) string {
	return gjson.Get(b.Param, key).String()
}
