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

import (
	"strings"
	"testing"

	"github.com/apache/answer/internal/schema/mcp_tools"
	"github.com/apache/answer/internal/service/embedding"
)

func toolNames(c *AIController, includeSemanticSearch bool) map[string]bool {
	tools := c.buildOpenAITools(mcp_tools.MCPToolsList, includeSemanticSearch)
	names := make(map[string]bool, len(tools))
	for _, t := range tools {
		names[t.Function.Name] = true
	}
	return names
}

func TestBuildOpenAIToolsIncludesSemanticSearch(t *testing.T) {
	c := &AIController{}
	names := toolNames(c, true)
	if !names[semanticSearchToolName] {
		t.Fatalf("semantic_search should be advertised when a vector search plugin is available: %v", names)
	}
	if len(names) != len(mcp_tools.MCPToolsList) {
		t.Fatalf("expect %d tools, got %d", len(mcp_tools.MCPToolsList), len(names))
	}
}

func TestBuildOpenAIToolsExcludesSemanticSearch(t *testing.T) {
	c := &AIController{}
	names := toolNames(c, false)
	if names[semanticSearchToolName] {
		t.Fatalf("semantic_search must not be advertised without a vector search plugin")
	}
	if len(names) != len(mcp_tools.MCPToolsList)-1 {
		t.Fatalf("expect %d tools, got %d", len(mcp_tools.MCPToolsList)-1, len(names))
	}
	if !names["get_questions"] || !names["get_user"] {
		t.Fatalf("other MCP tools must remain advertised: %v", names)
	}
}

func TestStripSemanticSearchLine(t *testing.T) {
	prompt := "You are an assistant.\n- get_questions: search questions\n- semantic_search: search by meaning\n- get_user: search users\n"
	got := stripSemanticSearchLine(prompt)
	if strings.Contains(got, "semantic_search") {
		t.Fatalf("semantic_search line not stripped: %q", got)
	}
	if !strings.Contains(got, "get_questions") || !strings.Contains(got, "get_user") {
		t.Fatalf("unrelated lines were dropped: %q", got)
	}
	if !strings.HasPrefix(got, "You are an assistant.\n") {
		t.Fatalf("leading lines must be kept: %q", got)
	}
}

func TestAdaptPromptToCapabilitiesStripsWhenUnavailable(t *testing.T) {
	// In tests no VectorSearch plugin is registered, so semantic search is
	// unavailable and the prompt must be adapted.
	c := &AIController{mcpController: &MCPController{embeddingService: &embedding.EmbeddingService{}}}
	prompt := "intro\n- semantic_search: search by meaning\noutro\n"
	got := c.adaptPromptToCapabilities(prompt)
	if strings.Contains(got, "semantic_search") {
		t.Fatalf("expected semantic_search line removed: %q", got)
	}
	if !strings.Contains(got, "intro") || !strings.Contains(got, "outro") {
		t.Fatalf("other content must be preserved: %q", got)
	}
}
