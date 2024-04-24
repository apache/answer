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

import "github.com/apache/incubator-answer/internal/base/constant"

// ActivityMsg activity message
type ActivityMsg struct {
	UserID           string
	TriggerUserID    int64
	ObjectID         string
	OriginalObjectID string
	ActivityTypeKey  constant.ActivityTypeKey
	RevisionID       string
	ExtraInfo        map[string]string
}

// GetObjectTimelineReq get object timeline request
type GetObjectTimelineReq struct {
	ObjectID string `validate:"omitempty,gt=0,lte=100" form:"object_id"`
	ShowVote bool   `validate:"omitempty" form:"show_vote"`
	UserID   string `json:"-"`
	IsAdmin  bool   `json:"-"`
}

// GetObjectTimelineResp get object timeline response
type GetObjectTimelineResp struct {
	ObjectInfo *ActObjectInfo       `json:"object_info"`
	Timeline   []*ActObjectTimeline `json:"timeline"`
}

// ActObjectTimeline act object timeline
type ActObjectTimeline struct {
	ActivityID   string         `json:"activity_id"`
	RevisionID   string         `json:"revision_id"`
	CreatedAt    int64          `json:"created_at"`
	ActivityType string         `json:"activity_type"`
	Comment      string         `json:"comment"`
	ObjectID     string         `json:"object_id"`
	ObjectType   string         `json:"object_type"`
	Cancelled    bool           `json:"cancelled"`
	CancelledAt  int64          `json:"cancelled_at"`
	UserInfo     *UserBasicInfo `json:"user_info,omitempty"`
}

// ActObjectInfo act object info
type ActObjectInfo struct {
	Title           string `json:"title"`
	ObjectType      string `json:"object_type"`
	QuestionID      string `json:"question_id"`
	AnswerID        string `json:"answer_id"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	MainTagSlugName string `json:"main_tag_slug_name"`
}

// GetObjectTimelineDetailReq get object timeline detail request
type GetObjectTimelineDetailReq struct {
	NewRevisionID string `validate:"required,gt=0,lte=100" form:"new_revision_id"`
	OldRevisionID string `validate:"required,gt=0,lte=100" form:"old_revision_id"`
	UserID        string `json:"-"`
}

// GetObjectTimelineDetailResp get object timeline detail response
type GetObjectTimelineDetailResp struct {
	NewRevision *ObjectTimelineDetail `json:"new_revision"`
	OldRevision *ObjectTimelineDetail `json:"old_revision"`
}

// ObjectTimelineDetail object timeline detail
type ObjectTimelineDetail struct {
	Title           string               `json:"title"`
	Tags            []*ObjectTimelineTag `json:"tags"`
	OriginalText    string               `json:"original_text"`
	SlugName        string               `json:"slug_name"`
	MainTagSlugName string               `json:"main_tag_slug_name"`
}

// ObjectTimelineTag object timeline tags
type ObjectTimelineTag struct {
	SlugName        string `json:"slug_name"`
	DisplayName     string `json:"display_name"`
	MainTagSlugName string `json:"main_tag_slug_name"`
	Recommend       bool   `json:"recommend"`
	Reserved        bool   `json:"reserved"`
}

// PassReviewActivity pass review activity
type PassReviewActivity struct {
	UserID           string `json:"user_id"`
	TriggerUserID    string `json:"trigger_user_id"`
	ObjectID         string `json:"object_id"`
	OriginalObjectID string `json:"original_object_id"`
	RevisionID       string `json:"revision_id"`
}
