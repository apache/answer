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

package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/service/auth"
	"github.com/gin-gonic/gin"
)

func TestRedirectToAuthLandingDoesNotExposeAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	service := auth.NewLoginHandoffService(&data.Data{Cache: cache})
	const accessToken = "session-token-that-must-not-appear-in-location"
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/callback", nil)

	if err := redirectToAuthLanding(ctx, "https://answer.example", service, accessToken); err != nil {
		t.Fatal(err)
	}
	location := w.Header().Get("Location")
	if strings.Contains(location, accessToken) || strings.Contains(location, "?") {
		t.Fatalf("redirect exposes access token in URL: %q", location)
	}
	if !strings.HasPrefix(location, "https://answer.example/users/auth-landing#code=") {
		t.Fatalf("unexpected redirect location: %q", location)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", w.Header().Get("Cache-Control"))
	}
}

func TestExchangeLoginHandoffReturnsTokenOnceAndDoesNotCacheResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	service := auth.NewLoginHandoffService(&data.Data{Cache: cache})
	const accessToken = "session-token-only-in-post-body"
	code, err := service.Create(context.Background(), accessToken)
	if err != nil {
		t.Fatal(err)
	}
	controller := &UserController{loginHandoffService: service}
	router := gin.New()
	router.POST("/handoff", controller.ExchangeLoginHandoff)

	req := httptest.NewRequest(http.MethodPost, "/handoff", strings.NewReader(`{"code":"`+code+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", w.Header().Get("Cache-Control"))
	}
	var response struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.AccessToken != accessToken {
		t.Fatalf("access token = %q, want %q", response.Data.AccessToken, accessToken)
	}

	req = httptest.NewRequest(http.MethodPost, "/handoff", strings.NewReader(`{"code":"`+code+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("second exchange status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestAuthLandingHandoffFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)

	service := auth.NewLoginHandoffService(&data.Data{Cache: cache})
	const accessToken = "session-token-returned-only-by-exchange"
	redirectWriter := httptest.NewRecorder()
	redirectContext, _ := gin.CreateTestContext(redirectWriter)
	redirectContext.Request = httptest.NewRequest(http.MethodGet, "/callback", nil)
	if err := redirectToAuthLanding(redirectContext, "https://answer.example", service, accessToken); err != nil {
		t.Fatal(err)
	}
	location, err := url.Parse(redirectWriter.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if code := location.Query().Get("code"); code != "" {
		t.Fatalf("handoff code must not be in query: %q", location.String())
	}
	fragment, err := url.ParseQuery(location.Fragment)
	if err != nil {
		t.Fatal(err)
	}
	code := fragment.Get("code")
	if code == "" {
		t.Fatalf("handoff code missing from fragment: %q", location.String())
	}

	controller := &UserController{loginHandoffService: service}
	router := gin.New()
	router.POST("/handoff", controller.ExchangeLoginHandoff)
	req := httptest.NewRequest(http.MethodPost, "/handoff", strings.NewReader(`{"code":"`+code+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), accessToken) {
		t.Fatalf("handoff exchange failed: status=%d body=%s", w.Code, w.Body.String())
	}
}
