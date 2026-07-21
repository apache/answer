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

package tag_common

import "testing"

func TestNormalizeTagDomain(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain", input: "Tag.Example.COM", want: "tag.example.com"},
		{name: "trailing dot", input: "tag.example.com.", want: "tag.example.com"},
		{name: "port", input: "tag.example.com:8080", want: "tag.example.com"},
		{name: "port and trailing dot", input: "tag.example.com.:443", want: "tag.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeTagDomain(tt.input); got != tt.want {
				t.Fatalf("normalizeTagDomain(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
