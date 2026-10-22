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

package activity

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/contexts"
	"xorm.io/xorm/schemas"
)

type voteResultActivityRepo struct {
	activity_common.ActivityRepo
	err error
}

func (r *voteResultActivityRepo) GetActivityTypeByObjectType(_ context.Context, objectType, action string) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if action == constant.ActVoteUp {
		return 1, nil
	}
	if objectType == constant.CommentObjectType {
		return 0, errors.New("comments have no downvote activity type")
	}
	return 2, nil
}

type voteResultHook struct {
	before func(*contexts.ContextHook) (context.Context, error)
}

func (h voteResultHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	return h.before(c)
}

func (h voteResultHook) AfterProcess(_ *contexts.ContextHook) error { return nil }

func newVoteResultRepo(t *testing.T, connection string) *VoteRepo {
	t.Helper()
	db, err := data.NewDB(false, &data.Database{Driver: string(schemas.SQLITE), Connection: connection})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Sync2(&entity.Question{}, &entity.Answer{}, &entity.Comment{}, &entity.Activity{}))
	return &VoteRepo{data: &data.Data{DB: db}, activityRepo: &voteResultActivityRepo{}}
}

func voteResultTarget(objectType string, count int) any {
	switch objectType {
	case constant.QuestionObjectType:
		return &entity.Question{
			ID: "10010000000000001", UserID: "101", LastEditUserID: "0",
			AcceptedAnswerID: "0", LastAnswerID: "0", RevisionID: "0", VoteCount: count,
		}
	case constant.AnswerObjectType:
		return &entity.Answer{
			ID: "10010000000000001", UserID: "101", QuestionID: "0",
			LastEditUserID: "0", RevisionID: "0", VoteCount: count,
		}
	default:
		return &entity.Comment{
			ID: "10010000000000001", UserID: "101", ObjectID: "0", QuestionID: "0", VoteCount: count,
		}
	}
}

func storedVoteResult(t *testing.T, repo *VoteRepo, objectType string) int64 {
	t.Helper()
	var result struct {
		Count int64 `xorm:"vote_count"`
	}
	exist, err := repo.data.DB.Table(voteResultTarget(objectType, 0)).
		Select("vote_count").Where("id = ?", "10010000000000001").Get(&result)
	require.NoError(t, err)
	require.True(t, exist)
	return result.Count
}

func TestGetAndSaveVoteResultRejectsStaleTotals(t *testing.T) {
	for _, objectType := range []string{constant.QuestionObjectType, constant.AnswerObjectType, constant.CommentObjectType} {
		for _, tc := range []struct {
			name        string
			cancel      bool
			newVoteType int
			up, down    int64
		}{
			{"new upvote", false, 1, 2, 0},
			{"cancel upvote", true, 0, 0, 0},
			{"new downvote", false, 2, 1, 1},
			{"switch to downvote", true, 2, 0, 1},
		} {
			if objectType == constant.CommentObjectType && tc.newVoteType == 2 {
				continue
			}
			t.Run(objectType+"/"+tc.name, func(t *testing.T) {
				connection := filepath.Join(t.TempDir(), "votes.db")
				slow := newVoteResultRepo(t, connection)
				fast := newVoteResultRepo(t, connection)
				_, err := slow.data.DB.Insert(voteResultTarget(objectType, 0))
				require.NoError(t, err)
				firstVote := &entity.Activity{ObjectID: "10010000000000001", OriginalObjectID: "0", UserID: "101", ActivityType: 1}
				_, err = slow.data.DB.Insert(firstVote)
				require.NoError(t, err)

				// Pause the older request just before its first target UPDATE.
				// The original implementation has already calculated its stale
				// counts here. A lock-before-count implementation has not.
				paused, resume := make(chan struct{}), make(chan struct{})
				unblock := sync.OnceFunc(func() { close(resume) })
				defer unblock()
				var once sync.Once
				slow.data.DB.AddHook(voteResultHook{before: func(c *contexts.ContextHook) (context.Context, error) {
					if strings.HasPrefix(strings.ToUpper(c.SQL), "UPDATE") && strings.Contains(c.SQL, "vote_count") {
						once.Do(func() { close(paused); <-resume })
					}
					return c.Ctx, nil
				}})
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				type result struct {
					up, down int64
					err      error
				}
				done := make(chan result, 1)
				go func() {
					up, down, err := slow.GetAndSaveVoteResult(ctx, "10010000000000001", objectType)
					done <- result{up, down, err}
				}()
				select {
				case <-paused:
				case <-ctx.Done():
					t.Fatal("older request did not reach the target update")
				}

				want := tc.up - tc.down
				if tc.cancel {
					_, err = fast.data.DB.ID(firstVote.ID).Cols("cancelled").
						Update(&entity.Activity{Cancelled: entity.ActivityCancelled})
					require.NoError(t, err)
				}
				if tc.newVoteType != 0 {
					_, err = fast.data.DB.Insert(&entity.Activity{
						ObjectID: "10010000000000001", OriginalObjectID: "0", UserID: "202", ActivityType: tc.newVoteType,
					})
					require.NoError(t, err)
				}
				_, _, err = fast.GetAndSaveVoteResult(ctx, "10010000000000001", objectType)
				require.NoError(t, err)
				require.Equal(t, want, storedVoteResult(t, fast, objectType))
				unblock()
				select {
				case got := <-done:
					require.NoError(t, got.err)
					require.Equal(t, want, storedVoteResult(t, fast, objectType))
					require.Equal(t, tc.up, got.up)
					require.Equal(t, tc.down, got.down)
				case <-ctx.Done():
					t.Fatal("older request did not complete")
				}
			})
		}
	}
}

func TestGetAndSaveVoteResultCountsActiveVotes(t *testing.T) {
	for _, objectType := range []string{constant.QuestionObjectType, constant.AnswerObjectType, constant.CommentObjectType} {
		t.Run(objectType, func(t *testing.T) {
			repo := newVoteResultRepo(t, filepath.Join(t.TempDir(), "votes.db"))
			_, err := repo.data.DB.Insert(voteResultTarget(objectType, 9))
			require.NoError(t, err)
			votes := []*entity.Activity{
				{ObjectID: "10010000000000001", OriginalObjectID: "0", UserID: "101", ActivityType: 1},
				{ObjectID: "10010000000000001", OriginalObjectID: "0", UserID: "102", ActivityType: 1, Cancelled: entity.ActivityCancelled},
				{ObjectID: "10010000000000001", OriginalObjectID: "0", UserID: "103", ActivityType: 7},
				{ObjectID: "10010000000000002", OriginalObjectID: "0", UserID: "104", ActivityType: 1},
			}
			if objectType != constant.CommentObjectType {
				votes = append(votes,
					&entity.Activity{ObjectID: "10010000000000001", OriginalObjectID: "0", UserID: "105", ActivityType: 2},
					&entity.Activity{ObjectID: "10010000000000001", OriginalObjectID: "0", UserID: "106", ActivityType: 2},
				)
			}
			_, err = repo.data.DB.Insert(votes)
			require.NoError(t, err)
			up, down, err := repo.GetAndSaveVoteResult(context.Background(), "10010000000000001", objectType)
			require.NoError(t, err)
			require.Equal(t, int64(1), up)
			if objectType == constant.CommentObjectType {
				require.Zero(t, down)
			} else {
				require.Equal(t, int64(2), down)
			}
			require.Equal(t, up-down, storedVoteResult(t, repo, objectType))

			_, err = repo.data.DB.Where("object_id = ?", "10010000000000001").
				Cols("cancelled").Update(&entity.Activity{Cancelled: entity.ActivityCancelled})
			require.NoError(t, err)
			up, down, err = repo.GetAndSaveVoteResult(context.Background(), "10010000000000001", objectType)
			require.NoError(t, err)
			require.Zero(t, up)
			require.Zero(t, down)
			require.Zero(t, storedVoteResult(t, repo, objectType))
		})
	}
}

func TestGetAndSaveVoteResultDoesNotOverwriteOnCountError(t *testing.T) {
	repo := newVoteResultRepo(t, filepath.Join(t.TempDir(), "votes.db"))
	_, err := repo.data.DB.Insert(voteResultTarget(constant.QuestionObjectType, 7))
	require.NoError(t, err)
	repo.data.DB.AddHook(voteResultHook{before: func(c *contexts.ContextHook) (context.Context, error) {
		if strings.Contains(strings.ToUpper(c.SQL), "COUNT(") {
			return c.Ctx, errors.New("count query failed")
		}
		return c.Ctx, nil
	}})
	_, _, err = repo.GetAndSaveVoteResult(context.Background(), "10010000000000001", constant.QuestionObjectType)
	require.Error(t, err)
	require.Equal(t, int64(7), storedVoteResult(t, repo, constant.QuestionObjectType))
}

func TestGetAndSaveVoteResultDoesNotOverwriteOnActivityTypeError(t *testing.T) {
	repo := newVoteResultRepo(t, filepath.Join(t.TempDir(), "votes.db"))
	_, err := repo.data.DB.Insert(voteResultTarget(constant.QuestionObjectType, 7))
	require.NoError(t, err)
	lookupErr := errors.New("activity configuration unavailable")
	repo.activityRepo = &voteResultActivityRepo{err: lookupErr}
	_, _, err = repo.GetAndSaveVoteResult(context.Background(), "10010000000000001", constant.QuestionObjectType)
	require.ErrorIs(t, err, lookupErr)
	require.Equal(t, int64(7), storedVoteResult(t, repo, constant.QuestionObjectType))
}
