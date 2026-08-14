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
	activityrepo "github.com/apache/answer/internal/repo/activity"
	answerrepo "github.com/apache/answer/internal/repo/answer"
	collectionrepo "github.com/apache/answer/internal/repo/collection"
	commentrepo "github.com/apache/answer/internal/repo/comment"
	commentservice "github.com/apache/answer/internal/service/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonalRepositoriesExcludeDeletedContentBeforePagination(t *testing.T) {
	ctx := context.Background()
	const (
		questionID   = "990000001"
		answerID     = "990000002"
		commentID    = "990000003"
		collectionID = "990000004"
		activityID   = "990000005"
		userID       = "990000006"
		activityType = 990001
	)

	question := &entity.Question{
		ID:     questionID,
		UserID: userID,
		Title:  "visible question",
		Status: entity.QuestionStatusAvailable,
	}
	answer := &entity.Answer{
		ID:           answerID,
		QuestionID:   questionID,
		UserID:       userID,
		OriginalText: "visible answer",
		ParsedText:   "visible answer",
		Status:       entity.AnswerStatusAvailable,
	}
	comment := &entity.Comment{
		ID:           commentID,
		ObjectID:     answerID,
		QuestionID:   questionID,
		UserID:       userID,
		OriginalText: "visible comment",
		ParsedText:   "visible comment",
		Status:       entity.CommentStatusAvailable,
	}
	collection := &entity.Collection{ID: collectionID, ObjectID: questionID, UserID: userID}
	activity := &entity.Activity{
		ID:           activityID,
		ObjectID:     answerID,
		UserID:       userID,
		ActivityType: activityType,
	}

	_, err := testDataSource.DB.Insert(question, answer, comment, collection, activity)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = testDataSource.DB.ID(activityID).Delete(&entity.Activity{})
		_, _ = testDataSource.DB.ID(collectionID).Delete(&entity.Collection{})
		_, _ = testDataSource.DB.ID(commentID).Delete(&entity.Comment{})
		_, _ = testDataSource.DB.ID(answerID).Delete(&entity.Answer{})
		_, _ = testDataSource.DB.ID(questionID).Delete(&entity.Question{})
	})

	answerRepo := answerrepo.NewAnswerRepo(testDataSource, nil, nil, nil)
	collectionRepo := collectionrepo.NewCollectionRepo(testDataSource, nil)
	commentRepo := commentrepo.NewCommentRepo(testDataSource, nil)
	voteRepo := activityrepo.NewVoteRepo(testDataSource, nil, nil, nil)

	personalCounts := func() (answers, collections, comments, votes int64) {
		_, answers, err = answerRepo.GetPersonalAnswerPage(ctx, &entity.PersonalAnswerPageQueryCond{
			Page: 1, PageSize: 10, UserID: userID,
		})
		require.NoError(t, err)

		_, collections, err = collectionRepo.SearchList(ctx, &entity.CollectionSearch{
			Collection: entity.Collection{UserID: userID}, Page: 1, PageSize: 10,
		})
		require.NoError(t, err)

		_, comments, err = commentRepo.GetCommentPage(ctx, &commentservice.CommentQuery{
			PageCond: pager.PageCond{Page: 1, PageSize: 10}, UserID: userID, ExcludeDeletedContent: true,
		})
		require.NoError(t, err)

		_, votes, err = voteRepo.ListUserVotes(ctx, userID, 1, 10, []int{activityType})
		require.NoError(t, err)
		return
	}

	answers, collections, comments, votes := personalCounts()
	assert.Equal(t, int64(1), answers)
	assert.Equal(t, int64(1), collections)
	assert.Equal(t, int64(1), comments)
	assert.Equal(t, int64(1), votes)

	_, err = testDataSource.DB.ID(answerID).Cols("status").Update(&entity.Answer{Status: entity.AnswerStatusDeleted})
	require.NoError(t, err)
	answers, collections, comments, votes = personalCounts()
	assert.Equal(t, int64(0), answers)
	assert.Equal(t, int64(1), collections)
	assert.Equal(t, int64(0), comments)
	assert.Equal(t, int64(0), votes)

	_, err = testDataSource.DB.ID(answerID).Cols("status").Update(&entity.Answer{Status: entity.AnswerStatusAvailable})
	require.NoError(t, err)
	_, err = testDataSource.DB.ID(questionID).Cols("status").Update(&entity.Question{Status: entity.QuestionStatusDeleted})
	require.NoError(t, err)
	answers, collections, comments, votes = personalCounts()
	assert.Equal(t, int64(0), answers)
	assert.Equal(t, int64(0), collections)
	assert.Equal(t, int64(0), comments)
	assert.Equal(t, int64(0), votes)
}
