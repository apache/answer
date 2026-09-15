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
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const thinkingTestBody = `{"model":"deepseek-v4-flash","messages":[{"role":"user","content":"hi"}],"stream":true}`

// newCaptureServer returns a server that records the last JSON body it
// received and the helper that posts the test request to it.
func newCaptureServer() (*httptest.Server, *map[string]any, func(t *testing.T, client *http.Client, path string)) {
	got := &map[string]any{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, got)
		w.WriteHeader(http.StatusOK)
	}))
	post := func(t *testing.T, client *http.Client, path string) {
		t.Helper()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
			srv.URL+path, strings.NewReader(thinkingTestBody))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
	}
	return srv, got, post
}

func TestThinkingParamForHost(t *testing.T) {
	if !reflect.DeepEqual(thinkingParamForHost("https://dashscope.aliyuncs.com/compatible-mode/v1"),
		map[string]any{"enable_thinking": true}) {
		t.Fatalf("DashScope host should get enable_thinking")
	}
	if !reflect.DeepEqual(thinkingParamForHost("https://Api.DeepSeek.com"),
		map[string]any{"thinking": map[string]any{"type": "enabled"}}) {
		t.Fatalf("DeepSeek host should get the thinking object")
	}
	for _, host := range []string{
		"https://generativelanguage.googleapis.com/v1beta/openai",
		"https://example.com",
		"",
	} {
		if thinkingParamForHost(host) != nil {
			t.Fatalf("host %q should get no provider-specific parameter", host)
		}
	}
}

func TestThinkingTransportInjectsDashScopeFlag(t *testing.T) {
	srv, got, post := newCaptureServer()
	defer srv.Close()

	client := newThinkingHTTPClient(thinkingParamForHost("https://dashscope.aliyuncs.com/compatible-mode/v1"))
	post(t, client, "/v1/chat/completions")

	if (*got)["enable_thinking"] != true {
		t.Fatalf("enable_thinking not injected: %v", (*got)["enable_thinking"])
	}
	if (*got)["model"] != "deepseek-v4-flash" {
		t.Fatalf("original fields lost: %v", *got)
	}
}

func TestThinkingTransportInjectsDeepSeekObject(t *testing.T) {
	srv, got, post := newCaptureServer()
	defer srv.Close()

	client := newThinkingHTTPClient(thinkingParamForHost("https://api.deepseek.com"))
	post(t, client, "/v1/chat/completions")

	thinking, ok := (*got)["thinking"].(map[string]any)
	if !ok || thinking["type"] != "enabled" {
		t.Fatalf("thinking object not injected: %v", (*got)["thinking"])
	}
	if _, exists := (*got)["enable_thinking"]; exists {
		t.Fatalf("gateway-specific flag must not leak to other providers: %v", *got)
	}
}

func TestThinkingTransportNoParamLeavesBodyUntouched(t *testing.T) {
	srv, got, post := newCaptureServer()
	defer srv.Close()

	client := newThinkingHTTPClient(nil)
	post(t, client, "/v1/chat/completions")

	if _, exists := (*got)["enable_thinking"]; exists {
		t.Fatalf("no provider-specific field must be injected: %v", *got)
	}
	if (*got)["model"] != "deepseek-v4-flash" {
		t.Fatalf("body was modified: %v", *got)
	}
}

func TestThinkingTransportLeavesOtherPaths(t *testing.T) {
	srv, _, post := newCaptureServer()
	defer srv.Close()

	client := newThinkingHTTPClient(thinkingParamForHost("https://dashscope.aliyuncs.com/compatible-mode/v1"))
	post(t, client, "/v1/models")
}
