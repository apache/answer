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
	"regexp"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/validator"
	"github.com/apache/answer/plugin"
)

type SearchDTO struct {
	Query       string `validate:"required,gte=1,lte=60" form:"q"`
	Page        int    `validate:"omitempty,min=1" form:"page,default=1"`
	Size        int    `validate:"omitempty,min=1,max=50" form:"size,default=30"`
	Order       string `validate:"required,oneof=newest active score relevance" form:"order,default=relevance" enums:"newest,active,score,relevance"`
	CaptchaID   string `form:"captcha_id"`
	CaptchaCode string `form:"captcha_code"`
	UserID      string `json:"-"`
}

func (s *SearchDTO) Check() (errField []*validator.FormErrorField, err error) {
	// Replace special characters.
	// Special characters will cause the search abnormal, such as search for "#" will get nearly all the content that Markdown format.
	replacedContent, patterns := ReplaceSearchContent(s.Query)
	s.Query = strings.Join(append(patterns, replacedContent), " ")

	return nil, nil
}

func ReplaceSearchContent(content string) (string, []string) {
	// Define the regular expressions for key:value pairs and [tag]
	keyValueRegex := regexp.MustCompile(`\w+:\S+`)
	tagRegex := regexp.MustCompile(`\[\w+\]`)
	// Define the pattern for characters to replace
	replaceCharsPattern := regexp.MustCompile(`[+#.<>\-_()*]`)

	// Extract key:value pairs
	keyValues := keyValueRegex.FindAllString(content, -1)
	// Extract [tag]
	tags := tagRegex.FindAllString(content, -1)

	// Replace key:value pairs and [tag] with empty string
	contentWithoutPatterns := keyValueRegex.ReplaceAllString(content, "")
	contentWithoutPatterns = tagRegex.ReplaceAllString(contentWithoutPatterns, "")

	// Replace characters with pattern [+#.<>_()*] with space
	replacedContent := replaceCharsPattern.ReplaceAllString(contentWithoutPatterns, " ")

	return strings.TrimSpace(replacedContent), append(keyValues, tags...)
}

type SearchCondition struct {
	// search target type: all/question/answer
	TargetType string
	// search query user id
	UserID string
	// vote amount
	VoteAmount int
	// only show not accepted answer's question
	NotAccepted bool
	// view amount
	Views int
	// answer count
	AnswerAmount int
	// only show accepted answer
	Accepted bool
	// only show this question's answer
	QuestionID string
	// search query tags
	Tags [][]string
	// search query keywords
	Words []string
}

// SearchAll check if search all
func (s *SearchCondition) SearchAll() bool {
	return len(s.TargetType) == 0
}

// SearchQuestion check if search only need question
func (s *SearchCondition) SearchQuestion() bool {
	return s.TargetType == constant.QuestionObjectType
}

// SearchAnswer check if search only need answer
func (s *SearchCondition) SearchAnswer() bool {
	return s.TargetType == constant.AnswerObjectType
}

// Convert2PluginSearchCond convert to plugin search condition
func (s *SearchCondition) Convert2PluginSearchCond(page, pageSize int, order string) *plugin.SearchBasicCond {
	basic := &plugin.SearchBasicCond{
		Page:         page,
		PageSize:     pageSize,
		Words:        s.Words,
		TagIDs:       s.Tags,
		UserID:       s.UserID,
		Order:        plugin.SearchOrderCond(order),
		QuestionID:   s.QuestionID,
		VoteAmount:   s.VoteAmount,
		ViewAmount:   s.Views,
		AnswerAmount: s.AnswerAmount,
	}
	if s.Accepted {
		basic.AnswerAccepted = plugin.AcceptedCondTrue
	} else {
		basic.AnswerAccepted = plugin.AcceptedCondAll
	}
	if s.NotAccepted {
		basic.QuestionAccepted = plugin.AcceptedCondFalse
	} else {
		basic.QuestionAccepted = plugin.AcceptedCondAll
	}
	return basic
}

type SearchObject struct {
	ID              string `json:"id"`
	QuestionID      string `json:"question_id"`
	Title           string `json:"title"`
	UrlTitle        string `json:"url_title"`
	Excerpt         string `json:"excerpt"`
	CreatedAtParsed int64  `json:"created_at"`
	VoteCount       int    `json:"vote_count"`
	Accepted        bool   `json:"accepted"`
	AnswerCount     int    `json:"answer_count"`
	// user info
	UserInfo *SearchObjectUser `json:"user_info"`
	// tags
	Tags []*TagResp `json:"tags"`
	// Status
	StatusStr string `json:"status"`
}

type SearchObjectUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Rank        int    `json:"rank"`
	Status      string `json:"status"`
}

type TagResp struct {
	ID          string `json:"-"`
	SlugName    string `json:"slug_name"`
	DisplayName string `json:"display_name"`
	// if main tag slug name is not empty, this tag is synonymous with the main tag
	MainTagSlugName string `json:"main_tag_slug_name"`
	Recommend       bool   `json:"recommend"`
	Reserved        bool   `json:"reserved"`
}

type SearchResult struct {
	// object_type
	ObjectType string `json:"object_type"`
	// this object
	Object *SearchObject `json:"object"`
}

type SearchResp struct {
	Total int64 `json:"count"`
	// search response
	SearchResults []*SearchResult `json:"list"`
}

type SearchDescResp struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
	Link string `json:"link"`
}
