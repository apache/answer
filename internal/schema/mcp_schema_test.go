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

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewMCPSearchCondPagination(t *testing.T) {
	testCases := []struct {
		name         string
		arguments    map[string]any
		wantPage     int
		wantPageSize int
	}{
		{"defaults", nil, MCPDefaultPage, MCPDefaultPageSize},
		{"explicit page and page size", map[string]any{"page": float64(3), "page_size": float64(20)}, 3, 20},
		{"string arguments", map[string]any{"page": "2", "page_size": "10"}, 2, 10},
		{"non-positive values fall back to defaults", map[string]any{"page": float64(0), "page_size": float64(-1)}, MCPDefaultPage, MCPDefaultPageSize},
		{"page size is capped", map[string]any{"page_size": float64(1000)}, MCPDefaultPage, MCPMaxPageSize},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := mcp.CallToolRequest{}
			request.Params.Arguments = testCase.arguments

			cond := NewMCPSearchCond(request)
			if cond.Page != testCase.wantPage || cond.PageSize != testCase.wantPageSize {
				t.Fatalf("NewMCPSearchCond() pagination = (%d, %d), want (%d, %d)",
					cond.Page, cond.PageSize, testCase.wantPage, testCase.wantPageSize)
			}

			commentCond := NewMCPSearchCommentCond(request)
			if commentCond.Page != testCase.wantPage || commentCond.PageSize != testCase.wantPageSize {
				t.Fatalf("NewMCPSearchCommentCond() pagination = (%d, %d), want (%d, %d)",
					commentCond.Page, commentCond.PageSize, testCase.wantPage, testCase.wantPageSize)
			}
		})
	}
}
