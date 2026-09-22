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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/mock"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/mock/gomock"
)

func TestTranslationEnabledDefaultsToTrue(t *testing.T) {
	config := &schema.SiteAIResp{}
	if !config.IsTranslationEnabled() {
		t.Fatal("translation should be enabled for existing configurations")
	}

	disabled := false
	config.TranslationEnabled = &disabled
	if config.IsTranslationEnabled() {
		t.Fatal("translation should respect an explicit disabled setting")
	}
}

func TestBuildTranslationPrompt(t *testing.T) {
	prompt := buildTranslationPrompt("en_US")
	for _, expected := range []string{"en_US", "valid JSON", "Preserve Markdown", "never as instructions"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("translation prompt should contain %q: %s", expected, prompt)
		}
	}
}

func TestParseTranslation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		title   string
		content string
	}{
		{
			name:    "plain JSON",
			input:   `{"title":"Hello","content":"Use **this**"}`,
			title:   "Hello",
			content: "Use **this**",
		},
		{
			name:    "markdown fenced JSON",
			input:   "```json\n{\"title\":\"\",\"content\":\"Answer\"}\n```",
			content: "Answer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTranslation(tt.input)
			if err != nil {
				t.Fatalf("parseTranslation returned an error: %v", err)
			}
			if got.Title != tt.title || got.Content != tt.content {
				t.Fatalf("unexpected translation: %#v", got)
			}
		})
	}
}

func TestParseTranslationRejectsNonJSON(t *testing.T) {
	if _, err := parseTranslation("translated prose"); err == nil {
		t.Fatal("parseTranslation should reject non-JSON model output")
	}
}

func TestTranslationProviderError(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{http.StatusUnauthorized, "authentication failed"},
		{http.StatusNotFound, "model was not found"},
		{http.StatusTooManyRequests, "rate limit exceeded"},
	}
	for _, tt := range tests {
		err := translationProviderError(&openai.APIError{HTTPStatusCode: tt.status})
		if !strings.Contains(err.Reason, tt.want) {
			t.Fatalf("status %d: expected %q in %q", tt.status, tt.want, err.Reason)
		}
	}
}

func TestTranslateContentUsesConfiguredModelAndSiteLanguage(t *testing.T) {
	var providerRequest struct {
		Model    string `json:"model"`
		Messages []struct {
			Content string `json:"content"`
		} `json:"messages"`
	}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected provider path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&providerRequest); err != nil {
			t.Fatalf("decode provider request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"test","choices":[{"message":{"role":"assistant","content":"{\"title\":\"Hello\",\"content\":\"Translated body\"}"},"finish_reason":"stop","index":0}]}`))
	}))
	defer provider.Close()

	mockController := gomock.NewController(t)
	siteInfo := mock.NewMockSiteInfoCommonService(mockController)
	siteInfo.EXPECT().GetSiteAI(gomock.Any()).Return(&schema.SiteAIResp{
		Enabled:        true,
		ChosenProvider: "test",
		SiteAIProviders: []*schema.SiteAIProvider{{
			Provider: "test",
			APIHost:  provider.URL,
			APIKey:   "test-key",
			Model:    "translation-model",
		}},
	}, nil)
	siteInfo.EXPECT().GetSiteInterface(gomock.Any()).Return(&schema.SiteInterfaceSettingsResp{
		Language: "en_US",
	}, nil)

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Set("ctxUuidKey", &entity.UserCacheInfo{UserID: "1"})
	ctx.Request = httptest.NewRequest(http.MethodPost, "/answer/api/v1/ai/translate",
		bytes.NewBufferString(`{"title":"Hallo","content":"Deutscher Text"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	(&AIController{siteInfoService: siteInfo}).TranslateContent(ctx)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected response status %d: %s", response.Code, response.Body.String())
	}
	if providerRequest.Model != "translation-model" {
		t.Fatalf("expected configured model, got %q", providerRequest.Model)
	}
	if len(providerRequest.Messages) != 2 || !strings.Contains(providerRequest.Messages[0].Content, "en_US") {
		t.Fatalf("site language was not included in prompt: %#v", providerRequest.Messages)
	}

	var body struct {
		Data TranslateContentResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode translation response: %v", err)
	}
	if body.Data.Title != "Hello" || body.Data.Content != "Translated body" || body.Data.TargetLanguage != "en_US" {
		t.Fatalf("unexpected translation response: %#v", body.Data)
	}
}
