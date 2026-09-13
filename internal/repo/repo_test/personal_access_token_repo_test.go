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
	"time"

	"github.com/apache/answer/internal/entity"
	personalaccesstoken "github.com/apache/answer/internal/repo/personal_access_token"
	"github.com/stretchr/testify/require"
)

func TestPersonalAccessTokenRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	repo := personalaccesstoken.NewRepository(testDataSource)
	expiresAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	token := &entity.PersonalAccessToken{
		UserID: "1", Name: "agent", TokenHash: "hash-one", TokenSuffix: "last",
		Scopes: `["question.read"]`, ExpiresAt: expiresAt,
	}

	require.NoError(t, repo.Create(ctx, token))
	require.NotZero(t, token.ID)

	found, exists, err := repo.FindByHash(ctx, "hash-one")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "1", found.UserID)
	require.Equal(t, expiresAt.Unix(), found.ExpiresAt.Unix())

	list, err := repo.ListByUserID(ctx, "1")
	require.NoError(t, err)
	require.Len(t, list, 1)

	revokedAt := time.Now().UTC().Truncate(time.Second)
	revoked, err := repo.Revoke(ctx, "1", token.ID, revokedAt)
	require.NoError(t, err)
	require.True(t, revoked)

	found, exists, err = repo.FindByHash(ctx, "hash-one")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, revokedAt.Unix(), found.RevokedAt.Unix())

	revoked, err = repo.Revoke(ctx, "2", token.ID, revokedAt)
	require.NoError(t, err)
	require.False(t, revoked)

	second := &entity.PersonalAccessToken{
		UserID: "2", Name: "second agent", TokenHash: "hash-two", TokenSuffix: "last",
		Scopes: `["answer.read"]`, ExpiresAt: expiresAt,
	}
	require.NoError(t, repo.Create(ctx, second))
	require.NoError(t, repo.RevokeAllByUserID(ctx, "2", revokedAt))
	found, exists, err = repo.FindByHash(ctx, "hash-two")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, revokedAt.Unix(), found.RevokedAt.Unix())
}
