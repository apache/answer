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

package search_parser

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/mock"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type usernameRepo struct {
	usercommon.UserRepo
	users map[string]*entity.User
}

func (r *usernameRepo) GetByUsername(_ context.Context, username string) (*entity.User, bool, error) {
	user, ok := r.users[username]
	return user, ok, nil
}

func TestParseStructureUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	siteInfo := mock.NewMockSiteInfoCommonService(ctrl)
	siteInfo.EXPECT().FormatAvatar(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&schema.AvatarInfo{}).AnyTimes()
	repo := &usernameRepo{users: map[string]*entity.User{
		"melissa": {ID: "2", Username: "melissa"},
	}}
	sp := NewSearchParser(nil, usercommon.NewUserCommon(repo, nil, nil, siteInfo))

	tests := []struct {
		query     string
		wantUser  string
		wantWords []string
	}{
		{query: "user:me", wantUser: "1"},
		{query: "user:me service", wantUser: "1", wantWords: []string{"service"}},
		{query: "user:melissa", wantUser: "2"},
		{query: "user:melissa service", wantUser: "2", wantWords: []string{"service"}},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			cond := sp.ParseStructure(context.Background(), &schema.SearchDTO{Query: tt.query, UserID: "1"})
			assert.Equal(t, tt.wantUser, cond.UserID)
			assert.ElementsMatch(t, tt.wantWords, cond.Words)
		})
	}
}
