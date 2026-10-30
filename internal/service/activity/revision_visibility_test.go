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
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/permission"
)

type timelineRevisionPermissionCheckerStub struct {
	can     []bool
	err     error
	calls   int
	userID  string
	actions []string
}

func (s *timelineRevisionPermissionCheckerStub) CheckOperationPermissions(_ context.Context, userID string, actions []string) ([]bool, error) {
	s.calls++
	s.userID = userID
	s.actions = actions
	return s.can, s.err
}

func TestIsTimelineRevisionPublic(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{name: "normal", status: entity.RevisionNormalStatus, want: true},
		{name: "approved", status: entity.RevisionReviewPassStatus, want: true},
		{name: "pending", status: entity.RevisionUnreviewedStatus, want: false},
		{name: "rejected", status: entity.RevisionReviewRejectStatus, want: false},
		{name: "unknown status", status: 99, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimelineRevisionPublic(tt.status); got != tt.want {
				t.Errorf("isTimelineRevisionPublic(%d) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestTimelineRevisionReviewPermission(t *testing.T) {
	tests := []struct {
		name       string
		objectType int
		want       string
		wantOK     bool
	}{
		{name: "question", objectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType], want: permission.QuestionAudit, wantOK: true},
		{name: "answer", objectType: constant.ObjectTypeStrMapping[constant.AnswerObjectType], want: permission.AnswerAudit, wantOK: true},
		{name: "tag", objectType: constant.ObjectTypeStrMapping[constant.TagObjectType], want: permission.TagAudit, wantOK: true},
		{name: "unsupported object", objectType: constant.ObjectTypeStrMapping[constant.CommentObjectType], wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := timelineRevisionReviewPermission(tt.objectType)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("timelineRevisionReviewPermission(%d) = (%q, %v), want (%q, %v)", tt.objectType, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestEnsureTimelineRevisionStatusVisible(t *testing.T) {
	tests := []struct {
		name       string
		revision   *entity.Revision
		checker    *timelineRevisionPermissionCheckerStub
		wantErr    bool
		wantCalls  int
		wantAction string
	}{
		{
			name:      "normal revision is public",
			revision:  &entity.Revision{Status: entity.RevisionNormalStatus},
			checker:   &timelineRevisionPermissionCheckerStub{},
			wantCalls: 0,
		},
		{
			name:       "pending question is denied to a normal user",
			revision:   &entity.Revision{Status: entity.RevisionUnreviewedStatus, ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType]},
			checker:    &timelineRevisionPermissionCheckerStub{can: []bool{false}},
			wantErr:    true,
			wantCalls:  1,
			wantAction: permission.QuestionAudit,
		},
		{
			name:       "rejected answer is denied to a normal user",
			revision:   &entity.Revision{Status: entity.RevisionReviewRejectStatus, ObjectType: constant.ObjectTypeStrMapping[constant.AnswerObjectType]},
			checker:    &timelineRevisionPermissionCheckerStub{can: []bool{false}},
			wantErr:    true,
			wantCalls:  1,
			wantAction: permission.AnswerAudit,
		},
		{
			name:       "authorized tag reviewer can access pending revision",
			revision:   &entity.Revision{Status: entity.RevisionUnreviewedStatus, ObjectType: constant.ObjectTypeStrMapping[constant.TagObjectType]},
			checker:    &timelineRevisionPermissionCheckerStub{can: []bool{true}},
			wantCalls:  1,
			wantAction: permission.TagAudit,
		},
		{
			name:      "unknown revision type fails closed",
			revision:  &entity.Revision{Status: entity.RevisionUnreviewedStatus, ObjectType: constant.ObjectTypeStrMapping[constant.CommentObjectType]},
			checker:   &timelineRevisionPermissionCheckerStub{can: []bool{true}},
			wantErr:   true,
			wantCalls: 0,
		},
		{
			name:       "permission lookup failure is returned",
			revision:   &entity.Revision{Status: entity.RevisionUnreviewedStatus, ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType]},
			checker:    &timelineRevisionPermissionCheckerStub{err: errors.New("permission lookup failed")},
			wantErr:    true,
			wantCalls:  1,
			wantAction: permission.QuestionAudit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &ActivityService{rankService: tt.checker}
			err := service.ensureTimelineRevisionStatusVisible(context.Background(), tt.revision, "reader")
			if (err != nil) != tt.wantErr {
				t.Fatalf("ensureTimelineRevisionStatusVisible error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checker.calls != tt.wantCalls {
				t.Fatalf("permission checks = %d, want %d", tt.checker.calls, tt.wantCalls)
			}
			if tt.wantAction != "" {
				if tt.checker.userID != "reader" {
					t.Errorf("permission user ID = %q, want reader", tt.checker.userID)
				}
				if len(tt.checker.actions) != 1 || tt.checker.actions[0] != tt.wantAction {
					t.Errorf("permission actions = %v, want [%s]", tt.checker.actions, tt.wantAction)
				}
			}
		})
	}
}
