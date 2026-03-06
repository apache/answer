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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/internal/service/role"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mock repos for AuthService ---

type mockAuthRepo struct {
	userCache *entity.UserCacheInfo
	err       error
}

func (m *mockAuthRepo) GetUserCacheInfo(_ context.Context, _ string) (*entity.UserCacheInfo, error) {
	return m.userCache, m.err
}
func (m *mockAuthRepo) SetUserCacheInfo(_ context.Context, _, _ string, _ *entity.UserCacheInfo) error {
	return nil
}
func (m *mockAuthRepo) GetUserVisitCacheInfo(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (m *mockAuthRepo) RemoveUserCacheInfo(_ context.Context, _ string) error      { return nil }
func (m *mockAuthRepo) RemoveUserVisitCacheInfo(_ context.Context, _ string) error { return nil }
func (m *mockAuthRepo) SetUserStatus(_ context.Context, _ string, _ *entity.UserCacheInfo) error {
	return nil
}
func (m *mockAuthRepo) GetUserStatus(_ context.Context, _ string) (*entity.UserCacheInfo, error) {
	return nil, nil
}
func (m *mockAuthRepo) RemoveUserStatus(_ context.Context, _ string) error { return nil }
func (m *mockAuthRepo) GetAdminUserCacheInfo(_ context.Context, _ string) (*entity.UserCacheInfo, error) {
	return nil, nil
}
func (m *mockAuthRepo) SetAdminUserCacheInfo(_ context.Context, _ string, _ *entity.UserCacheInfo) error {
	return nil
}
func (m *mockAuthRepo) RemoveAdminUserCacheInfo(_ context.Context, _ string) error { return nil }
func (m *mockAuthRepo) AddUserTokenMapping(_ context.Context, _, _ string) error   { return nil }
func (m *mockAuthRepo) RemoveUserTokens(_ context.Context, _ string, _ string)     {}

type mockAPIKeyRepo struct {
	key   *entity.APIKey
	exist bool
	err   error
}

func (m *mockAPIKeyRepo) GetAPIKeyList(_ context.Context) ([]*entity.APIKey, error) { return nil, nil }
func (m *mockAPIKeyRepo) GetAPIKey(_ context.Context, _ string) (*entity.APIKey, bool, error) {
	return m.key, m.exist, m.err
}
func (m *mockAPIKeyRepo) UpdateAPIKey(_ context.Context, _ entity.APIKey) error { return nil }
func (m *mockAPIKeyRepo) AddAPIKey(_ context.Context, _ entity.APIKey) error    { return nil }
func (m *mockAPIKeyRepo) DeleteAPIKey(_ context.Context, _ int) error           { return nil }

type mockUserRepo struct {
	user  *entity.User
	exist bool
	err   error
}

func (m *mockUserRepo) AddUser(_ context.Context, _ *entity.User) error                   { return nil }
func (m *mockUserRepo) IncreaseAnswerCount(_ context.Context, _ string, _ int) error       { return nil }
func (m *mockUserRepo) IncreaseQuestionCount(_ context.Context, _ string, _ int) error     { return nil }
func (m *mockUserRepo) UpdateQuestionCount(_ context.Context, _ string, _ int64) error     { return nil }
func (m *mockUserRepo) UpdateAnswerCount(_ context.Context, _ string, _ int) error         { return nil }
func (m *mockUserRepo) UpdateLastLoginDate(_ context.Context, _ string) error              { return nil }
func (m *mockUserRepo) UpdateEmailStatus(_ context.Context, _ string, _ int) error         { return nil }
func (m *mockUserRepo) UpdateNoticeStatus(_ context.Context, _ string, _ int) error        { return nil }
func (m *mockUserRepo) UpdateEmail(_ context.Context, _, _ string) error                   { return nil }
func (m *mockUserRepo) UpdateUserInterface(_ context.Context, _, _, _ string) error        { return nil }
func (m *mockUserRepo) UpdatePass(_ context.Context, _, _ string) error                    { return nil }
func (m *mockUserRepo) UpdateInfo(_ context.Context, _ *entity.User) error                 { return nil }
func (m *mockUserRepo) UpdateUserProfile(_ context.Context, _ *entity.User) error          { return nil }
func (m *mockUserRepo) BatchGetByID(_ context.Context, _ []string) ([]*entity.User, error) { return nil, nil }
func (m *mockUserRepo) GetByUsername(_ context.Context, _ string) (*entity.User, bool, error) {
	return nil, false, nil
}
func (m *mockUserRepo) GetByUsernames(_ context.Context, _ []string) ([]*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*entity.User, bool, error) {
	return nil, false, nil
}
func (m *mockUserRepo) GetUserCount(_ context.Context) (int64, error) { return 0, nil }
func (m *mockUserRepo) SearchUserListByName(_ context.Context, _ string, _ int, _ bool) ([]*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) IsAvatarFileUsed(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) GetByUserID(_ context.Context, _ string) (*entity.User, bool, error) {
	return m.user, m.exist, m.err
}

// mockUserRoleRelRepo implements role.UserRoleRelRepo for testing
type mockUserRoleRelRepo struct {
	roleID int
	exist  bool
}

func (m *mockUserRoleRelRepo) SaveUserRoleRel(_ context.Context, _ string, _ int) error { return nil }
func (m *mockUserRoleRelRepo) GetUserRoleRelList(_ context.Context, _ []string) ([]*entity.UserRoleRel, error) {
	return nil, nil
}
func (m *mockUserRoleRelRepo) GetUserRoleRelListByRoleID(_ context.Context, _ []int) ([]*entity.UserRoleRel, error) {
	return nil, nil
}
func (m *mockUserRoleRelRepo) GetUserRoleRel(_ context.Context, _ string) (*entity.UserRoleRel, bool, error) {
	if !m.exist {
		return nil, false, nil
	}
	return &entity.UserRoleRel{RoleID: m.roleID}, true, nil
}

// --- Helper ---

func newTestMiddleware(
	authRepo *mockAuthRepo,
	apiKeyRepo *mockAPIKeyRepo,
	userRepo *mockUserRepo,
	roleID int,
) *AuthUserMiddleware {
	svc := auth.NewAuthService(authRepo, apiKeyRepo)
	userRoleRelService := role.NewUserRoleRelService(&mockUserRoleRelRepo{roleID: roleID, exist: true}, nil)
	return NewAuthUserMiddleware(svc, nil, userRepo, userRoleRelService)
}

func performRequest(mw gin.HandlerFunc, method, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(mw)
	engine.Handle(method, path, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	req, _ := http.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer test-token")
	engine.ServeHTTP(w, req)
	return w
}

func performRequestNoToken(mw gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(w)
	engine.Use(mw)
	engine.Handle("GET", "/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)
	return w
}

// --- Tests ---

func TestAuthSessionOrAPIKey_ValidSession(t *testing.T) {
	m := newTestMiddleware(
		&mockAuthRepo{userCache: &entity.UserCacheInfo{
			UserID:      "100",
			UserStatus:  entity.UserStatusAvailable,
			EmailStatus: entity.EmailStatusAvailable,
			RoleID:      1,
		}},
		&mockAPIKeyRepo{exist: false},
		&mockUserRepo{exist: false},
		1,
	)
	w := performRequest(m.AuthSessionOrAPIKey(), "GET", "/test")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthSessionOrAPIKey_InvalidSessionFallbackValidAPIKey(t *testing.T) {
	m := newTestMiddleware(
		&mockAuthRepo{userCache: nil}, // session fails
		&mockAPIKeyRepo{
			key:   &entity.APIKey{AccessKey: "sk_test", Scope: "read-write", UserID: "200"},
			exist: true,
		},
		&mockUserRepo{
			user:  &entity.User{ID: "200", Status: entity.UserStatusAvailable, MailStatus: entity.EmailStatusAvailable},
			exist: true,
		},
		1,
	)
	w := performRequest(m.AuthSessionOrAPIKey(), "GET", "/answer/api/v1/question")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthSessionOrAPIKey_BothFail(t *testing.T) {
	m := newTestMiddleware(
		&mockAuthRepo{userCache: nil},
		&mockAPIKeyRepo{exist: false},
		&mockUserRepo{exist: false},
		1,
	)
	w := performRequest(m.AuthSessionOrAPIKey(), "GET", "/answer/api/v1/question")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthSessionOrAPIKey_NoToken(t *testing.T) {
	m := newTestMiddleware(
		&mockAuthRepo{userCache: nil},
		&mockAPIKeyRepo{exist: false},
		&mockUserRepo{exist: false},
		1,
	)
	w := performRequestNoToken(m.AuthSessionOrAPIKey())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthSessionOrAPIKey_ReadOnlyKeyPostRequest(t *testing.T) {
	m := newTestMiddleware(
		&mockAuthRepo{userCache: nil}, // session fails
		&mockAPIKeyRepo{
			key:   &entity.APIKey{AccessKey: "sk_ro", Scope: "read-only", UserID: "300"},
			exist: true,
		},
		&mockUserRepo{
			user:  &entity.User{ID: "300", Status: entity.UserStatusAvailable, MailStatus: entity.EmailStatusAvailable},
			exist: true,
		},
		1,
	)
	w := performRequest(m.AuthSessionOrAPIKey(), "POST", "/answer/api/v1/question")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthSessionOrAPIKey_APIKeyBlockedOnNonWhitelistedRoute(t *testing.T) {
	m := newTestMiddleware(
		&mockAuthRepo{userCache: nil},
		&mockAPIKeyRepo{
			key:   &entity.APIKey{AccessKey: "sk_test", Scope: "read-write", UserID: "400"},
			exist: true,
		},
		&mockUserRepo{
			user:  &entity.User{ID: "400", Status: entity.UserStatusAvailable, MailStatus: entity.EmailStatusAvailable},
			exist: true,
		},
		1,
	)
	w := performRequest(m.AuthSessionOrAPIKey(), "PUT", "/answer/api/v1/user/password")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
