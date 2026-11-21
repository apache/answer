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

package mcp_tools

import (
	"slices"
	"testing"

	"github.com/apache/answer/internal/schema"
)

func TestNewQuestionDetailTool(t *testing.T) {
	tool := NewQuestionDetailTool()

	if tool.Name != "get_question_detail" {
		t.Fatalf("tool name = %q, want %q", tool.Name, "get_question_detail")
	}
	if _, ok := tool.InputSchema.Properties[schema.MCPSearchCondQuestionID]; !ok {
		t.Fatalf("%s must be an input property: %v", schema.MCPSearchCondQuestionID, tool.InputSchema.Properties)
	}

	if !slices.Contains(tool.InputSchema.Required, schema.MCPSearchCondQuestionID) {
		t.Fatalf("%s must be required: %v", schema.MCPSearchCondQuestionID, tool.InputSchema.Required)
	}
}
