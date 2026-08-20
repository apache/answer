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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/answer"
	"github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/stretchr/testify/require"
)

func TestPersonalAnswerPageRespectsHiddenQuestionVisibility(t *testing.T) {
	ctx := context.Background()
	questionRepo := question.NewQuestionRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	answerRepo := answer.NewAnswerRepo(testDataSource, nil, nil, nil)

	newQuestion := func(title string, show int) *entity.Question {
		q := &entity.Question{
			UserID:       "personal-answer-question-owner",
			Title:        title,
			OriginalText: title,
			ParsedText:   title,
			Status:       entity.QuestionStatusAvailable,
			Show:         show,
		}
		require.NoError(t, questionRepo.AddQuestion(ctx, q))
		return q
	}

	visibleQuestion := newQuestion("visible question", entity.QuestionShow)
	hiddenQuestion := newQuestion("hidden question", entity.QuestionHide)
	answers := []*entity.Answer{
		{QuestionID: visibleQuestion.ID, UserID: "personal-answer-author", OriginalText: "visible answer", ParsedText: "visible answer", Status: entity.AnswerStatusAvailable},
		{QuestionID: hiddenQuestion.ID, UserID: "personal-answer-author", OriginalText: "hidden answer", ParsedText: "hidden answer", Status: entity.AnswerStatusAvailable},
	}
	for _, item := range answers {
		_, err := testDataSource.DB.Context(ctx).Insert(item)
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		for _, item := range answers {
			_, _ = testDataSource.DB.Context(ctx).ID(item.ID).Delete(&entity.Answer{})
		}
		for _, item := range []*entity.Question{visibleQuestion, hiddenQuestion} {
			_, _ = testDataSource.DB.Context(ctx).ID(item.ID).Delete(&entity.Question{})
		}
	})

	publicAnswers, publicTotal, err := answerRepo.GetPersonalAnswerPage(ctx, &entity.PersonalAnswerPageQueryCond{
		Page:     1,
		PageSize: 20,
		UserID:   "personal-answer-author",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), publicTotal)
	require.Len(t, publicAnswers, 1)
	require.Equal(t, visibleQuestion.ID, publicAnswers[0].QuestionID)

	ownerAnswers, ownerTotal, err := answerRepo.GetPersonalAnswerPage(ctx, &entity.PersonalAnswerPageQueryCond{
		Page:       1,
		PageSize:   20,
		UserID:     "personal-answer-author",
		ShowHidden: true,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), ownerTotal)
	require.Len(t, ownerAnswers, 2)
}
