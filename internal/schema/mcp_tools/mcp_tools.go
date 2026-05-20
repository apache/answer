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
	"github.com/apache/answer/internal/schema"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	MCPToolsList = []mcp.Tool{
		NewQuestionsTool(),
		NewAnswersTool(),
		NewCommentsTool(),
		NewTagsTool(),
		NewTagDetailTool(),
		NewUserTool(),
		NewSemanticSearchTool(),
	}
)

func NewQuestionsTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_questions",
		mcp.WithDescription("Searching for questions that already existed in the system. After the search, you can use the get_answers_by_question_id tool to get answers for the questions."),
		mcp.WithString(schema.MCPSearchCondKeyword,
			mcp.Description("Keyword to search for questions. Multiple keywords separated by spaces"),
		),
		mcp.WithString(schema.MCPSearchCondUsername,
			mcp.Description("Search for questions that contain only those created by the specified user"),
		),
		mcp.WithString(schema.MCPSearchCondTag,
			mcp.Description("Filter by tag (semicolon separated for multiple tags)"),
		),
		mcp.WithString(schema.MCPSearchCondScore,
			mcp.Description("Minimum score that the question must have"),
		),
	)
	return listFilesTool
}

func NewAnswersTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_answers_by_question_id",
		mcp.WithDescription("Search for all answers corresponding to the question ID. The question ID is provided by get_questions tool."),
		mcp.WithString(schema.MCPSearchCondQuestionID,
			mcp.Description("The ID of the question to which the answer belongs. The question ID is provided by get_questions tool."),
		),
	)
	return listFilesTool
}

func NewCommentsTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_comments",
		mcp.WithDescription("Searching for comments that already existed in the system"),
		mcp.WithString(schema.MCPSearchCondObjectID,
			mcp.Description("Queries comments on an object, either a question or an answer. object_id is the id of the object."),
		),
	)
	return listFilesTool
}

func NewTagsTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_tags",
		mcp.WithDescription("Searching for tags that already existed in the system"),
		mcp.WithString(schema.MCPSearchCondTagName,
			mcp.Description("Tag name"),
		),
	)
	return listFilesTool
}

func NewTagDetailTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_tag_detail",
		mcp.WithDescription("Get detailed information about a specific tag"),
		mcp.WithString(schema.MCPSearchCondTagName,
			mcp.Description("Tag name"),
		),
	)
	return listFilesTool
}

func NewUserTool() mcp.Tool {
	listFilesTool := mcp.NewTool("get_user",
		mcp.WithDescription("Searching for users that already existed in the system"),
		mcp.WithString(schema.MCPSearchCondUsername,
			mcp.Description("Username"),
		),
	)
	return listFilesTool
}

func NewSemanticSearchTool() mcp.Tool {
	tool := mcp.NewTool("semantic_search",
		mcp.WithDescription("Search questions and answers by semantic meaning. Use this when the user's question relates conceptually to existing content but may not match exact keywords. Returns the most semantically similar content."),
		mcp.WithString(schema.MCPSearchCondSemanticQuery,
			mcp.Required(),
			mcp.Description("The search query text to find semantically similar questions and answers"),
		),
		mcp.WithNumber(schema.MCPSearchCondTopK,
			mcp.Description("Maximum number of results to return (default 5)"),
		),
	)
	return tool
}
