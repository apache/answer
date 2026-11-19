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

	s.AddTool(mcp_tools.NewQuestionsTool(), a.mcpController.MCPQuestionsHandler())
	s.AddTool(mcp_tools.NewAnswersTool(), a.mcpController.MCPAnswersHandler())
	s.AddTool(mcp_tools.NewCommentsTool(), a.mcpController.MCPCommentsHandler())
	s.AddTool(mcp_tools.NewTagsTool(), a.mcpController.MCPTagsHandler())
	s.AddTool(mcp_tools.NewTagDetailTool(), a.mcpController.MCPTagDetailsHandler())
	s.AddTool(mcp_tools.NewUserTool(), a.mcpController.MCPUserDetailsHandler())

	sseServer := newMCPSSEServer(s)
	r.GET("/mcp/sse", gin.WrapH(sseServer.SSEHandler()))
	r.POST("/mcp/message", gin.WrapH(sseServer.MessageHandler()))
}
