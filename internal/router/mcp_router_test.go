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
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/mark3labs/mcp-go/server"
)

// TestMCPSSEEndpointConfig guards the MCP endpoint configuration against the
// "see"/"sse" typo and against the path drifting between the SSE server config,
// the gin mount (via the shared constants) and the advertised client URL.
func TestMCPSSEEndpointConfig(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "server name", got: constant.MCPServerName, want: "Apache Answer MCP Server"},
		{name: "server version", got: constant.MCPServerVersion, want: "1.0.0"},
		{name: "SSE path", got: constant.MCPSSEPath, want: "/answer/api/v1/mcp/sse"},
		{name: "message path", got: constant.MCPMessagePath, want: "/answer/api/v1/mcp/message"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	// CompleteSsePath is consumed by URL parsing / health checks / future path
	// validation in mcp-go. It must resolve to the same path the router mounts
	// and that clients are told to connect to.
	sseServer := newMCPSSEServer(server.NewMCPServer(constant.MCPServerName, constant.MCPServerVersion))
	if got, want := sseServer.CompleteSsePath(), "/answer/api/v1/mcp/sse"; got != want {
		t.Errorf("CompleteSsePath() = %q, want %q", got, want)
	}
	if got, want := sseServer.CompleteMessagePath(), "/answer/api/v1/mcp/message"; got != want {
		t.Errorf("CompleteMessagePath() = %q, want %q", got, want)
	}
}
