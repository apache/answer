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
	"time"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/comment"
	"github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/repo/unique"
	commentService "github.com/apache/answer/internal/service/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The personal comment list must paginate only the comments whose target is
// visible to the viewer; filtering after pagination leaves empty pages and an
// unfiltered count (https://github.com/apache/answer/issues/1645).
func TestPersonalCommentPageFiltersHiddenQuestions(t *testing.T) {
	ctx := context.Background()
	questionRepo := question.NewQuestionRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	commentRepo := comment.NewCommentRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	const (
		author        = "8101"
		stranger      = "8102"
		questionOwner = "8103"
	)

	newQuestion := func(title string, show int) *entity.Question {
		q := &entity.Question{
			UserID:       questionOwner,
			Title:        title,
			OriginalText: title,
			ParsedText:   title,
			Status:       entity.QuestionStatusAvailable,
			Show:         show,
		}
		require.NoError(t, questionRepo.AddQuestion(ctx, q))
		return q
	}

	// Same scenario as the issue: the author comments in the order C, B, A, so
	// the newest-first order is A, B, C. Only C's question stays public.
	hiddenQuestionA := newQuestion("hidden question A", entity.QuestionHide)
	hiddenQuestionB := newQuestion("hidden question B", entity.QuestionHide)
	visibleQuestionC := newQuestion("visible question C", entity.QuestionShow)

	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	addComment := func(objectID, questionID string, createdAt time.Time) *entity.Comment {
		c := &entity.Comment{
			UserID:       author,
			ObjectID:     objectID,
			QuestionID:   questionID,
			Status:       entity.CommentStatusAvailable,
			OriginalText: "comment",
			ParsedText:   "comment",
		}
		require.NoError(t, commentRepo.AddComment(ctx, c))
		// created_at is set by xorm on insert; pin it so the newest-first order
		// does not depend on insert timing.
		_, err := testDataSource.DB.Context(ctx).Table(&entity.Comment{}).ID(c.ID).
			Update(map[string]any{"created_at": createdAt})
		require.NoError(t, err)
		return c
	}

	commentC := addComment(visibleQuestionC.ID, visibleQuestionC.ID, baseTime)
	commentB := addComment(hiddenQuestionB.ID, hiddenQuestionB.ID, baseTime.Add(time.Hour))
	commentA := addComment(hiddenQuestionA.ID, hiddenQuestionA.ID, baseTime.Add(2*time.Hour))

	query := func(viewer string, isAdminModerator bool, page int) ([]*entity.Comment, int64) {
		list, total, err := commentRepo.GetPersonalCommentPage(ctx, &commentService.CommentQuery{
			PageCond:         pager.PageCond{Page: page, PageSize: 2},
			UserID:           author,
			QueryCond:        "created_at",
			LoginUserID:      viewer,
			IsAdminModerator: isAdminModerator,
		})
		require.NoError(t, err)
		return list, total
	}

	// A stranger only sees the comment on the visible question: the first page
	// contains C and the count is 1, instead of an empty page with count 3.
	page1, total := query(stranger, false, 1)
	assert.Equal(t, int64(1), total)
	require.Len(t, page1, 1)
	assert.Equal(t, commentC.ID, page1[0].ID)

	page2, total := query(stranger, false, 2)
	assert.Equal(t, int64(1), total)
	assert.Empty(t, page2)

	// The owner of the hidden questions sees all three comments.
	page1, total = query(questionOwner, false, 1)
	assert.Equal(t, int64(3), total)
	require.Len(t, page1, 2)
	assert.Equal(t, commentA.ID, page1[0].ID)
	assert.Equal(t, commentB.ID, page1[1].ID)

	// So does an admin or moderator.
	page1, total = query(stranger, true, 1)
	assert.Equal(t, int64(3), total)
	require.Len(t, page1, 2)
	assert.Equal(t, commentA.ID, page1[0].ID)
	assert.Equal(t, commentB.ID, page1[1].ID)
}

func TestPersonalCommentPageFiltersRestrictedAnswers(t *testing.T) {
	ctx := context.Background()
	questionRepo := question.NewQuestionRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	commentRepo := comment.NewCommentRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	const (
		author        = "8201"
		stranger      = "8202"
		answerAuthor  = "8204"
		questionOwner = "8203"
	)

	newQuestion := func(title string, show int, status int) *entity.Question {
		q := &entity.Question{
			UserID:       questionOwner,
			Title:        title,
			OriginalText: title,
			ParsedText:   title,
			Status:       status,
			Show:         show,
		}
		require.NoError(t, questionRepo.AddQuestion(ctx, q))
		return q
	}

	visibleQuestion := newQuestion("visible question", entity.QuestionShow, entity.QuestionStatusAvailable)
	hiddenQuestion := newQuestion("hidden question", entity.QuestionHide, entity.QuestionStatusAvailable)

	newAnswer := func(q *entity.Question, status int) *entity.Answer {
		a := &entity.Answer{
			QuestionID:   q.ID,
			UserID:       answerAuthor,
			OriginalText: "answer",
			ParsedText:   "answer",
			Status:       status,
		}
		_, err := testDataSource.DB.Context(ctx).Insert(a)
		require.NoError(t, err)
		return a
	}

	visibleAnswer := newAnswer(visibleQuestion, entity.AnswerStatusAvailable)
	answerOnHiddenQuestion := newAnswer(hiddenQuestion, entity.AnswerStatusAvailable)
	deletedAnswer := newAnswer(visibleQuestion, entity.AnswerStatusDeleted)

	baseTime := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	addComment := func(objectID, questionID string, createdAt time.Time) *entity.Comment {
		c := &entity.Comment{
			UserID:       author,
			ObjectID:     objectID,
			QuestionID:   questionID,
			Status:       entity.CommentStatusAvailable,
			OriginalText: "comment",
			ParsedText:   "comment",
		}
		require.NoError(t, commentRepo.AddComment(ctx, c))
		_, err := testDataSource.DB.Context(ctx).Table(&entity.Comment{}).ID(c.ID).
			Update(map[string]any{"created_at": createdAt})
		require.NoError(t, err)
		return c
	}

	commentOnVisible := addComment(visibleAnswer.ID, visibleQuestion.ID, baseTime)
	commentOnHiddenParent := addComment(answerOnHiddenQuestion.ID, hiddenQuestion.ID, baseTime.Add(time.Hour))
	commentOnDeleted := addComment(deletedAnswer.ID, visibleQuestion.ID, baseTime.Add(2*time.Hour))

	query := func(viewer string, isAdminModerator bool, page int) ([]*entity.Comment, int64) {
		list, total, err := commentRepo.GetPersonalCommentPage(ctx, &commentService.CommentQuery{
			PageCond:         pager.PageCond{Page: page, PageSize: 2},
			UserID:           author,
			QueryCond:        "created_at",
			LoginUserID:      viewer,
			IsAdminModerator: isAdminModerator,
		})
		require.NoError(t, err)
		return list, total
	}

	// A stranger only sees the comment on the available answer under a visible
	// question.
	page1, total := query(stranger, false, 1)
	assert.Equal(t, int64(1), total)
	require.Len(t, page1, 1)
	assert.Equal(t, commentOnVisible.ID, page1[0].ID)

	// The answer author also sees their comment on the deleted answer, but not
	// the one under the hidden question they don't own.
	page1, total = query(answerAuthor, false, 1)
	assert.Equal(t, int64(2), total)
	require.Len(t, page1, 2)
	assert.Equal(t, commentOnDeleted.ID, page1[0].ID)
	assert.Equal(t, commentOnVisible.ID, page1[1].ID)

	// The owner of the hidden question sees the comment under it instead.
	page1, total = query(questionOwner, false, 1)
	assert.Equal(t, int64(2), total)
	require.Len(t, page1, 2)
	assert.Equal(t, commentOnHiddenParent.ID, page1[0].ID)
	assert.Equal(t, commentOnVisible.ID, page1[1].ID)

	// Admins and moderators see all three.
	page1, total = query(stranger, true, 1)
	assert.Equal(t, int64(3), total)
	require.Len(t, page1, 2)
	assert.Equal(t, commentOnDeleted.ID, page1[0].ID)
	assert.Equal(t, commentOnHiddenParent.ID, page1[1].ID)
}
