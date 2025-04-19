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

import (
	"strings"

	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/pkg/converter"
)

// SearchTagLikeReq get tag list all request
type SearchTagLikeReq struct {
	// tag
	Tag     string `validate:"omitempty" form:"tag"`
	IsAdmin bool   `json:"-"`
}

// SearchTagsBySlugName search tags by slug name
type SearchTagsBySlugName struct {
	// slug name list split by ','
	Tags string `form:"tags"`
}

// GetTagInfoReq get tag info request
type GetTagInfoReq struct {
	// tag id
	ID string `validate:"omitempty" form:"id"`
	// tag slug name
	Name       string `validate:"omitempty,gt=0,lte=35" form:"name"`
	UserID     string `json:"-"`
	CanEdit    bool   `json:"-"`
	CanDelete  bool   `json:"-"`
	CanMerge   bool   `json:"-"`
	CanRecover bool   `json:"-"`
}

type GetTamplateTagInfoReq struct {
	// tag id
	ID string `validate:"omitempty" form:"id"`
	// tag slug name
	Name string `validate:"omitempty" form:"name"`
	// user id
	UserID   string `json:"-"`
	Page     int    `validate:"omitempty,min=1" form:"page"`
	PageSize int    `validate:"omitempty,min=1" form:"page_size"`
}

func (r *GetTagInfoReq) Check() (errFields []*validator.FormErrorField, err error) {
	r.Name = strings.ToLower(r.Name)
	return nil, nil
}

// GetTagResp get tag response
type GetTagResp struct {
	TagID         string                    `json:"tag_id"`
	CreatedAt     int64                     `json:"created_at"`
	UpdatedAt     int64                     `json:"updated_at"`
	SlugName      string                    `json:"slug_name"`
	DisplayName   string                    `json:"display_name"`
	Excerpt       string                    `json:"excerpt"`
	OriginalText  string                    `json:"original_text"`
	ParsedText    string                    `json:"parsed_text"`
	Description   string                    `json:"description"`
	FollowCount   int                       `json:"follow_count"`
	QuestionCount int                       `json:"question_count"`
	IsFollower    bool                      `json:"is_follower"`
	Status        string                    `json:"status"`
	MemberActions []*PermissionMemberAction `json:"member_actions"`
	// if main tag slug name is not empty, this tag is synonymous with the main tag
	MainTagSlugName string `json:"main_tag_slug_name"`
	Recommend       bool   `json:"recommend"`
	Reserved        bool   `json:"reserved"`
}

func (tr *GetTagResp) GetExcerpt() {
	excerpt := strings.TrimSpace(tr.ParsedText)
	idx := strings.Index(excerpt, "\n")
	if idx >= 0 {
		excerpt = excerpt[0:idx]
	}
	tr.Excerpt = excerpt
}

// GetTagPageResp get tag response
type GetTagPageResp struct {
	// tag_id
	TagID string `json:"tag_id"`
	// slug_name
	SlugName string `json:"slug_name"`
	// display_name
	DisplayName string `json:"display_name"`
	// excerpt
	Excerpt string `json:"excerpt"`
	//description
	Description string `json:"description"`
	// original text
	OriginalText string `json:"original_text"`
	// parsed_text
	ParsedText string `json:"parsed_text"`
	// follower amount
	FollowCount int `json:"follow_count"`
	// question amount
	QuestionCount int `json:"question_count"`
	// is follower
	IsFollower bool `json:"is_follower"`
	// created time
	CreatedAt int64 `json:"created_at"`
	// updated time
	UpdatedAt int64 `json:"updated_at"`
	Recommend bool  `json:"recommend"`
	Reserved  bool  `json:"reserved"`
}

func (tr *GetTagPageResp) GetExcerpt() {
	excerpt := strings.TrimSpace(tr.ParsedText)
	idx := strings.Index(excerpt, "\n")
	if idx >= 0 {
		excerpt = excerpt[0:idx]
	}
	tr.Excerpt = excerpt
}

type TagChange struct {
	ObjectID string     `json:"object_id"` // object_id
	Tags     []*TagItem `json:"tags"`      // tags name
	// user id
	UserID string `json:"-"`
}

type TagItem struct {
	// slug_name
	SlugName string `validate:"omitempty,gt=0,lte=35" json:"slug_name"`
	// display_name
	DisplayName string `validate:"omitempty,gt=0,lte=35" json:"display_name"`
	// original text
	OriginalText string `validate:"omitempty" json:"original_text"`
	// parsed text
	ParsedText string `json:"-"`
}

// RemoveTagReq delete tag request
type RemoveTagReq struct {
	// tag_id
	TagID string `validate:"required" json:"tag_id"`
	// user id
	UserID string `json:"-"`
}

// AddTagReq add tag request
type AddTagReq struct {
	// slug_name
	SlugName string `validate:"required,gt=0,lte=35" json:"slug_name"`
	// display_name
	DisplayName string `validate:"required,gt=0,lte=35" json:"display_name"`
	// original text
	OriginalText string `validate:"required,gt=0,lte=65536" json:"original_text"`
	// parsed text
	ParsedText string `json:"-"`
	// user id
	UserID string `json:"-"`
}

func (req *AddTagReq) Check() (errFields []*validator.FormErrorField, err error) {
	req.ParsedText = converter.Markdown2HTML(req.OriginalText)
	req.SlugName = strings.ToLower(req.SlugName)
	return nil, nil
}

// AddTagResp add tag response
type AddTagResp struct {
	SlugName string `json:"slug_name"`
}

// UpdateTagReq update tag request
type UpdateTagReq struct {
	// tag_id
	TagID string `validate:"required" json:"tag_id"`
	// slug_name
	SlugName string `validate:"omitempty,gt=0,lte=35" json:"slug_name"`
	// display_name
	DisplayName string `validate:"omitempty,gt=0,lte=35" json:"display_name"`
	// original text
	OriginalText string `validate:"omitempty" json:"original_text"`
	// parsed text
	ParsedText string `json:"-"`
	// edit summary
	EditSummary string `validate:"omitempty" json:"edit_summary"`
	// user id
	UserID       string `json:"-"`
	NoNeedReview bool   `json:"-"`
}

func (r *UpdateTagReq) Check() (errFields []*validator.FormErrorField, err error) {
	r.ParsedText = converter.Markdown2HTML(r.OriginalText)
	return nil, nil
}

// RecoverTagReq update tag request
type RecoverTagReq struct {
	TagID  string `validate:"required" json:"tag_id"`
	UserID string `json:"-"`
}

// UpdateTagResp update tag response
type UpdateTagResp struct {
	WaitForReview bool `json:"wait_for_review"`
}

// GetTagWithPageReq get tag list page request
type GetTagWithPageReq struct {
	// page
	Page int `validate:"omitempty,min=1" form:"page"`
	// page size
	PageSize int `validate:"omitempty,min=1" form:"page_size"`
	// slug_name
	SlugName string `validate:"omitempty,gt=0,lte=35" form:"slug_name"`
	// display_name
	DisplayName string `validate:"omitempty,gt=0,lte=35" form:"display_name"`
	// query condition
	QueryCond string `validate:"omitempty,oneof=popular name newest" form:"query_cond"`
	// user id
	UserID string `json:"-"`
}

// GetTagSynonymsReq get tag synonyms request
type GetTagSynonymsReq struct {
	// tag_id
	TagID string `validate:"required" form:"tag_id"`
	// user id
	UserID string `json:"-"`
	// whether user can edit it
	CanEdit bool `json:"-"`
}

// GetTagSynonymsResp get tag synonyms response
type GetTagSynonymsResp struct {
	// synonyms
	Synonyms []*TagSynonym `json:"synonyms"`
	// MemberActions
	MemberActions []*PermissionMemberAction `json:"member_actions"`
}

type TagSynonym struct {
	// tag id
	TagID string `json:"tag_id"`
	// slug name
	SlugName string `json:"slug_name"`
	// display name
	DisplayName string `json:"display_name"`
	// if main tag slug name is not empty, this tag is synonymous with the main tag
	MainTagSlugName string `json:"main_tag_slug_name"`
}

// UpdateTagSynonymReq update tag request
type UpdateTagSynonymReq struct {
	// tag_id
	TagID string `validate:"required" json:"tag_id"`
	// synonym tag list
	SynonymTagList []*TagItem `validate:"required,dive" json:"synonym_tag_list"`
	// user id
	UserID string `json:"-"`
}

func (req *UpdateTagSynonymReq) Format() {
	for _, item := range req.SynonymTagList {
		item.SlugName = strings.ToLower(item.SlugName)
	}
}

// GetFollowingTagsResp get following tags response
type GetFollowingTagsResp struct {
	// tag id
	TagID string `json:"tag_id"`
	// slug name
	SlugName string `json:"slug_name"`
	// display name
	DisplayName string `json:"display_name"`
	// if main tag slug name is not empty, this tag is synonymous with the main tag
	MainTagSlugName string `json:"main_tag_slug_name"`
	Recommend       bool   `json:"recommend"`
	Reserved        bool   `json:"reserved"`
}

// GetTagBasicResp get tag basic response
type GetTagBasicResp struct {
	TagID       string `json:"tag_id"`
	SlugName    string `json:"slug_name"`
	DisplayName string `json:"display_name"`
	Recommend   bool   `json:"recommend"`
	Reserved    bool   `json:"reserved"`
}

// MergeTagReq merge tag request
type MergeTagReq struct {
	// source tag id
	SourceTagID string `validate:"required" json:"source_tag_id"`
	// target tag id
	TargetTagID string `validate:"required" json:"target_tag_id"`
	// user id
	UserID string `json:"-"`
}

// MergeTagResp merge tag response
type MergeTagResp struct {
}
