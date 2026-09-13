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

package personal_access_token_test

import (
	"context"
	"testing"
	"time"

	"github.com/apache/answer/internal/entity"
	pat "github.com/apache/answer/internal/service/personal_access_token"
	"github.com/stretchr/testify/require"
)

func TestUserCanCreateAndAuthenticatePersonalAccessToken(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	service := pat.NewService(repo,
		pat.WithClock(func() time.Time { return now }),
		pat.WithSecretGenerator(func() (string, error) { return "0123456789abcdefghijklmnopqrstuvwxyzAB", nil }),
	)

	created, err := service.Create(context.Background(), pat.CreateInput{
		UserID:    "42",
		Name:      "coding agent",
		Scopes:    []string{pat.ScopeVoteWrite, pat.ScopeQuestionRead, pat.ScopeQuestionRead},
		ExpiresAt: now.Add(30 * 24 * time.Hour),
	})
	require.NoError(t, err)
	require.Equal(t, "answer_pat_0123456789abcdefghijklmnopqrstuvwxyzAB", created.Token)
	require.Equal(t, "yzAB", created.Info.TokenSuffix)
	require.Equal(t, []string{pat.ScopeQuestionRead, pat.ScopeVoteWrite}, created.Info.Scopes)

	authenticated, err := service.Authenticate(context.Background(), created.Token)
	require.NoError(t, err)
	require.Equal(t, "42", authenticated.UserID)
	require.Equal(t, []string{pat.ScopeQuestionRead, pat.ScopeVoteWrite}, authenticated.Scopes)

	stored := repo.byID[created.Info.ID]
	require.NotContains(t, stored.TokenHash, created.Token)
	require.NotContains(t, stored.Scopes, created.Token)
}

func TestRevokedAndExpiredPersonalAccessTokensCannotAuthenticate(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	repo := newMemoryRepository()
	service := pat.NewService(repo,
		pat.WithClock(func() time.Time { return now }),
		pat.WithSecretGenerator(func() (string, error) { return "0123456789abcdefghijklmnopqrstuvwxyzAB", nil }),
	)

	created, err := service.Create(context.Background(), pat.CreateInput{
		UserID: "42", Name: "coding agent", Scopes: []string{pat.ScopeQuestionRead},
		ExpiresAt: now.Add(time.Hour),
	})
	require.NoError(t, err)
	require.NoError(t, service.Revoke(context.Background(), "42", created.Info.ID))

	_, err = service.Authenticate(context.Background(), created.Token)
	require.ErrorIs(t, err, pat.ErrInvalidToken)

	listed, err := service.List(context.Background(), "42")
	require.NoError(t, err)
	require.Len(t, listed, 1)
	require.Equal(t, pat.StatusRevoked, listed[0].Status)

	now = now.Add(2 * time.Hour)
	repo.byID[created.Info.ID].RevokedAt = time.Time{}
	listed, err = service.List(context.Background(), "42")
	require.NoError(t, err)
	require.Equal(t, pat.StatusExpired, listed[0].Status)
}

func TestPersonalAccessTokenScopeMustBeKnownAndExpiryIsBounded(t *testing.T) {
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	service := pat.NewService(newMemoryRepository(), pat.WithClock(func() time.Time { return now }))

	_, err := service.Create(context.Background(), pat.CreateInput{
		UserID: "42", Name: "invalid", Scopes: []string{"admin.access"}, ExpiresAt: now.Add(time.Hour),
	})
	require.ErrorIs(t, err, pat.ErrInvalidScope)

	_, err = service.Create(context.Background(), pat.CreateInput{
		UserID: "42", Name: "too long", Scopes: []string{pat.ScopeQuestionRead}, ExpiresAt: now.Add(366 * 24 * time.Hour),
	})
	require.ErrorIs(t, err, pat.ErrInvalidExpiry)
}

type memoryRepository struct {
	nextID int64
	byID   map[int64]*entity.PersonalAccessToken
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{nextID: 1, byID: make(map[int64]*entity.PersonalAccessToken)}
}

func (r *memoryRepository) Create(_ context.Context, token *entity.PersonalAccessToken) error {
	copy := *token
	copy.ID = r.nextID
	r.nextID++
	r.byID[copy.ID] = &copy
	token.ID = copy.ID
	return nil
}

func (r *memoryRepository) FindByHash(_ context.Context, hash string) (*entity.PersonalAccessToken, bool, error) {
	for _, token := range r.byID {
		if token.TokenHash == hash {
			copy := *token
			return &copy, true, nil
		}
	}
	return nil, false, nil
}

func (r *memoryRepository) ListByUserID(_ context.Context, userID string) ([]*entity.PersonalAccessToken, error) {
	var result []*entity.PersonalAccessToken
	for _, token := range r.byID {
		if token.UserID == userID {
			copy := *token
			result = append(result, &copy)
		}
	}
	return result, nil
}

func (r *memoryRepository) Revoke(_ context.Context, userID string, id int64, revokedAt time.Time) (bool, error) {
	token, ok := r.byID[id]
	if !ok || token.UserID != userID {
		return false, nil
	}
	if token.RevokedAt.IsZero() {
		token.RevokedAt = revokedAt
	}
	return true, nil
}

func (r *memoryRepository) RevokeAllByUserID(_ context.Context, userID string, revokedAt time.Time) error {
	for _, token := range r.byID {
		if token.UserID == userID && token.RevokedAt.IsZero() {
			token.RevokedAt = revokedAt
		}
	}
	return nil
}
