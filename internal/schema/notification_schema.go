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

const (
	NotificationTypeInbox        = 1
	NotificationTypeAchievement  = 2
	NotificationNotRead          = 1
	NotificationRead             = 2
	NotificationStatusNormal     = 1
	NotificationStatusDelete     = 10
	NotificationInboxTypeAll     = 0
	NotificationInboxTypePosts   = 1
	NotificationInboxTypeVotes   = 2
	NotificationInboxTypeInvites = 3
)

var NotificationType = map[string]int{
	"inbox":       NotificationTypeInbox,
	"achievement": NotificationTypeAchievement,
}

var NotificationInboxType = map[string]int{
	"all":     NotificationInboxTypeAll,
	"posts":   NotificationInboxTypePosts,
	"invites": NotificationInboxTypeInvites,
	"votes":   NotificationInboxTypeVotes,
}

type NotificationContent struct {
	ID                 string         `json:"id"`
	TriggerUserID      string         `json:"-"` //show userid
	ReceiverUserID     string         `json:"-"` // receiver userid
	UserInfo           *UserBasicInfo `json:"user_info,omitempty"`
	ObjectInfo         ObjectInfo     `json:"object_info"`
	Rank               int            `json:"rank"`
	NotificationAction string         `json:"notification_action,omitempty"`
	Type               int            `json:"-"` //	1 inbox 2 achievement
	IsRead             bool           `json:"is_read"`
	UpdateTime         int64          `json:"update_time"`
}

type GetRedDot struct {
	CanReviewQuestion bool   `json:"-"`
	CanReviewAnswer   bool   `json:"-"`
	CanReviewTag      bool   `json:"-"`
	UserID            string `json:"-"`
	IsAdmin           bool   `json:"-"`
}

// NotificationMsg notification message
type NotificationMsg struct {
	// trigger notification user id
	TriggerUserID string
	// receive notification user id
	ReceiverUserID string
	// type 1 inbox 2 achievement
	Type int
	// notification title
	Title string
	// notification object
	ObjectID string
	// notification object type
	ObjectType string
	// notification action
	NotificationAction string
	// if true no need to send notification to all followers
	NoNeedPushAllFollow bool
	// extra info
	ExtraInfo map[string]string
}

type ObjectInfo struct {
	Title      string            `json:"title"`
	ObjectID   string            `json:"object_id"`
	ObjectMap  map[string]string `json:"object_map"`
	ObjectType string            `json:"object_type"`
}

type RedDot struct {
	Inbox       int64 `json:"inbox"`
	Achievement int64 `json:"achievement"`
	Revision    int64 `json:"revision"`
	CanRevision bool  `json:"can_revision"`
}

type NotificationSearch struct {
	Page         int    `json:"page" form:"page"`           //Query number of pages
	PageSize     int    `json:"page_size" form:"page_size"` //Search page size
	Type         int    `json:"-" form:"-"`
	TypeStr      string `json:"type" form:"type"`             // inbox achievement
	InboxTypeStr string `json:"inbox_type" form:"inbox_type"` // inbox achievement
	InboxType    int    `json:"-" form:"-"`                   // inbox achievement
	UserID       string `json:"-"`
}

type NotificationClearRequest struct {
	UserID            string `json:"-"`
	TypeStr           string `json:"type" form:"type"` // inbox achievement
	CanReviewQuestion bool   `json:"-"`
	CanReviewAnswer   bool   `json:"-"`
	CanReviewTag      bool   `json:"-"`
}

type NotificationClearIDRequest struct {
	UserID string `json:"-"`
	ID     string `json:"id" form:"id"`
}
