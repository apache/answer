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

package schema

import "testing"

func TestNormalizeTagDomain(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty", input: "", want: ""},
		{name: "normalizes case and trailing dot", input: "Tag.Example.COM.", want: "tag.example.com"},
		{name: "allows localhost", input: "localhost", want: "localhost"},
		{name: "rejects url", input: "https://tag.example.com", wantErr: true},
		{name: "rejects invalid label", input: "tag..example.com", wantErr: true},
		{name: "rejects underscore", input: "tag_name.example.com", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTagDomain(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeTagDomain(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("normalizeTagDomain(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
