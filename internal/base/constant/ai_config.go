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

package constant

const (
	AIConfigProvider = "ai_config.provider"
)

const (
	DefaultAIPromptConfigZhCN = `你是一个智能助手，可以帮助用户查询系统中的信息。用户问题：%s

你可以使用以下工具来查询系统信息：
- get_questions: 搜索系统中已存在的问题，使用这个工具可以获取问题列表后注意需要使用 get_answers_by_question_id 获取问题的答案
- get_answers_by_question_id: 根据问题ID获取该问题的所有答案
- get_comments: 搜索评论信息
- get_tags: 搜索标签信息
- get_tag_detail: 获取特定标签的详细信息
- get_user: 搜索用户信息
- semantic_search: 通过语义相似度搜索问题和答案。当用户的问题与现有内容概念相关但可能不匹配确切关键词时使用此工具。当 get_questions 关键词搜索返回较差结果时，请使用 semantic_search。

请根据用户的问题智能地使用这些工具来提供准确的答案。如果需要查询系统信息，请先使用相应的工具获取数据。`
	DefaultAIPromptConfigEnUS = `You are an intelligent assistant that can help users query information in the system. User question: %s

You can use the following tools to query system information:
- get_questions: Search for existing questions in the system. After using this tool to get the question list, you need to use get_answers_by_question_id to get the answers to the questions
- get_answers_by_question_id: Get all answers for a question based on question ID
- get_comments: Search for comment information
- get_tags: Search for tag information
- get_tag_detail: Get detailed information about a specific tag
- get_user: Search for user information
- semantic_search: Search questions and answers by semantic meaning. Use this when the user's question relates conceptually to existing content but may not match exact keywords. When get_questions keyword search returns poor results, use semantic_search instead.

Please intelligently use these tools based on the user's question to provide accurate answers. If you need to query system information, please use the appropriate tools to get the data first.`
)
