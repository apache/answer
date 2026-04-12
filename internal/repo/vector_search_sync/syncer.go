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

package vector_search_sync

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"
)

// NewPluginSyncer creates a new VectorSearchSyncer that reads from the database.
func NewPluginSyncer(data *data.Data) plugin.VectorSearchSyncer {
	return &PluginSyncer{data: data}
}

// PluginSyncer implements plugin.VectorSearchSyncer.
// It aggregates question/answer text with comments for vector embedding.
type PluginSyncer struct {
	data *data.Data
}

// GetQuestionsPage returns a page of questions with aggregated text
// (question title + body + all answers + all comments).
func (p *PluginSyncer) GetQuestionsPage(ctx context.Context, page, pageSize int) (
	[]*plugin.VectorSearchContent, error) {
	questions := make([]*entity.Question, 0)
	startNum := (page - 1) * pageSize
	err := p.data.DB.Context(ctx).Limit(pageSize, startNum).Find(&questions)
	if err != nil {
		return nil, err
	}
	return p.buildQuestionContents(ctx, questions)
}

// GetAnswersPage returns a page of answers with aggregated text
// (parent question title + answer body + answer comments).
func (p *PluginSyncer) GetAnswersPage(ctx context.Context, page, pageSize int) (
	[]*plugin.VectorSearchContent, error) {
	answers := make([]*entity.Answer, 0)
	startNum := (page - 1) * pageSize
	err := p.data.DB.Context(ctx).Limit(pageSize, startNum).Find(&answers)
	if err != nil {
		return nil, err
	}
	return p.buildAnswerContents(ctx, answers)
}

// buildQuestionContents aggregates each question with its answers and comments.
func (p *PluginSyncer) buildQuestionContents(ctx context.Context, questions []*entity.Question) (
	[]*plugin.VectorSearchContent, error) {
	result := make([]*plugin.VectorSearchContent, 0, len(questions))
	for _, q := range questions {
		meta := plugin.VectorSearchMetadata{
			QuestionID: uid.DeShortID(q.ID),
		}

		var parts []string
		parts = append(parts, fmt.Sprintf("Question: %s\n%s", q.Title, q.OriginalText))

		// Get answers for this question
		answers := make([]*entity.Answer, 0)
		err := p.data.DB.Context(ctx).Where("question_id = ?", q.ID).Find(&answers)
		if err != nil {
			log.Warnf("get answers for question %s failed: %v", q.ID, err)
		} else {
			for _, a := range answers {
				parts = append(parts, fmt.Sprintf("Answer: %s", a.OriginalText))
				answerMeta := plugin.VectorSearchMetadataAnswer{
					AnswerID: uid.DeShortID(a.ID),
				}

				// Get comments on this answer
				answerComments := make([]*entity.Comment, 0)
				err := p.data.DB.Context(ctx).Where("object_id = ?", a.ID).
					OrderBy("created_at ASC").Limit(50).Find(&answerComments)
				if err != nil {
					log.Warnf("get comments for answer %s failed: %v", a.ID, err)
				} else {
					for _, c := range answerComments {
						parts = append(parts, fmt.Sprintf("Comment on answer: %s", c.OriginalText))
						answerMeta.Comments = append(answerMeta.Comments, plugin.VectorSearchMetadataComment{
							CommentID: uid.DeShortID(c.ID),
						})
					}
				}
				meta.Answers = append(meta.Answers, answerMeta)
			}
		}

		// Get comments on the question
		questionComments := make([]*entity.Comment, 0)
		err = p.data.DB.Context(ctx).Where("object_id = ?", q.ID).
			OrderBy("created_at ASC").Limit(50).Find(&questionComments)
		if err != nil {
			log.Warnf("get comments for question %s failed: %v", q.ID, err)
		} else {
			for _, c := range questionComments {
				parts = append(parts, fmt.Sprintf("Comment: %s", c.OriginalText))
				meta.Comments = append(meta.Comments, plugin.VectorSearchMetadataComment{
					CommentID: uid.DeShortID(c.ID),
				})
			}
		}

		metaJSON, _ := json.Marshal(meta)
		result = append(result, &plugin.VectorSearchContent{
			ObjectID:   uid.DeShortID(q.ID),
			ObjectType: "question",
			Title:      q.Title,
			Content:    strings.Join(parts, "\n\n"),
			Metadata:   string(metaJSON),
		})
	}
	return result, nil
}

// buildAnswerContents aggregates each answer with its parent question title and comments.
func (p *PluginSyncer) buildAnswerContents(ctx context.Context, answers []*entity.Answer) (
	[]*plugin.VectorSearchContent, error) {
	result := make([]*plugin.VectorSearchContent, 0, len(answers))
	for _, a := range answers {
		// Get parent question for title
		question := &entity.Question{}
		exist, err := p.data.DB.Context(ctx).Where("id = ?", a.QuestionID).Get(question)
		if err != nil {
			log.Errorf("get question %s failed: %v", a.QuestionID, err)
			continue
		}
		if !exist {
			continue
		}

		meta := plugin.VectorSearchMetadata{
			QuestionID: uid.DeShortID(a.QuestionID),
			AnswerID:   uid.DeShortID(a.ID),
		}

		var parts []string
		parts = append(parts, fmt.Sprintf("Question: %s", question.Title))
		parts = append(parts, fmt.Sprintf("Answer: %s", a.OriginalText))

		answerMeta := plugin.VectorSearchMetadataAnswer{
			AnswerID: uid.DeShortID(a.ID),
		}

		// Get comments on this answer
		answerComments := make([]*entity.Comment, 0)
		err = p.data.DB.Context(ctx).Where("object_id = ?", a.ID).
			OrderBy("created_at ASC").Limit(50).Find(&answerComments)
		if err != nil {
			log.Warnf("get comments for answer %s failed: %v", a.ID, err)
		} else {
			for _, c := range answerComments {
				parts = append(parts, fmt.Sprintf("Comment: %s", c.OriginalText))
				answerMeta.Comments = append(answerMeta.Comments, plugin.VectorSearchMetadataComment{
					CommentID: uid.DeShortID(c.ID),
				})
			}
		}
		meta.Answers = append(meta.Answers, answerMeta)

		metaJSON, _ := json.Marshal(meta)
		result = append(result, &plugin.VectorSearchContent{
			ObjectID:   uid.DeShortID(a.ID),
			ObjectType: "answer",
			Title:      question.Title,
			Content:    strings.Join(parts, "\n\n"),
			Metadata:   string(metaJSON),
		})
	}
	return result, nil
}
