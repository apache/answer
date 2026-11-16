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
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	authrepo "github.com/apache/answer/internal/repo/auth"
	authservice "github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/gin-gonic/gin"
)

type visitAuthTestSiteInfoRepo struct{}

func (visitAuthTestSiteInfoRepo) SaveByType(context.Context, string, *entity.SiteInfo) error {
	return nil
}

func (visitAuthTestSiteInfoRepo) GetByType(context.Context, string, ...bool) (*entity.SiteInfo, bool, error) {
	return &entity.SiteInfo{Content: `{"login_required":true}`}, true, nil
}

func (visitAuthTestSiteInfoRepo) IsBrandingFileUsed(context.Context, string) (bool, error) {
	return false, nil
}

func TestVisitAuthRejectsRevokedAndSuspendedSessions(t *testing.T) {
	ctx := context.Background()
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}
	t.Cleanup(cleanup)

	repo := authrepo.NewAuthRepo(&data.Data{Cache: cache})
	service := authservice.NewAuthService(repo, nil)
	accessToken, visitToken, err := service.SetUserCacheInfo(ctx, &entity.UserCacheInfo{
		UserID:      "visit-auth-user",
		UserStatus:  entity.UserStatusAvailable,
		EmailStatus: entity.EmailStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	filePath := filepath.Join(t.TempDir(), "private.txt")
	if err := os.WriteFile(filePath, []byte("private upload"), 0o600); err != nil {
		t.Fatalf("create private upload: %v", err)
	}
	serve := func(token string) *httptest.ResponseRecorder {
		gin.SetMode(gin.TestMode)
		engine := gin.New()
		siteInfo := siteinfo_common.NewSiteInfoCommonService(visitAuthTestSiteInfoRepo{})
		authMiddleware := NewAuthUserMiddleware(service, siteInfo)
		engine.Use(authMiddleware.VisitAuth())
		engine.StaticFile("/uploads/post/private.txt", filePath)
		req := httptest.NewRequest(http.MethodGet, "/uploads/post/private.txt", nil)
		req.AddCookie(&http.Cookie{Name: constant.UserVisitCookiesCacheKey, Value: token})
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, req)
		return recorder
	}

	if response := serve(visitToken); response.Code != http.StatusOK {
		t.Fatalf("fresh visit token returned %d, want %d", response.Code, http.StatusOK)
	}

	service.RemoveUserAllTokens(ctx, "visit-auth-user")
	if userInfo, err := service.GetUserCacheInfo(ctx, accessToken); err != nil || userInfo != nil {
		t.Fatalf("revoked access token remained valid: userInfo=%v err=%v", userInfo, err)
	}
	if response := serve(visitToken); response.Code != http.StatusFound || response.Header().Get("Location") != "/403" {
		t.Fatalf("revoked visit token returned status=%d location=%q, want 302 /403", response.Code, response.Header().Get("Location"))
	}

	_, suspendedVisitToken, err := service.SetUserCacheInfo(ctx, &entity.UserCacheInfo{
		UserID:      "suspended-visit-auth-user",
		UserStatus:  entity.UserStatusAvailable,
		EmailStatus: entity.EmailStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create suspended-user session: %v", err)
	}
	if err := service.SetUserStatus(ctx, &entity.UserCacheInfo{
		UserID:      "suspended-visit-auth-user",
		UserStatus:  entity.UserStatusSuspended,
		EmailStatus: entity.EmailStatusAvailable,
	}); err != nil {
		t.Fatalf("suspend user: %v", err)
	}
	if response := serve(suspendedVisitToken); response.Code != http.StatusFound || response.Header().Get("Location") != "/403" {
		t.Fatalf("suspended visit token returned status=%d location=%q, want 302 /403", response.Code, response.Header().Get("Location"))
	}
}

func TestRemoveTokensExceptCurrentUserRevokesOnlyOtherVisitTokens(t *testing.T) {
	ctx := context.Background()
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}
	t.Cleanup(cleanup)

	service := authservice.NewAuthService(authrepo.NewAuthRepo(&data.Data{Cache: cache}), nil)
	currentAccessToken, currentVisitToken, err := service.SetUserCacheInfo(ctx, &entity.UserCacheInfo{
		UserID:      "multi-session-user",
		UserStatus:  entity.UserStatusAvailable,
		EmailStatus: entity.EmailStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create current session: %v", err)
	}
	_, otherVisitToken, err := service.SetUserCacheInfo(ctx, &entity.UserCacheInfo{
		UserID:      "multi-session-user",
		UserStatus:  entity.UserStatusAvailable,
		EmailStatus: entity.EmailStatusAvailable,
	})
	if err != nil {
		t.Fatalf("create other session: %v", err)
	}

	service.RemoveTokensExceptCurrentUser(ctx, "multi-session-user", currentAccessToken)
	if !service.CheckUserVisitToken(ctx, currentVisitToken) {
		t.Fatal("current visit token was revoked")
	}
	if service.CheckUserVisitToken(ctx, otherVisitToken) {
		t.Fatal("other visit token remained valid")
	}
}
