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

	"github.com/apache/answer/pkg/converter"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	MCPSearchCondKeyword       = "keyword"
	MCPSearchCondUsername      = "username"
	MCPSearchCondScore         = "score"
	MCPSearchCondTag           = "tag"
	MCPSearchCondPage          = "page"
	MCPSearchCondPageSize      = "page_size"
	MCPSearchCondTagName       = "tag_name"
	MCPSearchCondQuestionID    = "question_id"
	MCPSearchCondObjectID      = "object_id"
	MCPSearchCondSemanticQuery = "query"
	MCPSearchCondTopK          = "top_k"
)

type MCPSearchCond struct {
	Keyword    string   `json:"keyword"`
	Username   string   `json:"username"`
	Score      int      `json:"score"`
	Tags       []string `json:"tags"`
	QuestionID string   `json:"question_id"`
}

type MCPSearchQuestionDetail struct {
	QuestionID string `json:"question_id"`
}

type MCPSearchCommentCond struct {
	ObjectID string `json:"object_id"`
}

type MCPSearchTagCond struct {
	TagName string `json:"tag_name"`
}

type MCPSearchUserCond struct {
	Username string `json:"username"`
}

type MCPSearchQuestionInfoResp struct {
	QuestionID string `json:"question_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Link       string `json:"link"`
}

type MCPSearchAnswerInfoResp struct {
	QuestionID    string `json:"question_id"`
	QuestionTitle string `json:"question_title,omitempty"`
	AnswerID      string `json:"answer_id"`
	AnswerContent string `json:"answer_content"`
	Link          string `json:"link"`
}

type MCPSearchTagResp struct {
	TagName     string `json:"tag_name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

type MCPSearchUserInfoResp struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Link        string `json:"link"`
}

type MCPSearchCommentInfoResp struct {
	CommentID string `json:"comment_id"`
	Content   string `json:"content"`
	ObjectID  string `json:"object_id"`
	Link      string `json:"link"`
}

// MCPSemanticSearchCond is the condition for semantic search.
type MCPSemanticSearchCond struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

// MCPSemanticSearchResp is a single semantic search result.
type MCPSemanticSearchResp struct {
	ObjectID   string                      `json:"object_id"`
	ObjectType string                      `json:"object_type"`
	Title      string                      `json:"title"`
	Content    string                      `json:"content"`
	Score      float64                     `json:"score"`
	Link       string                      `json:"link"`
	Answers    []*MCPSemanticSearchAnswer  `json:"answers,omitempty"`
	Comments   []*MCPSemanticSearchComment `json:"comments,omitempty"`
}

// MCPSemanticSearchAnswer is an answer in a semantic search result.
type MCPSemanticSearchAnswer struct {
	AnswerID string                      `json:"answer_id"`
	Content  string                      `json:"content"`
	Comments []*MCPSemanticSearchComment `json:"comments,omitempty"`
}

// MCPSemanticSearchComment is a comment in a semantic search result.
type MCPSemanticSearchComment struct {
	CommentID string `json:"comment_id"`
	Content   string `json:"content"`
}

func NewMCPSemanticSearchCond(request mcp.CallToolRequest) *MCPSemanticSearchCond {
	cond := &MCPSemanticSearchCond{TopK: 5}
	if query, ok := getRequestValue(request, MCPSearchCondSemanticQuery); ok {
		cond.Query = query
	}
	if topK, ok := getRequestNumber(request, MCPSearchCondTopK); ok && topK > 0 {
		cond.TopK = topK
	}
	return cond
}

func NewMCPSearchCond(request mcp.CallToolRequest) *MCPSearchCond {
	cond := &MCPSearchCond{}
	if keyword, ok := getRequestValue(request, MCPSearchCondKeyword); ok {
		cond.Keyword = keyword
	}
	if username, ok := getRequestValue(request, MCPSearchCondUsername); ok {
		cond.Username = username
	}
	if score, ok := getRequestNumber(request, MCPSearchCondScore); ok {
		cond.Score = score
	}
	if tag, ok := getRequestValue(request, MCPSearchCondTag); ok {
		cond.Tags = strings.Split(tag, ",")
	}
	if questionID, ok := getRequestValue(request, MCPSearchCondQuestionID); ok {
		cond.QuestionID = questionID
	}
	return cond
}

func NewMCPSearchAnswerCond(request mcp.CallToolRequest) *MCPSearchCond {
	cond := &MCPSearchCond{}
	if questionID, ok := getRequestValue(request, MCPSearchCondQuestionID); ok {
		cond.QuestionID = questionID
	}
	return cond
}

func NewMCPSearchQuestionDetail(request mcp.CallToolRequest) *MCPSearchQuestionDetail {
	cond := &MCPSearchQuestionDetail{}
	if questionID, ok := getRequestValue(request, MCPSearchCondQuestionID); ok {
		cond.QuestionID = questionID
	}
	return cond
}

func NewMCPSearchCommentCond(request mcp.CallToolRequest) *MCPSearchCommentCond {
	cond := &MCPSearchCommentCond{}
	if keyword, ok := getRequestValue(request, MCPSearchCondObjectID); ok {
		cond.ObjectID = keyword
	}
	return cond
}

func NewMCPSearchTagCond(request mcp.CallToolRequest) *MCPSearchTagCond {
	cond := &MCPSearchTagCond{}
	if tagName, ok := getRequestValue(request, MCPSearchCondTagName); ok {
		cond.TagName = tagName
	}
	return cond
}

func NewMCPSearchUserCond(request mcp.CallToolRequest) *MCPSearchUserCond {
	cond := &MCPSearchUserCond{}
	if username, ok := getRequestValue(request, MCPSearchCondUsername); ok {
		cond.Username = username
	}
	return cond
}

func getRequestValue(request mcp.CallToolRequest, key string) (string, bool) {
	value, ok := request.GetArguments()[key].(string)
	if !ok {
		return "", false
	}
	return value, true
}

func getRequestNumber(request mcp.CallToolRequest, key string) (int, bool) {
	value, ok := request.GetArguments()[key].(float64)
	if !ok {
		return 0, false
	}
	return int(value), true
}

func (cond *MCPSearchCond) ToQueryString() string {
	var queryBuilder strings.Builder
	if len(cond.Keyword) > 0 {
		queryBuilder.WriteString(cond.Keyword)
	}
	if len(cond.Username) > 0 {
		queryBuilder.WriteString(" user:" + cond.Username)
	}
	if cond.Score > 0 {
		queryBuilder.WriteString(" score:" + converter.IntToString(int64(cond.Score)))
	}
	if len(cond.Tags) > 0 {
		for _, tag := range cond.Tags {
			queryBuilder.WriteString(" [" + tag + "]")
		}
	}
	return strings.TrimSpace(queryBuilder.String())
}
