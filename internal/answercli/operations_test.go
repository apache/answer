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

package answercli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuestionCreateReadsBodyFromStdin(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "/answer/api/v1/question", request.URL.Path)
		assert.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"code":200,"reason":"base.success","msg":"Success.","data":{"id":"1001"}}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, SaveConfig(configPath, &Config{
		CurrentProfile: "default",
		Profiles:       map[string]Profile{"default": {Server: server.URL, Token: "answer_pat_secret", AllowInsecureHTTP: true}},
	}))
	stdout := &bytes.Buffer{}
	command := NewRootCommand(Options{
		ConfigPath: configPath,
		Stdin:      bytes.NewBufferString("A detailed question body"),
		Stdout:     stdout,
		Stderr:     &bytes.Buffer{},
	})
	command.SetArgs([]string{"question", "create", "--title", "How does this work?", "--tag", "support", "--body-file", "-"})

	require.NoError(t, command.Execute())
	require.Equal(t, "How does this work?", payload["title"])
	require.Equal(t, "A detailed question body", payload["content"])
	tags := payload["tags"].([]any)
	require.Equal(t, "support", tags[0].(map[string]any)["slug_name"])
	require.JSONEq(t, `{"ok":true,"data":{"id":"1001"}}`, stdout.String())
}

func TestCLIExitCodesAreStableByErrorCategory(t *testing.T) {
	require.Equal(t, 2, ExitCode(ErrProfileNotConfigured))
	require.Equal(t, 3, ExitCode(&APIError{HTTPStatus: http.StatusUnauthorized}))
	require.Equal(t, 4, ExitCode(&APIError{HTTPStatus: http.StatusForbidden}))
	require.Equal(t, 5, ExitCode(&APIError{HTTPStatus: http.StatusBadRequest}))
	require.Equal(t, 6, ExitCode(&APIError{OutcomeUnknown: true}))
}

func TestAPIErrorIsNormalizedForCaptcha(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(`{"code":400,"reason":"error.object.captcha_verification_failed","msg":"Captcha wrong.","data":[{"error_field":"captcha_code"}]}`))
	}))
	defer server.Close()

	client, err := NewClient(Profile{Server: server.URL, Token: "answer_pat_secret", AllowInsecureHTTP: true}, nil)
	require.NoError(t, err)
	_, err = client.Do(t.Context(), http.MethodPost, "/answer/api/v1/question", nil, map[string]any{"title": "Question"})
	require.Error(t, err)
	require.Equal(t, "captcha_required", normalizeError(err).Type)
}
