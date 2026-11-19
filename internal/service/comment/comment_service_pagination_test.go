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

package comment_test

import (
	"context"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	answerrepo "github.com/apache/answer/internal/repo/answer"
	commentrepo "github.com/apache/answer/internal/repo/comment"
	questionrepo "github.com/apache/answer/internal/repo/question"
	tagrepo "github.com/apache/answer/internal/repo/tag_common"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/object_info"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

func newCommentPaginationService(t *testing.T, records ...any) *comment.CommentService {
	t.Helper()
	db, err := data.NewDB(false, &data.Database{Driver: string(schemas.SQLITE), Connection: ":memory:"})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	require.NoError(t, db.Sync2(&entity.Comment{}, &entity.Question{}, &entity.Answer{}, &entity.Tag{}))
	for _, record := range records {
		_, err = db.NoAutoTime().Insert(record)
		require.NoError(t, err)
	}

	dbData := &data.Data{DB: db}
	comments := commentrepo.NewCommentRepo(dbData, nil)
	objects := object_info.NewObjService(
		answerrepo.NewAnswerRepo(dbData, nil, nil, nil),
		questionrepo.NewQuestionRepo(dbData, nil),
		commentrepo.NewCommentCommonRepo(dbData, nil),
		tagrepo.NewTagCommonRepo(dbData, nil), nil,
	)
	return comment.NewCommentService(comments, nil, nil, objects, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

func paginationComment(id, objectID string, createdAt int64) *entity.Comment {
	return &entity.Comment{
		ID: id, ObjectID: objectID, UserID: "303", Status: entity.CommentStatusAvailable,
		CreatedAt: time.Unix(createdAt, 0), ParsedText: "comment " + id,
	}
}

func requirePersonalCommentPage(t *testing.T, service *comment.CommentService,
	req *schema.GetCommentPersonalWithPageReq, count int64, ids ...string) {
	t.Helper()
	page, err := service.GetCommentPersonalWithPage(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, count, page.Count)
	list := page.List.([]*schema.GetCommentPersonalWithPageResp)
	require.NotNil(t, list)
	got := make([]string, 0, len(list))
	for _, item := range list {
		got = append(got, item.CommentID)
	}
	if len(ids) == 0 {
		require.Empty(t, got)
	} else {
		require.Equal(t, ids, got)
	}
}

func TestPersonalCommentPaginationFiltersBeforePaging(t *testing.T) {
	service := newCommentPaginationService(t,
		&entity.Question{ID: "10010000000000001", UserID: "101", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		&entity.Question{ID: "10010000000000002", UserID: "101", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		&entity.Question{ID: "10010000000000003", UserID: "101", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow},
		paginationComment("10070000000000003", "10010000000000001", 3),
		paginationComment("10070000000000002", "10010000000000002", 2),
		paginationComment("10070000000000001", "10010000000000003", 1),
	)
	for _, tc := range []struct {
		name   string
		viewer string
		admin  bool
		count  int64
		pages  [][]string
	}{
		{"anonymous", "", false, 1, [][]string{{"10070000000000001"}, {}}},
		{"comment author", "303", false, 1, [][]string{{"10070000000000001"}, {}}},
		{"question author", "101", false, 3, [][]string{{"10070000000000003", "10070000000000002"}, {"10070000000000001"}}},
		{"moderator", "", true, 3, [][]string{{"10070000000000003", "10070000000000002"}, {"10070000000000001"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for i, ids := range tc.pages {
				requirePersonalCommentPage(t, service, &schema.GetCommentPersonalWithPageReq{
					UserID: "303", LoginUserID: tc.viewer, IsAdminModerator: tc.admin,
					Page: i + 1, PageSize: 2,
				}, tc.count, ids...)
			}
		})
	}
}

func TestPersonalCommentPaginationCountsCommentsOnRepeatedTargets(t *testing.T) {
	otherUser := paginationComment("10070000000000008", "10010000000000003", 8)
	otherUser.UserID = "404"
	deleted := paginationComment("10070000000000007", "10010000000000003", 7)
	deleted.Status = entity.CommentStatusDeleted
	pending := paginationComment("10070000000000006", "10010000000000003", 6)
	pending.Status = entity.CommentStatusPending
	service := newCommentPaginationService(t,
		&entity.Question{ID: "10010000000000001", UserID: "101", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		&entity.Question{ID: "10010000000000003", UserID: "101", Status: entity.QuestionStatusClosed, Show: entity.QuestionShow},
		otherUser, deleted, pending,
		paginationComment("10070000000000005", "10010000000000001", 5),
		paginationComment("10070000000000004", "10010000000000003", 4),
		paginationComment("10070000000000003", "10010000000000001", 3),
		paginationComment("10070000000000002", "10010000000000003", 2),
		paginationComment("10070000000000001", "10010000000000003", 1),
	)
	for i, ids := range [][]string{{"10070000000000004", "10070000000000002"}, {"10070000000000001"}, {}} {
		requirePersonalCommentPage(t, service, &schema.GetCommentPersonalWithPageReq{
			UserID: "303", Page: i + 1, PageSize: 2,
		}, 3, ids...)
	}
}

func TestPersonalCommentPaginationEmptyResults(t *testing.T) {
	service := newCommentPaginationService(t,
		&entity.Question{ID: "10010000000000001", UserID: "101", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		paginationComment("10070000000000003", "10010000000000001", 3),
		paginationComment("10070000000000002", "10010000000000009", 2), // Missing target.
		paginationComment("10070000000000001", "0", 1),                 // Invalid target.
	)
	for _, userID := range []string{"303", "404"} {
		requirePersonalCommentPage(t, service, &schema.GetCommentPersonalWithPageReq{
			UserID: userID, Page: 1, PageSize: 2,
		}, 0)
	}
}

func TestPersonalCommentPaginationChecksAnswersAndParentQuestions(t *testing.T) {
	service := newCommentPaginationService(t,
		&entity.Question{ID: "10010000000000001", UserID: "101", Status: entity.QuestionStatusAvailable, Show: entity.QuestionHide},
		&entity.Question{ID: "10010000000000003", UserID: "101", Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow},
		&entity.Answer{ID: "10020000000000001", QuestionID: "10010000000000001", UserID: "202", Status: entity.AnswerStatusAvailable},
		&entity.Answer{ID: "10020000000000002", QuestionID: "10010000000000003", UserID: "202", Status: entity.AnswerStatusDeleted},
		&entity.Answer{ID: "10020000000000003", QuestionID: "10010000000000001", UserID: "202", Status: entity.AnswerStatusPending},
		paginationComment("10070000000000003", "10020000000000001", 3),
		paginationComment("10070000000000002", "10020000000000002", 2),
		paginationComment("10070000000000001", "10020000000000003", 1),
	)
	for _, tc := range []struct {
		name   string
		viewer string
		admin  bool
		ids    []string
	}{
		{"anonymous", "", false, nil},
		{"answer author cannot bypass hidden question", "202", false, []string{"10070000000000002"}},
		{"question author cannot bypass pending answer", "101", false, []string{"10070000000000003"}},
		{"moderator", "", true, []string{"10070000000000003", "10070000000000002", "10070000000000001"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requirePersonalCommentPage(t, service, &schema.GetCommentPersonalWithPageReq{
				UserID: "303", LoginUserID: tc.viewer, IsAdminModerator: tc.admin, Page: 1, PageSize: 10,
			}, int64(len(tc.ids)), tc.ids...)
		})
	}
}

func TestPersonalCommentPaginationPreservesTagTargets(t *testing.T) {
	service := newCommentPaginationService(t,
		&entity.Tag{ID: "10030000000000001", UserID: "101", SlugName: "public", Status: entity.TagStatusAvailable},
		&entity.Tag{ID: "10030000000000002", UserID: "101", SlugName: "deleted", Status: entity.TagStatusDeleted},
		paginationComment("10070000000000002", "10030000000000002", 2),
		paginationComment("10070000000000001", "10030000000000001", 1),
	)
	requirePersonalCommentPage(t, service, &schema.GetCommentPersonalWithPageReq{
		UserID: "303", Page: 1, PageSize: 1,
	}, 1, "10070000000000001")
	requirePersonalCommentPage(t, service, &schema.GetCommentPersonalWithPageReq{
		UserID: "303", LoginUserID: "101", Page: 1, PageSize: 1,
	}, 2, "10070000000000002")
}
