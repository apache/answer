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
	"strings"
)

type scopeRequirement struct {
	all       []string
	any       []string
	scopeFree bool
}

var routeScopes = map[string]scopeRequirement{
	route(http.MethodGet, "/question/info"):           {all: []string{ScopeQuestionRead}},
	route(http.MethodGet, "/question/page"):           {all: []string{ScopeQuestionRead}},
	route(http.MethodGet, "/question/recommend/page"): {all: []string{ScopeQuestionRead}},
	route(http.MethodGet, "/question/similar/tag"):    {all: []string{ScopeQuestionRead}},
	route(http.MethodGet, "/question/link"):           {all: []string{ScopeQuestionRead}},
	route(http.MethodGet, "/question/similar"):        {any: []string{ScopeQuestionRead, ScopeQuestionCreate}},

	route(http.MethodGet, "/answer/info"): {all: []string{ScopeAnswerRead}},
	route(http.MethodGet, "/answer/page"): {all: []string{ScopeAnswerRead}},

	route(http.MethodGet, "/search"): {all: []string{ScopeQuestionRead, ScopeAnswerRead}},

	route(http.MethodGet, "/question/tags"): {any: []string{ScopeQuestionRead, ScopeQuestionCreate}},
	route(http.MethodGet, "/tags/page"):     {any: []string{ScopeQuestionRead, ScopeQuestionCreate}},
	route(http.MethodGet, "/tag"):           {any: []string{ScopeQuestionRead, ScopeQuestionCreate}},
	route(http.MethodGet, "/tags"):          {any: []string{ScopeQuestionRead, ScopeQuestionCreate}},
	route(http.MethodGet, "/tag/synonyms"):  {any: []string{ScopeQuestionRead, ScopeQuestionCreate}},

	route(http.MethodPost, "/question"):  {all: []string{ScopeQuestionCreate}},
	route(http.MethodPost, "/answer"):    {all: []string{ScopeAnswerCreate}},
	route(http.MethodPost, "/vote/up"):   {all: []string{ScopeVoteWrite}},
	route(http.MethodPost, "/vote/down"): {all: []string{ScopeVoteWrite}},

	route(http.MethodGet, "/personal-access-tokens/current"): {scopeFree: true},
}

func route(method, path string) string {
	return method + " " + path
}

// AuthorizeRoute applies the deny-by-default PAT route policy.
func AuthorizeRoute(method, fullPath string, scopes []string) (allowed bool, required []string) {
	path := apiPath(fullPath)
	requirement, exists := routeScopes[route(method, path)]
	if !exists {
		return false, []string{"unsupported"}
	}
	if requirement.scopeFree {
		return true, nil
	}
	have := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		have[scope] = struct{}{}
	}
	for _, scope := range requirement.all {
		if _, ok := have[scope]; !ok {
			return false, requirement.all
		}
	}
	if len(requirement.any) > 0 {
		for _, scope := range requirement.any {
			if _, ok := have[scope]; ok {
				return true, nil
			}
		}
		return false, requirement.any
	}
	return true, nil
}

func apiPath(fullPath string) string {
	const prefix = "/answer/api/v1"
	if index := strings.Index(fullPath, prefix); index >= 0 {
		return fullPath[index+len(prefix):]
	}
	return fullPath
}
