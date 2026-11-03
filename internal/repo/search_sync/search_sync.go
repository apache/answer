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

package search_sync

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"
)

const cleanupPageSize = 500

func NewPluginSyncer(data *data.Data) plugin.SearchSyncer {
	return &PluginSyncer{data: data}
}

type PluginSyncer struct {
	data *data.Data
}

func (p *PluginSyncer) GetAnswersPage(ctx context.Context, page, pageSize int) (
	answerList []*plugin.SearchContent, err error) {
	answers := make([]*entity.Answer, 0)
	startNum := (page - 1) * pageSize
	err = p.data.DB.Context(ctx).
		Join("INNER", "question", "answer.question_id = question.id").
		Where("answer.status = ? AND question.show = ? AND question.status IN (?, ?)",
			entity.AnswerStatusAvailable, entity.QuestionShow,
			entity.QuestionStatusAvailable, entity.QuestionStatusClosed).
		OrderBy("answer.id ASC").Limit(pageSize, startNum).Find(&answers)
	if err != nil {
		return nil, err
	}
	return p.convertAnswers(ctx, answers)
}

func (p *PluginSyncer) GetQuestionsPage(ctx context.Context, page, pageSize int) (
	questionList []*plugin.SearchContent, err error) {
	questions := make([]*entity.Question, 0)
	startNum := (page - 1) * pageSize
	err = p.data.DB.Context(ctx).
		Where("show = ? AND status IN (?, ?)", entity.QuestionShow,
			entity.QuestionStatusAvailable, entity.QuestionStatusClosed).
		OrderBy("id ASC").Limit(pageSize, startNum).Find(&questions)
	if err != nil {
		return nil, err
	}
	return p.convertQuestions(ctx, questions)
}

func (p *PluginSyncer) convertAnswers(ctx context.Context, answers []*entity.Answer) (
	answerList []*plugin.SearchContent, err error) {
	for _, answer := range answers {
		if answer.Status != entity.AnswerStatusAvailable {
			continue
		}
		question := &entity.Question{}
		exist, err := p.data.DB.Context(ctx).Where("id = ?", answer.QuestionID).Get(question)
		if err != nil {
			log.Errorf("get question failed %s", err)
			continue
		}
		if !exist || !isPublicQuestion(question) {
			continue
		}

		tagListList := make([]*entity.TagRel, 0)
		tags := make([]string, 0)
		err = p.data.DB.Context(ctx).Where("object_id = ?", uid.DeShortID(question.ID)).
			Where("status = ?", entity.TagRelStatusAvailable).Find(&tagListList)
		if err != nil {
			log.Errorf("get tag list failed %s", err)
		}
		for _, tag := range tagListList {
			tags = append(tags, tag.TagID)
		}

		content := &plugin.SearchContent{
			ObjectID:    answer.ID,
			Title:       question.Title,
			Type:        constant.AnswerObjectType,
			Content:     answer.ParsedText,
			Answers:     0,
			Status:      plugin.SearchContentStatus(answer.Status),
			Tags:        tags,
			QuestionID:  answer.QuestionID,
			UserID:      answer.UserID,
			Views:       int64(question.ViewCount),
			Created:     answer.CreatedAt.Unix(),
			Active:      answer.UpdatedAt.Unix(),
			Score:       int64(answer.VoteCount),
			HasAccepted: answer.Accepted == schema.AnswerAcceptedEnable,
		}
		answerList = append(answerList, content)
	}
	return answerList, nil
}

func (p *PluginSyncer) convertQuestions(ctx context.Context, questions []*entity.Question) (
	questionList []*plugin.SearchContent, err error) {
	for _, question := range questions {
		if !isPublicQuestion(question) {
			continue
		}
		tagListList := make([]*entity.TagRel, 0)
		tags := make([]string, 0)
		err := p.data.DB.Context(ctx).Where("object_id = ?", question.ID).
			Where("status = ?", entity.TagRelStatusAvailable).Find(&tagListList)
		if err != nil {
			log.Errorf("get tag list failed %s", err)
		}
		for _, tag := range tagListList {
			tags = append(tags, tag.TagID)
		}
		content := &plugin.SearchContent{
			ObjectID:    question.ID,
			Title:       question.Title,
			Type:        constant.QuestionObjectType,
			Content:     question.ParsedText,
			Answers:     int64(question.AnswerCount),
			Status:      plugin.SearchContentStatus(question.Status),
			Tags:        tags,
			QuestionID:  question.ID,
			UserID:      question.UserID,
			Views:       int64(question.ViewCount),
			Created:     question.CreatedAt.Unix(),
			Active:      question.UpdatedAt.Unix(),
			Score:       int64(question.VoteCount),
			HasAccepted: question.AcceptedAnswerID != "" && question.AcceptedAnswerID != "0",
		}
		questionList = append(questionList, content)
	}
	return questionList, nil
}

func isPublicQuestion(question *entity.Question) bool {
	return question != nil && question.Show == entity.QuestionShow &&
		(question.Status == entity.QuestionStatusAvailable || question.Status == entity.QuestionStatusClosed)
}

// RemoveRestrictedContents removes documents written before visibility filtering
// was enforced. It uses the existing Search interface so search plugins remain
// source and binary compatible.
func RemoveRestrictedContents(ctx context.Context, data *data.Data, search plugin.Search) error {
	if data == nil || data.DB == nil || search == nil {
		return nil
	}

	questions := make([]*entity.Question, 0, cleanupPageSize)
	// Object IDs are positive numeric IDs. Start from zero so the predicate
	// also works with databases that store the ID column as BIGINT.
	lastID := "0"
	for {
		questions = questions[:0]
		if err := data.DB.Context(ctx).Where("id > ?", lastID).OrderBy("id ASC").Limit(cleanupPageSize).Find(&questions); err != nil {
			return fmt.Errorf("list questions for search cleanup: %w", err)
		}
		for _, question := range questions {
			if isPublicQuestion(question) {
				answers := make([]*entity.Answer, 0)
				if err := data.DB.Context(ctx).Where("question_id = ? AND status <> ?", question.ID,
					entity.AnswerStatusAvailable).Find(&answers); err != nil {
					return fmt.Errorf("list restricted answers for question %s: %w", question.ID, err)
				}
				for _, answer := range answers {
					if err := search.DeleteContent(ctx, uid.DeShortID(answer.ID)); err != nil {
						return fmt.Errorf("delete restricted answer search document %s: %w", answer.ID, err)
					}
				}
				continue
			}

			if err := search.DeleteContent(ctx, uid.DeShortID(question.ID)); err != nil {
				return fmt.Errorf("delete restricted question search document %s: %w", question.ID, err)
			}
			answers := make([]*entity.Answer, 0)
			if err := data.DB.Context(ctx).Where("question_id = ?", question.ID).Find(&answers); err != nil {
				return fmt.Errorf("list answers for restricted question %s: %w", question.ID, err)
			}
			for _, answer := range answers {
				if err := search.DeleteContent(ctx, uid.DeShortID(answer.ID)); err != nil {
					return fmt.Errorf("delete answer search document %s: %w", answer.ID, err)
				}
			}
		}
		if len(questions) < cleanupPageSize {
			break
		}
		lastID = questions[len(questions)-1].ID
	}
	return nil
}
