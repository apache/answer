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
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuthStatusUsesConfiguredBearerTokenAndWritesJSON(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authorization = request.Header.Get("Authorization")
		require.Equal(t, "/answer/api/v1/personal-access-tokens/current", request.URL.Path)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"code":200,"reason":"base.success","msg":"Success.","data":{"user":{"id":"42","username":"alice","display_name":"Alice"},"token":{"name":"agent","scopes":["question.read"]}}}`))
	}))
	defer server.Close()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, SaveConfig(configPath, &Config{
		CurrentProfile: "default",
		Profiles:       map[string]Profile{"default": {Server: server.URL, Token: "answer_pat_secret", AllowInsecureHTTP: true}},
	}))
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	command := NewRootCommand(Options{ConfigPath: configPath, Stdout: stdout, Stderr: stderr})
	command.SetArgs([]string{"auth", "status"})

	require.NoError(t, command.Execute())
	require.Equal(t, "Bearer answer_pat_secret", authorization)
	require.JSONEq(t, `{"ok":true,"data":{"user":{"id":"42","username":"alice","display_name":"Alice"},"token":{"name":"agent","scopes":["question.read"]}}}`, stdout.String())
	require.Empty(t, stderr.String())
}

func TestNonLocalHTTPRequiresExplicitOptIn(t *testing.T) {
	err := ValidateServerURL("http://answer.example.com", false)
	require.ErrorIs(t, err, ErrInsecureHTTP)
	require.NoError(t, ValidateServerURL("http://127.0.0.1:9080", false))
	require.NoError(t, ValidateServerURL("https://answer.example.com", false))
}
