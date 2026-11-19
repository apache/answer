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

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/comment"
	"github.com/apache/answer/internal/repo/unique"
	commentService "github.com/apache/answer/internal/service/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	require.NoError(t, err)
}

func Test_commentRepo_GetCommentPage(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	require.NoError(t, err)

	resp, total, err := commentRepo.GetCommentPage(context.TODO(), &commentService.CommentQuery{
		PageCond: pager.PageCond{
			Page:     1,
			PageSize: 10,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, resp[0].ID, testCommentEntity.ID)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	require.NoError(t, err)
}

func Test_commentRepo_UpdateComment(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	commonCommentRepo := comment.NewCommentCommonRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()
	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	require.NoError(t, err)

	testCommentEntity.ParsedText = "test"
	err = commentRepo.UpdateCommentContent(context.TODO(), testCommentEntity.ID, "test", "test")
	require.NoError(t, err)

	newComment, exist, err := commonCommentRepo.GetComment(context.TODO(), testCommentEntity.ID)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testCommentEntity.ParsedText, newComment.ParsedText)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	require.NoError(t, err)
}

func Test_commentRepo_CannotGetDeletedComment(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	commentRepo := comment.NewCommentRepo(testDataSource, uniqueIDRepo)
	testCommentEntity := buildCommentEntity()

	err := commentRepo.AddComment(context.TODO(), testCommentEntity)
	require.NoError(t, err)

	err = commentRepo.RemoveComment(context.TODO(), testCommentEntity.ID)
	require.NoError(t, err)

	_, exist, err := commentRepo.GetComment(context.TODO(), testCommentEntity.ID)
	require.NoError(t, err)
	assert.False(t, exist)
}

func TestCommentRepoObjectFiltering(t *testing.T) {
	ctx := context.Background()
	commentRepo := comment.NewCommentRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	const userID = "700101"
	const firstObjectID = "10010000000700101"
	const secondObjectID = "10010000000700102"
	comments := []*entity.Comment{
		{UserID: userID, ObjectID: firstObjectID, Status: entity.CommentStatusAvailable},
		{UserID: userID, ObjectID: firstObjectID, Status: entity.CommentStatusAvailable},
		{UserID: userID, ObjectID: secondObjectID, Status: entity.CommentStatusAvailable},
		{UserID: userID, ObjectID: "10010000000700103", Status: entity.CommentStatusDeleted},
		{UserID: userID, ObjectID: "10010000000700104", Status: entity.CommentStatusPending},
		{UserID: "700102", ObjectID: "10010000000700105", Status: entity.CommentStatusAvailable},
	}
	t.Cleanup(func() {
		for _, item := range comments {
			_, err := testDataSource.DB.Context(ctx).ID(item.ID).Delete(&entity.Comment{})
			require.NoError(t, err)
		}
	})
	for _, item := range comments {
		require.NoError(t, commentRepo.AddComment(ctx, item))
	}

	objectIDs, err := commentRepo.GetUserCommentObjectIDs(ctx, userID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{firstObjectID, secondObjectID}, objectIDs)
	objectIDs, err = commentRepo.GetUserCommentObjectIDs(ctx, "700103")
	require.NoError(t, err)
	require.NotNil(t, objectIDs)
	require.Empty(t, objectIDs)

	for _, tc := range []struct {
		name      string
		objectIDs []string
		count     int64
	}{
		{"unrestricted", nil, 3},
		{"empty filter", []string{}, 0},
		{"repeated target", []string{firstObjectID}, 2},
		{"both targets", []string{firstObjectID, secondObjectID}, 3},
		{"missing target", []string{"10010000000700109"}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for page := 1; page <= 2; page++ {
				list, count, err := commentRepo.GetCommentPage(ctx, &commentService.CommentQuery{
					PageCond: pager.PageCond{Page: page, PageSize: 2}, UserID: userID, ObjectIDs: tc.objectIDs,
				})
				require.NoError(t, err)
				require.Equal(t, tc.count, count)
				wantLength := min(2, max(0, int(tc.count)-(page-1)*2))
				require.Len(t, list, wantLength)
				for _, item := range list {
					if tc.objectIDs != nil {
						require.Contains(t, tc.objectIDs, item.ObjectID)
					}
				}
			}
		})
	}
}
