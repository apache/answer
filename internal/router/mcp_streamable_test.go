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

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

// mcpTestGroupPath mirrors the group internal/base/server/http.go hands to
// RegisterMCPRouter. The group there carries AuthMcpEnable and AuthAPIKey; the
// test leaves them off so the transports are exercised without a database.
const mcpTestGroupPath = "/answer/api/v1"

type mcpJSONRPCResponse struct {
	Result struct {
		ServerInfo struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	} `json:"result"`
}

func newMCPTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	(&AnswerAPIRouter{}).RegisterMCPRouter(engine.Group(mcpTestGroupPath))
	return engine
}

// postMCP sends a JSON-RPC message the way an MCP client does: JSON body, both
// response media types accepted, and the session id on every call after
// initialize.
func postMCP(t *testing.T, engine *gin.Engine, path, sessionID string, message map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal %v: %v", message["method"], err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set(server.HeaderKeySessionID, sessionID)
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)
	return recorder
}

func TestRegisterMCPRouterServesStreamableHTTP(t *testing.T) {
	engine := newMCPTestEngine()
	endpoint := mcpTestGroupPath + "/mcp"

	initialize := postMCP(t, engine, endpoint, "", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "answer-router-test", "version": "0"},
		},
	})
	if initialize.Code != http.StatusOK {
		t.Fatalf("initialize status = %d, want %d: %s", initialize.Code, http.StatusOK, initialize.Body.String())
	}
	sessionID := initialize.Header().Get(server.HeaderKeySessionID)
	if sessionID == "" {
		t.Fatal("streamable HTTP initialize returned no Mcp-Session-Id header")
	}

	var initResponse mcpJSONRPCResponse
	if err := json.Unmarshal(initialize.Body.Bytes(), &initResponse); err != nil {
		t.Fatalf("decode initialize response: %v", err)
	}
	if initResponse.Result.ServerInfo.Name != "Answer Enterprise MCP Server" {
		t.Fatalf("initialize serverInfo.name = %q, want %q", initResponse.Result.ServerInfo.Name, "Answer Enterprise MCP Server")
	}

	initialized := postMCP(t, engine, endpoint, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	if initialized.Code != http.StatusAccepted {
		t.Fatalf("notifications/initialized status = %d, want %d", initialized.Code, http.StatusAccepted)
	}

	list := postMCP(t, engine, endpoint, sessionID, map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	})
	if list.Code != http.StatusOK {
		t.Fatalf("tools/list status = %d, want %d: %s", list.Code, http.StatusOK, list.Body.String())
	}
	var listResponse mcpJSONRPCResponse
	if err := json.Unmarshal(list.Body.Bytes(), &listResponse); err != nil {
		t.Fatalf("decode tools/list response: %v", err)
	}
	got := make(map[string]bool, len(listResponse.Result.Tools))
	for _, tool := range listResponse.Result.Tools {
		got[tool.Name] = true
	}
	for _, want := range []string{
		"get_questions",
		"get_answers_by_question_id",
		"get_comments",
		"get_tags",
		"get_tag_detail",
		"get_user",
	} {
		if !got[want] {
			t.Fatalf("streamable HTTP tools/list is missing %q: %v", want, listResponse.Result.Tools)
		}
	}
}

func TestRegisterMCPRouterKeepsLegacySSETransport(t *testing.T) {
	engine := newMCPTestEngine()

	// The legacy message endpoint answers with its own invalid-params error, so a
	// 400 carrying that message proves the route is still wired to the SSE server
	// rather than to the new handler or a gin 404.
	message := postMCP(t, engine, mcpTestGroupPath+"/mcp/message", "", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "ping",
	})
	if message.Code != http.StatusBadRequest {
		t.Fatalf("legacy message endpoint status = %d, want %d", message.Code, http.StatusBadRequest)
	}
	if !strings.Contains(message.Body.String(), "Missing sessionId") {
		t.Fatalf("legacy message endpoint answered %q, want its own missing session id error", message.Body.String())
	}

	var sseRegistered bool
	for _, route := range engine.Routes() {
		if route.Method == http.MethodGet && route.Path == mcpTestGroupPath+"/mcp/sse" {
			sseRegistered = true
		}
	}
	if !sseRegistered {
		t.Fatal("GET " + mcpTestGroupPath + "/mcp/sse is no longer registered")
	}
}
