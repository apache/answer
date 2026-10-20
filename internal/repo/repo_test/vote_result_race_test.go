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
	"fmt"
	"sync"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/activity"
	activityCommonRepo "github.com/apache/answer/internal/repo/activity_common"
	configRepo "github.com/apache/answer/internal/repo/config"
	"github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/pkg/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func newVoteRepoForTest() content.VoteRepo {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	configService := config.NewConfigService(configRepo.NewConfigRepo(testDataSource))
	activityRepo := activityCommonRepo.NewActivityRepo(testDataSource, uniqueIDRepo, configService)
	// GetAndSaveVoteResult only needs the data handle and the activity repo.
	return activity.NewVoteRepo(testDataSource, activityRepo, nil, nil)
}

func voteActivityTypeID(t *testing.T, ctx context.Context, key string) int {
	cfg := &entity.Config{}
	exist, err := testDataSource.DB.Context(ctx).Where(builder.Eq{"key": key}).Get(cfg)
	require.NoError(t, err)
	require.True(t, exist, "config %s must exist in the seeded test data", key)
	return cfg.ID
}

func addVoteActivity(ctx context.Context, objectID, voterUserID string, activityType int) error {
	_, err := testDataSource.DB.Context(ctx).Insert(&entity.Activity{
		ObjectID:         objectID,
		OriginalObjectID: objectID,
		UserID:           "9501",
		TriggerUserID:    converter.StringToInt64(voterUserID),
		ActivityType:     activityType,
		Cancelled:        entity.ActivityAvailable,
	})
	return err
}

func addVoteQuestion(t *testing.T, ctx context.Context, title string) *entity.Question {
	questionRepo := question.NewQuestionRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	q := &entity.Question{
		UserID:       "9502",
		Title:        title,
		OriginalText: title,
		ParsedText:   title,
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	}
	require.NoError(t, questionRepo.AddQuestion(ctx, q))
	return q
}

func TestGetAndSaveVoteResultCountsActiveVotes(t *testing.T) {
	ctx := context.Background()
	voteRepo := newVoteRepoForTest()
	upType := voteActivityTypeID(t, ctx, "question.vote_up")
	downType := voteActivityTypeID(t, ctx, "question.vote_down")
	q := addVoteQuestion(t, ctx, "vote result count question")

	for i := range 3 {
		require.NoError(t, addVoteActivity(ctx, q.ID, fmt.Sprintf("9510%d", i), upType))
	}
	require.NoError(t, addVoteActivity(ctx, q.ID, "9519", downType))

	up, down, err := voteRepo.GetAndSaveVoteResult(ctx, q.ID, "question")
	require.NoError(t, err)
	assert.Equal(t, int64(3), up)
	assert.Equal(t, int64(1), down)

	stored := &entity.Question{}
	exist, err := testDataSource.DB.Context(ctx).ID(q.ID).Get(stored)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, 2, stored.VoteCount)
}

// Reproduces https://github.com/apache/answer/issues/1646: recalculating the
// total and saving it as an absolute value lets an earlier request overwrite a
// newer result once vote activities commit between the count and the write.
// A single shared connection keeps sqlite free of locking errors while the
// statements of the concurrent calls still interleave freely. One round can
// get lucky and write the correct total last, so the burst repeats over
// several fresh questions; every round must end with the exact total.
func TestGetAndSaveVoteResultConcurrentVotes(t *testing.T) {
	ctx := context.Background()
	voteRepo := newVoteRepoForTest()
	upType := voteActivityTypeID(t, ctx, "question.vote_up")

	testDataSource.DB.SetMaxOpenConns(1)
	defer testDataSource.DB.SetMaxOpenConns(0)

	const rounds = 10
	const voters = 32
	for round := range rounds {
		q := addVoteQuestion(t, ctx, fmt.Sprintf("vote result race question %d", round))
		var wg sync.WaitGroup
		for i := range voters {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				assert.NoError(t, addVoteActivity(ctx, q.ID, fmt.Sprintf("952%02d%02d", round, i), upType))
				_, _, err := voteRepo.GetAndSaveVoteResult(ctx, q.ID, "question")
				assert.NoError(t, err)
			}(i)
		}
		wg.Wait()

		stored := &entity.Question{}
		exist, err := testDataSource.DB.Context(ctx).ID(q.ID).Get(stored)
		require.NoError(t, err)
		require.True(t, exist)
		assert.Equal(t, voters, stored.VoteCount,
			"round %d: stored vote_count must equal the number of active vote activities", round)
	}
}
