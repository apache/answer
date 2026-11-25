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

package badge_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	answerrepo "github.com/apache/answer/internal/repo/answer"
	badgeawardrepo "github.com/apache/answer/internal/repo/badge_award"
	questionrepo "github.com/apache/answer/internal/repo/question"
	userrepo "github.com/apache/answer/internal/repo/user"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/badge"
	"github.com/apache/answer/internal/service/mock"
	"github.com/apache/answer/internal/service/object_info"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"go.uber.org/mock/gomock"
)

func TestBadgeAwardListRedactsHiddenParentQuestionWithSQLite(t *testing.T) {
	ctx := context.Background()
	db, err := data.NewDB(false, &data.Database{Driver: "sqlite", Connection: ":memory:"})
	if err != nil {
		t.Fatalf("create SQLite database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Sync2(new(entity.BadgeAward), new(entity.Question), new(entity.Answer), new(entity.User)); err != nil {
		t.Fatalf("create tables: %v", err)
	}

	const (
		questionID = "10010000000000001"
		answerID   = "10020000000000001"
		awardID    = "10090000000000001"
		userID     = "10000000000000001"
	)
	if _, err := db.Context(ctx).Insert(&entity.Question{
		ID:         questionID,
		UserID:     "10000000000000002",
		Title:      "hidden parent question",
		ParsedText: "hidden body",
		Status:     entity.QuestionStatusAvailable,
		Show:       entity.QuestionHide,
	}); err != nil {
		t.Fatalf("insert hidden question: %v", err)
	}
	if _, err := db.Context(ctx).Insert(&entity.Answer{
		ID:         answerID,
		QuestionID: questionID,
		UserID:     userID,
		ParsedText: "answer body",
		Status:     entity.AnswerStatusAvailable,
	}); err != nil {
		t.Fatalf("insert answer: %v", err)
	}
	if _, err := db.Context(ctx).Insert(&entity.BadgeAward{
		ID:             awardID,
		BadgeID:        "10080000000000001",
		UserID:         userID,
		AwardKey:       answerID,
		IsBadgeDeleted: entity.IsBadgeNotDeleted,
	}); err != nil {
		t.Fatalf("insert badge award: %v", err)
	}
	if _, err := db.Context(ctx).Insert(&entity.User{ID: userID, Username: "award-user"}); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	dataSource := &data.Data{DB: db}
	ctrl := gomock.NewController(t)
	siteInfo := mock.NewMockSiteInfoCommonService(ctrl)
	siteInfo.EXPECT().FormatAvatar(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&schema.AvatarInfo{}).AnyTimes()
	users := usercommon.NewUserCommon(userrepo.NewUserRepo(dataSource), nil, nil, siteInfo)
	objects := object_info.NewObjService(
		answerrepo.NewAnswerRepo(dataSource, nil, nil, nil),
		questionrepo.NewQuestionRepo(dataSource, nil),
		nil,
		nil,
		nil,
	)
	service := badge.NewBadgeAwardService(
		badgeawardrepo.NewBadgeAwardRepo(dataSource, nil),
		nil,
		users,
		objects,
		nil,
	)

	awards, total, err := service.GetBadgeAwardList(ctx, &schema.GetBadgeAwardWithPageReq{
		BadgeID:  "10080000000000001",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("get badge award list: %v", err)
	}
	if total != 1 || len(awards) != 1 {
		t.Fatalf("awards = %d (total %d), want 1", len(awards), total)
	}
	award := awards[0]
	if award.ObjectID != "" || award.QuestionID != "" || award.AnswerID != "" || award.UrlTitle != "" {
		t.Fatalf("hidden parent question leaked through badge award: %+v", award)
	}
}
