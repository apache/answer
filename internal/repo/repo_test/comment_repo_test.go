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

	"github.com/apache/incubator-answer/internal/base/pager"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/repo/comment"
	"github.com/apache/incubator-answer/internal/repo/unique"
	commentService "github.com/apache/incubator-answer/internal/service/comment"
	"github.com/stretchr/testify/assert"
)

func buildCommentEntity() *entity.Comment {
	return &entity.Comment{
		UserID:       "1",
		ObjectID:     "1",
		QuestionID:   "1",
		VoteCount:    1,
		Status:       entity.CommentStatusAvailable,
		OriginalText: "# title",
		ParsedText:   "<h1>Title</h1>",
	}
}

func Test_commentRepo_AddComment(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	assert.NoError(t, err)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
}

func Test_commentRepo_GetCommentPage(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	assert.NoError(t, err)

	resp, total, err := commentRepo.GetCommentPage(context.TODO(), &commentService.CommentQuery{
		PageCond: pager.PageCond{
			Page:     1,
			PageSize: 10,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, total, int64(1))
	assert.Equal(t, resp[0].ID, testCommentEntity.ID)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
}

func Test_commentRepo_UpdateComment(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	commonCommentRepo := comment.NewCommentCommonRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	assert.NoError(t, err)

	testCommentEntity.ParsedText = "test"
	err = commentRepo.UpdateCommentContent(context.TODO(), testCommentEntity.ID, "test", "test")
	assert.NoError(t, err)

	newComment, exist, err := commonCommentRepo.GetComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testCommentEntity.ParsedText, newComment.ParsedText)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
}

func Test_commentRepo_CannotGetDeletedComment(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()

	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	assert.NoError(t, err)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)

	_, exist, err := commentRepo.GetComment(context.TODO(), testCommentEntity.ID)
	assert.NoError(t, err)
	assert.False(t, exist)
}
