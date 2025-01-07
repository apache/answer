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

package schema

import "github.com/apache/answer/internal/entity"

const (
	BadgeStatusActive   BadgeStatus = "active"
	BadgeStatusInactive BadgeStatus = "inactive"
)

type BadgeStatus string

var BadgeStatusMap = map[int8]BadgeStatus{
	entity.BadgeStatusActive:   BadgeStatusActive,
	entity.BadgeStatusInactive: BadgeStatusInactive,
}

var BadgeStatusEMap = map[BadgeStatus]int8{
	BadgeStatusActive:   entity.BadgeStatusActive,
	BadgeStatusInactive: entity.BadgeStatusInactive,
}

// BadgeListInfo get badge list response
type BadgeListInfo struct {
	// badge id
	ID string `json:"id" `
	// badge name
	Name string `json:"name" `
	// badge icon
	Icon string `json:"icon" `
	// badge award count
	AwardCount int `json:"award_count" `
	// badge earned count
	EarnedCount int64 `json:"earned_count" `
	// badge level
	Level entity.BadgeLevel `json:"level" `
}

type GetBadgeListResp struct {
	// badge list info
	Badges []*BadgeListInfo `json:"badges" `
	// badge group name
	GroupName string `json:"group_name" `
}

type UpdateBadgeStatusReq struct {
	// badge id
	ID string `validate:"required" json:"id"`
	// badge status
	Status BadgeStatus `validate:"required" json:"status"`
}

type GetBadgeListPagedReq struct {
	// page
	Page int `validate:"omitempty,min=1" form:"page"`
	// page size
	PageSize int `validate:"omitempty,min=1" form:"page_size"`
	// badge status
	Status BadgeStatus `validate:"omitempty" form:"status"`
	// query condition
	Query string `validate:"omitempty" form:"q"`
}

type GetBadgeListPagedResp struct {
	// badge id
	ID string `json:"id" `
	// badge name
	Name string `json:"name" `
	// badge description
	Description string `json:"description" `
	// badge icon
	Icon string `json:"icon" `
	// badge award count
	AwardCount int `json:"award_count" `
	// badge earned count
	Earned bool `json:"earned" `
	// badge level
	Level entity.BadgeLevel `json:"level" `
	// badge group name
	GroupName string `json:"group_name" `
	// badge status
	Status BadgeStatus `json:"status"`
}

type GetBadgeInfoResp struct {
	// badge id
	ID string `json:"id" `
	// badge name
	Name string `json:"name" `
	// badge description
	Description string `json:"description" `
	// badge icon
	Icon string `json:"icon" `
	// badge award count
	AwardCount int `json:"award_count" `
	// badge earned count
	EarnedCount int64 `json:"earned_count" `
	// badge is single or multiple
	IsSingle bool `json:"is_single" `
	// badge level
	Level entity.BadgeLevel `json:"level" `
}

type GetBadgeAwardWithPageReq struct {
	// page
	Page int `validate:"omitempty,min=1" form:"page"`
	// page size
	PageSize int `validate:"omitempty,min=1" form:"page_size"`
	// badge id
	BadgeID string `validate:"required" form:"badge_id"`
	// username
	Username string `validate:"omitempty,gt=0,lte=100" form:"username"`
	// user id
	UserID string `json:"-"`
}

type GetBadgeAwardWithPageResp struct {
	// created time
	CreatedAt int64 `json:"created_at"`
	// object id
	ObjectID string `json:"object_id"`
	// question id
	QuestionID string `json:"question_id"`
	// answer id
	AnswerID string `json:"answer_id"`
	// comment id
	CommentID string `json:"comment_id"`
	// object type
	ObjectType string `json:"object_type" enums:"question,answer,comment"`
	// url title
	UrlTitle string `json:"url_title"`
	// author user info
	AuthorUserInfo UserBasicInfo `json:"author_user_info"`
}

type GetUserBadgeAwardListReq struct {
	// username
	Username string `validate:"required,gt=0,lte=100" form:"username"`
	// user id
	UserID string `json:"-"`
	Limit  int    `json:"-"`
}

type GetUserBadgeAwardListResp struct {
	// badge id
	ID string `json:"id" `
	// badge name
	Name string `json:"name" `
	// badge icon
	Icon string `json:"icon" `
	// badge award count
	EarnedCount int64 `json:"earned_count" `
	// badge level
	Level entity.BadgeLevel `json:"level" `
}

type BadgeTplData struct {
	ProfileURL string
}
