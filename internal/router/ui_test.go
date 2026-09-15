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

package router

import "testing"

func TestIsMissingStaticAsset(t *testing.T) {
	cases := []struct {
		name, urlPath, base string
		want                bool
	}{
		{"chunk of a previous build", "/static/js/6487.02e56745.chunk.js", "", true},
		{"with base url", "/answer/static/css/main.css", "/answer", true},
		{"spa route", "/users/alice", "", false},
		{"spa route with base url", "/answer/questions", "/answer", false},
		{"static under a different base", "/static/js/main.js", "/answer", false},
		{"root", "/", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isMissingStaticAsset(c.urlPath, c.base); got != c.want {
				t.Fatalf("isMissingStaticAsset(%q, %q) = %v, want %v", c.urlPath, c.base, got, c.want)
			}
		})
	}
}
