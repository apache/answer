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
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/controller"
	"github.com/apache/answer/internal/schema/mcp_tools"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

// newMCPSSEServer builds the SSE server from the shared MCP endpoint constants
// so the configured endpoints cannot drift from the paths mounted below.
func newMCPSSEServer(s *server.MCPServer) *server.SSEServer {
	return server.NewSSEServer(s,
		server.WithSSEEndpoint(constant.MCPSSEPath),
		server.WithMessageEndpoint(constant.MCPMessagePath),
	)
}

func (a *AnswerAPIRouter) RegisterMCPRouter(r *gin.RouterGroup) {
	s := server.NewMCPServer(constant.MCPServerName, constant.MCPServerVersion)

	addMCPTools(s, a.mcpController, a.mcpController.SemanticSearchAvailable())

	sseServer := newMCPSSEServer(s)
	r.GET("/mcp/sse", gin.WrapH(sseServer.SSEHandler()))
	r.POST("/mcp/message", gin.WrapH(sseServer.MessageHandler()))
}

// addMCPTools registers the tools the public MCP server exposes. semantic_search
// is registered only when a VectorSearch plugin is enabled, so an external client
// is never offered a capability that always fails, matching the tool list the AI
// chat builds.
func addMCPTools(s *server.MCPServer, c *controller.MCPController, semanticSearchAvailable bool) {
	s.AddTool(mcp_tools.NewQuestionsTool(), c.MCPQuestionsHandler())
	s.AddTool(mcp_tools.NewQuestionDetailTool(), c.MCPQuestionDetailHandler())
	s.AddTool(mcp_tools.NewAnswersTool(), c.MCPAnswersHandler())
	s.AddTool(mcp_tools.NewCommentsTool(), c.MCPCommentsHandler())
	s.AddTool(mcp_tools.NewTagsTool(), c.MCPTagsHandler())
	s.AddTool(mcp_tools.NewTagDetailTool(), c.MCPTagDetailsHandler())
	s.AddTool(mcp_tools.NewUserTool(), c.MCPUserDetailsHandler())
	if semanticSearchAvailable {
		s.AddTool(mcp_tools.NewSemanticSearchTool(), c.MCPSemanticSearchHandler())
	}
}
