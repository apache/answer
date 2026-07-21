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

import "testing"

func TestTagDomainRedirectPath(t *testing.T) {
	tests := []struct {
		name        string
		requestPath string
		tagSlug     string
		want        string
	}{
		{name: "root", requestPath: "/", tagSlug: "go", want: "/tags/go"},
		{name: "base path", requestPath: "/answer", tagSlug: "go", want: "/answer/tags/go"},
		{name: "base path trailing slash", requestPath: "/answer/", tagSlug: "go", want: "/answer/tags/go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tagDomainRedirectPath(tt.requestPath, tt.tagSlug); got != tt.want {
				t.Fatalf("tagDomainRedirectPath(%q, %q) = %q, want %q", tt.requestPath, tt.tagSlug, got, tt.want)
			}
		})
	}
}
