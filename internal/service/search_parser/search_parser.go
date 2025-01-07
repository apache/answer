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

package search_parser

import (
	"context"
	"fmt"
	"github.com/apache/answer/internal/base/constant"
	"regexp"
	"strings"

	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/converter"
)

type SearchParser struct {
	tagCommonService *tag_common.TagCommonService
	userCommon       *usercommon.UserCommon
}

func NewSearchParser(tagCommonService *tag_common.TagCommonService, userCommon *usercommon.UserCommon) *SearchParser {
	return &SearchParser{
		tagCommonService: tagCommonService,
		userCommon:       userCommon,
	}
}

// ParseStructure parse search structure, maybe match one of type all/questions/answers,
// but if match two type, it will return false
func (sp *SearchParser) ParseStructure(ctx context.Context, dto *schema.SearchDTO) (cond *schema.SearchCondition) {
	cond = &schema.SearchCondition{}
	var (
		query      = dto.Query
		limitWords = 5
	)

	// match tags
	cond.Tags = sp.parseTags(ctx, &query)

	// match all
	cond.UserID = sp.parseUserID(ctx, &query, dto.UserID)
	cond.VoteAmount = sp.parseVotes(&query)
	cond.Words = sp.parseWithin(&query)

	// match questions
	cond.NotAccepted = sp.parseNotAccepted(&query)
	if cond.NotAccepted {
		cond.TargetType = constant.QuestionObjectType
	}
	cond.Views = sp.parseViews(&query)
	if cond.Views != -1 {
		cond.TargetType = constant.QuestionObjectType
	}
	cond.AnswerAmount = sp.parseAnswers(&query)
	if cond.AnswerAmount != -1 {
		cond.TargetType = constant.QuestionObjectType
	}

	// match answers
	cond.Accepted = sp.parseAccepted(&query)
	if cond.Accepted {
		cond.TargetType = constant.AnswerObjectType
	}
	cond.QuestionID = sp.parseQuestionID(&query)
	if cond.QuestionID != "" {
		cond.TargetType = constant.AnswerObjectType
	}

	if sp.parseIsQuestion(&query) {
		cond.TargetType = constant.QuestionObjectType
	}
	if sp.parseIsAnswer(&query) {
		cond.TargetType = constant.AnswerObjectType
	}

	if len(strings.TrimSpace(query)) > 0 {
		words := strings.Split(strings.TrimSpace(query), " ")
		cond.Words = append(cond.Words, words...)
	}

	// check limit words
	if len(cond.Words) > limitWords {
		cond.Words = cond.Words[:limitWords]
	}
	return
}

// parseTags parse search tags, return tag ids array
func (sp *SearchParser) parseTags(ctx context.Context, query *string) (tags [][]string) {
	var (
		// expire tag pattern
		exprTag = `\[(.*?)\]`
		q       = *query
		limit   = 5
	)

	re := regexp.MustCompile(exprTag)
	res := re.FindAllStringSubmatch(q, -1)
	if len(res) == 0 {
		return
	}

	tags = make([][]string, 0)
	for _, item := range res {
		tagGroup := make([]string, 0)
		tag, exists, err := sp.tagCommonService.GetTagBySlugName(ctx, item[1])
		if err != nil || !exists {
			continue
		}
		tagGroup = append(tagGroup, tag.ID)
		if tag.MainTagID > 0 {
			tagGroup = append(tagGroup, fmt.Sprintf("%d", tag.MainTagID))
		}
		synIDs, err := sp.tagCommonService.GetTagIDsByMainTagID(ctx, tag.ID)
		if err != nil || !exists {
			continue
		}
		tagGroup = append(tagGroup, tag.ID)
		tagGroup = append(tagGroup, synIDs...)
		tagGroup = converter.UniqueArray(tagGroup)
		tags = append(tags, tagGroup)
	}

	// limit maximum 5 tags
	if len(tags) > limit {
		tags = tags[:limit]
	}

	q = strings.TrimSpace(re.ReplaceAllString(q, ""))
	*query = q
	return
}

// parseUserID return user id or current login user id
func (sp *SearchParser) parseUserID(ctx context.Context, query *string, currentUserID string) (userID string) {
	var (
		exprUsername = `user:(\S+)`
		exprMe       = "user:me"
		q            = *query
	)

	re := regexp.MustCompile(exprUsername)
	res := re.FindStringSubmatch(q)
	if strings.Contains(q, exprMe) {
		userID = currentUserID
		q = strings.ReplaceAll(q, exprMe, "")
	} else if len(res) > 1 {
		name := res[1]
		user, has, err := sp.userCommon.GetUserBasicInfoByUserName(ctx, name)
		if err == nil && has {
			userID = user.ID
			q = re.ReplaceAllString(q, "")
		}
	}
	*query = strings.TrimSpace(q)
	return
}

// parseVotes return the votes of search query
func (sp *SearchParser) parseVotes(query *string) (votes int) {
	var (
		expr = `score:(\d+)`
		q    = *query
	)
	votes = -1

	re := regexp.MustCompile(expr)
	res := re.FindStringSubmatch(q)
	if len(res) > 1 {
		votes = converter.StringToInt(res[1])
		q = re.ReplaceAllString(q, "")
	}

	*query = strings.TrimSpace(q)
	return
}

// parseWithin parse quotes within words like: "hello world"
func (sp *SearchParser) parseWithin(query *string) (words []string) {
	var (
		q    = *query
		expr = `(?U)(".+")`
	)
	re := regexp.MustCompile(expr)
	matches := re.FindAllStringSubmatch(q, -1)
	words = []string{}
	for _, match := range matches {
		if len(match[1]) == 0 {
			continue
		}
		words = append(words, match[1])
	}
	q = re.ReplaceAllString(q, "")
	*query = strings.TrimSpace(q)
	return
}

// parseNotAccepted return the question has not accepted the answer
func (sp *SearchParser) parseNotAccepted(query *string) (notAccepted bool) {
	var (
		q    = *query
		expr = `hasaccepted:no`
	)

	if strings.Contains(q, expr) {
		q = strings.ReplaceAll(q, expr, "")
		notAccepted = true
	}

	*query = strings.TrimSpace(q)
	return
}

// parseIsQuestion check the result if only limit question or not
func (sp *SearchParser) parseIsQuestion(query *string) (isQuestion bool) {
	var (
		q    = *query
		expr = `is:question`
	)

	if strings.Contains(q, expr) {
		q = strings.ReplaceAll(q, expr, "")
		isQuestion = true
	}

	*query = strings.TrimSpace(q)
	return
}

// parseViews check search has views or not
func (sp *SearchParser) parseViews(query *string) (views int) {
	var (
		q    = *query
		expr = `views:(\d+)`
	)
	views = -1

	re := regexp.MustCompile(expr)
	res := re.FindStringSubmatch(q)
	if len(res) > 1 {
		views = converter.StringToInt(res[1])
		q = re.ReplaceAllString(q, "")
	}
	*query = strings.TrimSpace(q)
	return
}

// parseAnswers check whether specified answer count for question
func (sp *SearchParser) parseAnswers(query *string) (answers int) {
	var (
		q    = *query
		expr = `answers:(\d+)`
	)
	answers = -1

	re := regexp.MustCompile(expr)
	res := re.FindStringSubmatch(q)
	if len(res) > 1 {
		answers = converter.StringToInt(res[1])
		q = re.ReplaceAllString(q, "")
	}

	*query = strings.TrimSpace(q)
	return
}

// parseAccepted check the search is limit accepted answer or not
func (sp *SearchParser) parseAccepted(query *string) (accepted bool) {
	var (
		q    = *query
		expr = `isaccepted:yes`
	)

	if strings.Contains(q, expr) {
		accepted = true
		q = strings.ReplaceAll(q, expr, "")
	}

	*query = strings.TrimSpace(q)
	return
}

// parseQuestionID check whether specified question's id
func (sp *SearchParser) parseQuestionID(query *string) (questionID string) {
	var (
		q    = *query
		expr = `inquestion:(\d+)`
	)

	re := regexp.MustCompile(expr)
	res := re.FindStringSubmatch(q)
	if len(res) == 2 {
		questionID = res[1]
		q = re.ReplaceAllString(q, "")
	}

	*query = strings.TrimSpace(q)
	return
}

// parseIsAnswer check the result if only limit answer or not
func (sp *SearchParser) parseIsAnswer(query *string) (isAnswer bool) {
	var (
		q    = *query
		expr = `is:answer`
	)

	if strings.Contains(q, expr) {
		isAnswer = true
		q = strings.ReplaceAll(q, expr, "")
	}

	*query = strings.TrimSpace(q)
	return
}
