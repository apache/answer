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

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	pat "github.com/apache/answer/internal/service/personal_access_token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPersonalAccessTokenRequestAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	repo := &middlewarePATRepository{}
	service := pat.NewService(repo,
		pat.WithClock(func() time.Time { return now }),
		pat.WithSecretGenerator(func() (string, error) { return "0123456789abcdefghijklmnopqrstuvwxyzAB", nil }),
	)
	created, err := service.Create(context.Background(), pat.CreateInput{
		UserID: "42", Name: "agent", Scopes: []string{pat.ScopeQuestionRead}, ExpiresAt: now.Add(time.Hour),
	})
	require.NoError(t, err)

	authorizer := newPATRequestAuthorizer(service, middlewarePATUserLoader{}, middlewarePATSecurityLoader{enabled: true})
	router := gin.New()
	router.GET("/answer/api/v1/question/info", func(ctx *gin.Context) {
		if handled, user := authorizer.Resolve(ctx); handled {
			if user != nil {
				ctx.Status(http.StatusNoContent)
			}
			return
		}
		ctx.Status(http.StatusNoContent)
	})
	router.GET("/answer/api/v1/answer/info", func(ctx *gin.Context) {
		if handled, user := authorizer.Resolve(ctx); handled {
			if user != nil {
				ctx.Status(http.StatusNoContent)
			}
			return
		}
		ctx.Status(http.StatusNoContent)
	})

	t.Run("valid scope establishes owner", func(t *testing.T) {
		response := performPATRequest(router, "/answer/api/v1/question/info", "Bearer "+created.Token)
		require.Equal(t, http.StatusNoContent, response.Code)
	})

	t.Run("missing scope is forbidden", func(t *testing.T) {
		response := performPATRequest(router, "/answer/api/v1/answer/info", "Bearer "+created.Token)
		require.Equal(t, http.StatusForbidden, response.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
		require.Equal(t, "error.personal_access_token.insufficient_scope", body["reason"])
	})

	t.Run("bare PAT is rejected", func(t *testing.T) {
		response := performPATRequest(router, "/answer/api/v1/question/info", created.Token)
		require.Equal(t, http.StatusUnauthorized, response.Code)
	})
}

func performPATRequest(router http.Handler, path, authorization string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Authorization", authorization)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

type middlewarePATRepository struct {
	token *entity.PersonalAccessToken
}

func (r *middlewarePATRepository) Create(_ context.Context, token *entity.PersonalAccessToken) error {
	copy := *token
	copy.ID = 1
	token.ID = 1
	r.token = &copy
	return nil
}

func (r *middlewarePATRepository) FindByHash(_ context.Context, hash string) (*entity.PersonalAccessToken, bool, error) {
	if r.token == nil || r.token.TokenHash != hash {
		return nil, false, nil
	}
	copy := *r.token
	return &copy, true, nil
}

func (*middlewarePATRepository) ListByUserID(context.Context, string) ([]*entity.PersonalAccessToken, error) {
	return nil, nil
}

func (*middlewarePATRepository) Revoke(context.Context, string, int64, time.Time) (bool, error) {
	return false, nil
}

func (*middlewarePATRepository) RevokeAllByUserID(context.Context, string, time.Time) error {
	return nil
}

type middlewarePATUserLoader struct{}

func (middlewarePATUserLoader) GetUserCacheInfoByUserID(context.Context, string) (*entity.UserCacheInfo, error) {
	return &entity.UserCacheInfo{
		UserID: "42", UserStatus: entity.UserStatusAvailable, EmailStatus: entity.EmailStatusAvailable, RoleID: 1,
	}, nil
}

type middlewarePATSecurityLoader struct {
	enabled bool
}

func (s middlewarePATSecurityLoader) GetSiteSecurity(context.Context) (*schema.SiteSecurityResp, error) {
	return &schema.SiteSecurityResp{PersonalAccessTokensEnabled: s.enabled}, nil
}
