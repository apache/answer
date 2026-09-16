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

package activity_log

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/eventqueue"
)

type fakeRepo struct{ added chan *entity.ActivityLog }

func (f *fakeRepo) Add(_ context.Context, e *entity.ActivityLog) error    { f.added <- e; return nil }
func (f *fakeRepo) ExistsSourceRef(context.Context, string) (bool, error) { return false, nil }
func (f *fakeRepo) Page(context.Context, *Query, int, int) ([]*entity.ActivityLog, int64, error) {
	return nil, 0, nil
}
func (f *fakeRepo) Iterate(context.Context, *Query, int, func([]*entity.ActivityLog) bool) error {
	return nil
}
func (f *fakeRepo) ActionCounts(context.Context, *Query) (map[string]int64, error) { return nil, nil }
func (f *fakeRepo) SearchUserIDs(context.Context, string, int) ([]string, error)   { return nil, nil }
func (f *fakeRepo) DeletePageViewsBefore(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type fakeCommentRepo struct{ reply string }

func (f *fakeCommentRepo) GetComment(_ context.Context, id string) (*entity.Comment, bool, error) {
	c := &entity.Comment{ID: id, UserID: "7"}
	if id == "42" { // the comment being replied to
		c.UserID = "3"
	} else if f.reply != "" {
		c.ReplyCommentID = sql.NullInt64{Int64: 42, Valid: true}
	}
	return c, true, nil
}
func (f *fakeCommentRepo) GetCommentWithoutStatus(ctx context.Context, id string) (*entity.Comment, bool, error) {
	return f.GetComment(ctx, id)
}
func (f *fakeCommentRepo) GetCommentCount(context.Context) (int64, error)         { return 0, nil }
func (f *fakeCommentRepo) RemoveAllUserComment(context.Context, string) error     { return nil }
func (f *fakeCommentRepo) UpdateCommentStatus(context.Context, string, int) error { return nil }

func newTestService(t *testing.T, reply string) (*ActivityLogService, *fakeRepo) {
	t.Helper()
	repo := &fakeRepo{added: make(chan *entity.ActivityLog, 10)}
	s := NewActivityLogService(repo, &fakeCommentRepo{reply: reply}, eventqueue.NewService())
	return s, repo
}

func next(t *testing.T, repo *fakeRepo) *entity.ActivityLog {
	t.Helper()
	select {
	case e := <-repo.added:
		return e
	case <-time.After(2 * time.Second):
		t.Fatal("no log entry written")
		return nil
	}
}

func TestHandleEvent_VoteDirection(t *testing.T) {
	s, repo := newTestService(t, "")
	msg := schema.NewEvent(constant.EventAnswerVote, "7").TID("10020000000000055").AID("10020000000000055", "9")
	msg.AddExtra("vote_direction", "down").AddExtra("vote_up_amount", "0").AddExtra("vote_down_amount", "1")
	if err := s.handleEvent(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	e := next(t, repo)
	if e.Action != "answer.vote_down" || e.UserID != "7" || e.TargetUserID != "9" || e.ObjectID != "10020000000000055" || e.AnswerID != "10020000000000055" {
		t.Fatalf("unexpected entry %+v", e)
	}
	d := DecodeDetail(e.Detail)
	if d["vote_down_amount"] != "1" || d["vote_direction"] != nil {
		t.Fatalf("unexpected detail %v", d)
	}
}

func TestHandleEvent_CommentReply(t *testing.T) {
	s, repo := newTestService(t, "42")
	msg := schema.NewEvent(constant.EventCommentCreate, "7").TID("10070000000000300").CID("10070000000000300", "7").QID("10010000000000100", "3")
	if err := s.handleEvent(context.Background(), msg); err != nil {
		t.Fatal(err)
	}
	e := next(t, repo)
	if e.Action != ActionCommentReply || e.QuestionID != "10010000000000100" || e.TargetUserID != "3" {
		t.Fatalf("unexpected entry %+v", e)
	}
	if DecodeDetail(e.Detail)["reply_to"] != "42" {
		t.Fatalf("reply_to missing: %s", e.Detail)
	}
}

func TestHandleEvent_SelfTargetCleared(t *testing.T) {
	s, repo := newTestService(t, "")
	msg := schema.NewEvent(constant.EventQuestionCreate, "3").TID("10010000000000100").QID("10010000000000100", "3")
	_ = s.handleEvent(context.Background(), msg)
	e := next(t, repo)
	if e.Action != "question.create" || e.TargetUserID != "0" || e.ObjectType != "question" {
		t.Fatalf("unexpected entry %+v", e)
	}
}

func TestLogPageView_Dedup(t *testing.T) {
	s, repo := newTestService(t, "")
	if !s.LogPageView(context.Background(), "5", "/questions/100/slug", "T") {
		t.Fatal("first view should be recorded")
	}
	if s.LogPageView(context.Background(), "5", "/questions/100/slug", "T") {
		t.Fatal("repeat within window should be skipped")
	}
	e := next(t, repo)
	if e.Action != ActionPageView || e.QuestionID != "100" || e.ObjectType != "question" {
		t.Fatalf("unexpected entry %+v", e)
	}
}
