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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const currentUserID = "current-login-user-id"

// fakeUserRepo answers username lookups from a map. Methods that parseUserID
// never reaches are left to the embedded nil interface.
type fakeUserRepo struct {
	usercommon.UserRepo
	users map[string]*entity.User
}

func (r *fakeUserRepo) GetByUsername(_ context.Context, username string) (*entity.User, bool, error) {
	user, exist := r.users[username]
	if !exist {
		// The real repository always returns an allocated entity.
		return &entity.User{}, false, nil
	}
	return user, exist, nil
}

func newTestSearchParser(t *testing.T, users map[string]*entity.User) *SearchParser {
	t.Helper()

	ctrl := gomock.NewController(t)
	siteInfoService := mock.NewMockSiteInfoCommonService(ctrl)
	siteInfoService.EXPECT().FormatAvatar(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&schema.AvatarInfo{}).AnyTimes()

	userCommon := usercommon.NewUserCommon(&fakeUserRepo{users: users}, nil, nil, siteInfoService)
	return NewSearchParser(nil, userCommon)
}

func TestSearchParserParseUserID(t *testing.T) {
	users := map[string]*entity.User{
		"melissa": {ID: "melissa-id", Username: "melissa"},
		"bob":     {ID: "bob-id", Username: "bob"},
		"megan":   {ID: "megan-id", Username: "megan"},
	}

	testCases := []struct {
		name          string
		query         string
		currentUserID string
		wantUserID    string
		wantQuery     string
	}{
		{
			name:          "user:me resolves to the current login user",
			query:         "user:me",
			currentUserID: currentUserID,
			wantUserID:    currentUserID,
			wantQuery:     "",
		},
		{
			name:          "user:me keeps the rest of the query as keywords",
			query:         "user:me parking",
			currentUserID: currentUserID,
			wantUserID:    currentUserID,
			wantQuery:     "parking",
		},
		{
			name:          "username starting with me is looked up instead of the current user",
			query:         "user:melissa",
			currentUserID: currentUserID,
			wantUserID:    "melissa-id",
			wantQuery:     "",
		},
		{
			name:          "username starting with me keeps its keywords",
			query:         "user:melissa service",
			currentUserID: currentUserID,
			wantUserID:    "melissa-id",
			wantQuery:     "service",
		},
		{
			name:          "another me-prefixed username is not truncated",
			query:         "user:megan",
			currentUserID: currentUserID,
			wantUserID:    "megan-id",
			wantQuery:     "",
		},
		{
			name:          "ordinary username still resolves",
			query:         "user:bob",
			currentUserID: currentUserID,
			wantUserID:    "bob-id",
			wantQuery:     "",
		},
		{
			name:          "unknown me-prefixed username is left in the query",
			query:         "user:meadow",
			currentUserID: currentUserID,
			wantUserID:    "",
			wantQuery:     "user:meadow",
		},
		{
			name:          "query without a user filter is untouched",
			query:         "parking service",
			currentUserID: currentUserID,
			wantUserID:    "",
			wantQuery:     "parking service",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			searchParser := newTestSearchParser(t, users)
			query := testCase.query
			userID := searchParser.parseUserID(context.Background(), &query, testCase.currentUserID)
			assert.Equal(t, testCase.wantUserID, userID)
			assert.Equal(t, testCase.wantQuery, query)
		})
	}
}

func TestSearchParserParseStructureResolvesMePrefixedUsername(t *testing.T) {
	users := map[string]*entity.User{
		"melissa": {ID: "melissa-id", Username: "melissa"},
	}

	searchParser := newTestSearchParser(t, users)
	cond := searchParser.ParseStructure(context.Background(), &schema.SearchDTO{
		Query:  "user:melissa",
		UserID: currentUserID,
	})

	require.NotNil(t, cond)
	assert.Equal(t, "melissa-id", cond.UserID)
	assert.Empty(t, cond.Words)
}
