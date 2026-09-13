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

package personal_access_token

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoutePolicyIsDenyByDefault(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		scopes []string
		allow  bool
	}{
		{name: "question read", method: http.MethodGet, path: "/answer/api/v1/question/info", scopes: []string{ScopeQuestionRead}, allow: true},
		{name: "question read missing", method: http.MethodGet, path: "/answer/api/v1/question/info", scopes: []string{ScopeAnswerRead}},
		{name: "search requires both reads", method: http.MethodGet, path: "/answer/api/v1/search", scopes: []string{ScopeQuestionRead}},
		{name: "search with both reads", method: http.MethodGet, path: "/answer/api/v1/search", scopes: []string{ScopeQuestionRead, ScopeAnswerRead}, allow: true},
		{name: "tag lookup supports creation", method: http.MethodGet, path: "/answer/api/v1/question/tags", scopes: []string{ScopeQuestionCreate}, allow: true},
		{name: "create question", method: http.MethodPost, path: "/answer/api/v1/question", scopes: []string{ScopeQuestionCreate}, allow: true},
		{name: "update question denied", method: http.MethodPut, path: "/answer/api/v1/question", scopes: []string{ScopeQuestionCreate}},
		{name: "notification denied", method: http.MethodGet, path: "/answer/api/v1/notification/page", scopes: []string{ScopeQuestionRead, ScopeAnswerRead}},
		{name: "self inspection", method: http.MethodGet, path: "/answer/api/v1/personal-access-tokens/current", allow: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, required := AuthorizeRoute(tt.method, tt.path, tt.scopes)
			require.Equal(t, tt.allow, allowed)
			if !tt.allow {
				require.NotEmpty(t, required)
			}
		})
	}
}
