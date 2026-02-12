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
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/schema/mcp_tools"
	"github.com/apache/answer/internal/service/ai_conversation"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/feature_toggle"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/siteinfo_common"
	tagcommonser "github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/i18n"
	"github.com/segmentfault/pacman/log"
)

type AIController struct {
	searchService         *content.SearchService
	siteInfoService       siteinfo_common.SiteInfoCommonService
	tagCommonService      *tagcommonser.TagCommonService
	questioncommon        *questioncommon.QuestionCommon
	commentRepo           comment.CommentRepo
	userCommon            *usercommon.UserCommon
	answerRepo            answercommon.AnswerRepo
	mcpController         *MCPController
	aiConversationService ai_conversation.AIConversationService
	featureToggleSvc      *feature_toggle.FeatureToggleService
}

// NewAIController new site info controller.
func NewAIController(
	searchService *content.SearchService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	tagCommonService *tagcommonser.TagCommonService,
	questioncommon *questioncommon.QuestionCommon,
	commentRepo comment.CommentRepo,
	userCommon *usercommon.UserCommon,
	answerRepo answercommon.AnswerRepo,
	mcpController *MCPController,
	aiConversationService ai_conversation.AIConversationService,
	featureToggleSvc *feature_toggle.FeatureToggleService,
) *AIController {
	return &AIController{
		searchService:         searchService,
		siteInfoService:       siteInfoService,
		tagCommonService:      tagCommonService,
		questioncommon:        questioncommon,
		commentRepo:           commentRepo,
		userCommon:            userCommon,
		answerRepo:            answerRepo,
		mcpController:         mcpController,
		aiConversationService: aiConversationService,
		featureToggleSvc:      featureToggleSvc,
	}
}

func (c *AIController) ensureAIChatEnabled(ctx *gin.Context) bool {
	if c.featureToggleSvc == nil {
		return true
	}
	if err := c.featureToggleSvc.EnsureEnabled(ctx, feature_toggle.FeatureAIChatbot); err != nil {
		handler.HandleResponse(ctx, err, nil)
		return false
	}
	return true
}

type ChatCompletionsRequest struct {
	Messages       []Message `validate:"required,gte=1" json:"messages"`
	ConversationID string    `json:"conversation_id"`
	UserID         string    `json:"-"`
}

type Message struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ChatCompletionsResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type StreamResponse struct {
	ChatCompletionID string         `json:"chat_completion_id"`
	Object           string         `json:"object"`
	Created          int64          `json:"created"`
	Model            string         `json:"model"`
	Choices          []StreamChoice `json:"choices"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type StreamChoice struct {
	Index        int     `json:"index"`
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ConversationContext struct {
	ConversationID    string
	UserID            string
	UserQuestion      string
	Messages          []*ai_conversation.ConversationMessage
	IsNewConversation bool
	Model             string
}

func (c *ConversationContext) GetOpenAIMessages() []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, len(c.Messages))
	for i, msg := range c.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return messages
}

// sendStreamData
func sendStreamData(w http.ResponseWriter, data StreamResponse) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	_, _ = fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (c *AIController) ChatCompletions(ctx *gin.Context) {
	if !c.ensureAIChatEnabled(ctx) {
		return
	}
	aiConfig, err := c.siteInfoService.GetSiteAI(context.Background())
	if err != nil {
		log.Errorf("Failed to get AI config: %v", err)
		handler.HandleResponse(ctx, errors.BadRequest("AI service configuration error"), nil)
		return
	}

	if !aiConfig.Enabled {
		handler.HandleResponse(ctx, errors.ServiceUnavailable("AI service is not enabled"), nil)
		return
	}

	aiProvider := aiConfig.GetProvider()

	req := &ChatCompletionsRequest{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	req.UserID = middleware.GetLoginUserIDFromContext(ctx)

	data, _ := json.Marshal(req)
	log.Infof("ai chat request data: %s", string(data))

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	ctx.Status(http.StatusOK)

	w := ctx.Writer

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	chatcmplID := "chatcmpl-" + token.GenerateToken()
	created := time.Now().Unix()

	firstResponse := StreamResponse{
		ChatCompletionID: chatcmplID,
		Object:           "chat.completion.chunk",
		Created:          time.Now().Unix(),
		Model:            aiProvider.Model,
		Choices:          []StreamChoice{{Index: 0, Delta: Delta{Role: "assistant"}, FinishReason: nil}},
	}

	sendStreamData(w, firstResponse)

	conversationCtx := c.initializeConversationContext(ctx, aiProvider.Model, req)
	if conversationCtx == nil {
		log.Error("Failed to initialize conversation context")
		c.sendErrorResponse(w, chatcmplID, aiProvider.Model, "Failed to initialize conversation context")
		return
	}

	c.redirectRequestToAI(ctx, w, chatcmplID, conversationCtx)

	finishReason := "stop"
	endResponse := StreamResponse{
		ChatCompletionID: chatcmplID,
		Object:           "chat.completion.chunk",
		Created:          created,
		Model:            aiProvider.Model,
		Choices:          []StreamChoice{{Index: 0, Delta: Delta{}, FinishReason: &finishReason}},
	}

	sendStreamData(w, endResponse)

	_, _ = fmt.Fprintf(w, "data: [DONE]\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	c.saveConversationRecord(ctx, chatcmplID, conversationCtx)
}

func (c *AIController) redirectRequestToAI(ctx *gin.Context, w http.ResponseWriter, id string, conversationCtx *ConversationContext) {
	client := c.createOpenAIClient()

	c.handleAIConversation(ctx, w, id, client, conversationCtx)
}

// createOpenAIClient
func (c *AIController) createOpenAIClient() *openai.Client {
	config := openai.DefaultConfig("")
	config.BaseURL = ""

	aiConfig, err := c.siteInfoService.GetSiteAI(context.Background())
	if err != nil {
		log.Errorf("Failed to get AI config: %v", err)
		return openai.NewClientWithConfig(config)
	}

	if !aiConfig.Enabled {
		log.Warn("AI feature is disabled")
		return openai.NewClientWithConfig(config)
	}

	aiProvider := aiConfig.GetProvider()

	if aiProvider.Provider == "azure_ai" {
		return c.createAzureAIClient(aiProvider)
	}

	config = openai.DefaultConfig(aiProvider.APIKey)
	config.BaseURL = aiProvider.APIHost
	if !strings.HasSuffix(config.BaseURL, "/v1") {
		config.BaseURL += "/v1"
	}
	return openai.NewClientWithConfig(config)
}

// createAzureAIClient creates an OpenAI client configured for Azure AI.
// Uses the Azure OpenAI compatibility endpoint: https://{resource}.openai.azure.com/openai/v1
func (c *AIController) createAzureAIClient(aiProvider *schema.SiteAIProvider) *openai.Client {
	azureBaseURL := strings.TrimRight(aiProvider.APIHost, "/") + "/openai/v1"

	config := openai.DefaultConfig(aiProvider.APIKey)
	config.BaseURL = azureBaseURL
	config.HTTPClient = &http.Client{
		Transport: &azureAPIKeyTransport{
			apiKey:    aiProvider.APIKey,
			transport: http.DefaultTransport,
		},
	}
	return openai.NewClientWithConfig(config)
}

// azureAPIKeyTransport is an http.RoundTripper that replaces the Authorization
// header with the Azure-style api-key header for Azure OpenAI requests.
type azureAPIKeyTransport struct {
	apiKey    string
	transport http.RoundTripper
}

func (t *azureAPIKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Del("Authorization")
	req.Header.Set("Api-Key", t.apiKey)
	return t.transport.RoundTrip(req)
}

// getPromptByLanguage
func (c *AIController) getPromptByLanguage(language i18n.Language, question string) string {
	aiConfig, err := c.siteInfoService.GetSiteAI(context.Background())
	if err != nil {
		log.Errorf("Failed to get AI config: %v", err)
		return c.getDefaultPrompt(language, question)
	}

	var promptTemplate string

	switch language {
	case i18n.LanguageChinese:
		promptTemplate = aiConfig.PromptConfig.ZhCN
	case i18n.LanguageEnglish:
		promptTemplate = aiConfig.PromptConfig.EnUS
	default:
		promptTemplate = aiConfig.PromptConfig.EnUS
	}

	if promptTemplate == "" {
		return c.getDefaultPrompt(language, question)
	}

	return fmt.Sprintf(promptTemplate, question)
}

// getDefaultPrompt prompt
func (c *AIController) getDefaultPrompt(language i18n.Language, question string) string {
	switch language {
	case i18n.LanguageChinese:
		return fmt.Sprintf(constant.DefaultAIPromptConfigZhCN, question)
	case i18n.LanguageEnglish:
		return fmt.Sprintf(constant.DefaultAIPromptConfigEnUS, question)
	default:
		return fmt.Sprintf(constant.DefaultAIPromptConfigEnUS, question)
	}
}

// initializeConversationContext
func (c *AIController) initializeConversationContext(ctx *gin.Context, model string, req *ChatCompletionsRequest) *ConversationContext {
	if len(req.ConversationID) == 0 {
		req.ConversationID = token.GenerateToken()
	}
	conversationCtx := &ConversationContext{
		UserID:         req.UserID,
		Messages:       make([]*ai_conversation.ConversationMessage, 0),
		ConversationID: req.ConversationID,
		Model:          model,
	}

	conversationDetail, exist, err := c.aiConversationService.GetConversationDetail(ctx, &schema.AIConversationDetailReq{
		ConversationID: req.ConversationID,
		UserID:         req.UserID,
	})
	if err != nil {
		log.Errorf("Failed to get conversation detail: %v", err)
		return nil
	}
	if !exist {
		conversationCtx.UserQuestion = req.Messages[0].Content
		conversationCtx.Messages = c.buildInitialMessages(ctx, req)
		conversationCtx.IsNewConversation = true
		return conversationCtx
	}
	conversationCtx.IsNewConversation = false

	for _, record := range conversationDetail.Records {
		conversationCtx.Messages = append(conversationCtx.Messages, &ai_conversation.ConversationMessage{
			ChatCompletionID: record.ChatCompletionID,
			Role:             record.Role,
			Content:          record.Content,
		})
	}
	conversationCtx.Messages = append(conversationCtx.Messages, &ai_conversation.ConversationMessage{
		Role:    req.Messages[0].Role,
		Content: req.Messages[0].Content,
	})
	return conversationCtx
}

// buildInitialMessages
func (c *AIController) buildInitialMessages(ctx *gin.Context, req *ChatCompletionsRequest) []*ai_conversation.ConversationMessage {
	question := ""
	if len(req.Messages) == 1 {
		question = req.Messages[0].Content
	} else {
		messages := make([]*ai_conversation.ConversationMessage, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = &ai_conversation.ConversationMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		return messages
	}

	currentLang := handler.GetLangByCtx(ctx)

	prompt := c.getPromptByLanguage(currentLang, question)

	return []*ai_conversation.ConversationMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}}
}

// saveConversationRecord
func (c *AIController) saveConversationRecord(ctx context.Context, chatcmplID string, conversationCtx *ConversationContext) {
	if conversationCtx == nil || len(conversationCtx.Messages) == 0 {
		return
	}

	if conversationCtx.IsNewConversation {
		topic := conversationCtx.UserQuestion
		if topic == "" {
			log.Warn("No user message found for new conversation")
			return
		}

		err := c.aiConversationService.CreateConversation(ctx, conversationCtx.UserID, conversationCtx.ConversationID, topic)
		if err != nil {
			log.Errorf("Failed to create conversation: %v", err)
			return
		}
	}

	err := c.aiConversationService.SaveConversationRecords(ctx, conversationCtx.ConversationID, chatcmplID, conversationCtx.Messages)
	if err != nil {
		log.Errorf("Failed to save conversation records: %v", err)
	}
}

func (c *AIController) handleAIConversation(ctx *gin.Context, w http.ResponseWriter, id string, client *openai.Client, conversationCtx *ConversationContext) {
	maxRounds := 10
	messages := conversationCtx.GetOpenAIMessages()

	for round := range maxRounds {
		log.Debugf("AI conversation round: %d", round+1)

		aiReq := openai.ChatCompletionRequest{
			Model:    conversationCtx.Model,
			Messages: messages,
			Tools:    c.getMCPTools(),
			Stream:   true,
		}

		toolCalls, newMessages, finished, aiResponse := c.processAIStream(ctx, w, id, conversationCtx.Model, client, aiReq, messages)
		messages = newMessages

		if aiResponse != "" {
			conversationCtx.Messages = append(conversationCtx.Messages, &ai_conversation.ConversationMessage{
				Role:    "assistant",
				Content: aiResponse,
			})
		}

		if finished {
			return
		}

		if len(toolCalls) > 0 {
			messages = c.executeToolCalls(ctx, w, id, conversationCtx.Model, toolCalls, messages)
		} else {
			return
		}
	}

	log.Warnf("AI conversation reached maximum rounds limit: %d", maxRounds)
}

// processAIStream
func (c *AIController) processAIStream(
	_ *gin.Context, w http.ResponseWriter, id, model string, client *openai.Client, aiReq openai.ChatCompletionRequest, messages []openai.ChatCompletionMessage) (
	[]openai.ToolCall, []openai.ChatCompletionMessage, bool, string) {
	stream, err := client.CreateChatCompletionStream(context.Background(), aiReq)
	if err != nil {
		log.Errorf("Failed to create stream: %v", err)
		c.sendErrorResponse(w, id, model, "Failed to create AI stream")
		return nil, messages, true, ""
	}
	defer func() {
		_ = stream.Close()
	}()

	var currentToolCalls []openai.ToolCall
	var accumulatedContent strings.Builder
	var accumulatedMessage openai.ChatCompletionMessage
	toolCallsMap := make(map[int]*openai.ToolCall)

	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				log.Info("Stream finished")
				break
			}
			log.Errorf("Stream error: %v", err)
			break
		}

		if len(response.Choices) == 0 {
			continue
		}

		choice := response.Choices[0]

		if len(choice.Delta.ToolCalls) > 0 {
			for _, deltaToolCall := range choice.Delta.ToolCalls {
				index := *deltaToolCall.Index

				if _, exists := toolCallsMap[index]; !exists {
					toolCallsMap[index] = &openai.ToolCall{
						ID:   deltaToolCall.ID,
						Type: deltaToolCall.Type,
						Function: openai.FunctionCall{
							Name:      deltaToolCall.Function.Name,
							Arguments: deltaToolCall.Function.Arguments,
						},
					}
				} else {
					if deltaToolCall.Function.Arguments != "" {
						toolCallsMap[index].Function.Arguments += deltaToolCall.Function.Arguments
					}
					if deltaToolCall.Function.Name != "" {
						toolCallsMap[index].Function.Name = deltaToolCall.Function.Name
					}
				}
			}
		}

		if choice.Delta.Content != "" {
			accumulatedContent.WriteString(choice.Delta.Content)

			contentResponse := StreamResponse{
				ChatCompletionID: id,
				Object:           "chat.completion.chunk",
				Created:          time.Now().Unix(),
				Model:            model,
				Choices: []StreamChoice{
					{
						Index: 0,
						Delta: Delta{
							Content: choice.Delta.Content,
						},
						FinishReason: nil,
					},
				},
			}
			sendStreamData(w, contentResponse)
		}

		if len(choice.FinishReason) > 0 {
			if choice.FinishReason == "tool_calls" {
				for _, toolCall := range toolCallsMap {
					currentToolCalls = append(currentToolCalls, *toolCall)
				}
				return currentToolCalls, messages, false, accumulatedContent.String()
			} else {
				aiResponseContent := accumulatedContent.String()
				if aiResponseContent != "" {
					accumulatedMessage = openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: aiResponseContent,
					}
					messages = append(messages, accumulatedMessage)
				}
				return nil, messages, true, aiResponseContent
			}
		}
	}

	aiResponseContent := accumulatedContent.String()
	if aiResponseContent != "" {
		accumulatedMessage = openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: aiResponseContent,
		}
		messages = append(messages, accumulatedMessage)
	}

	if len(toolCallsMap) > 0 {
		for _, toolCall := range toolCallsMap {
			currentToolCalls = append(currentToolCalls, *toolCall)
		}
		return currentToolCalls, messages, false, aiResponseContent
	}

	return currentToolCalls, messages, len(currentToolCalls) == 0, aiResponseContent
}

// executeToolCalls
func (c *AIController) executeToolCalls(ctx *gin.Context, _ http.ResponseWriter, _, _ string, toolCalls []openai.ToolCall, messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	validToolCalls := make([]openai.ToolCall, 0)
	for _, toolCall := range toolCalls {
		if toolCall.ID == "" || toolCall.Function.Name == "" {
			log.Errorf("Invalid tool call: missing required fields. ID: %s, Function: %v", toolCall.ID, toolCall.Function)
			continue
		}

		if toolCall.Function.Arguments == "" {
			toolCall.Function.Arguments = "{}"
		}

		validToolCalls = append(validToolCalls, toolCall)
		log.Debugf("Valid tool call: ID=%s, Name=%s, Arguments=%s", toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments)
	}

	if len(validToolCalls) == 0 {
		log.Warn("No valid tool calls found")
		return messages
	}

	assistantMsg := openai.ChatCompletionMessage{
		Role:      openai.ChatMessageRoleAssistant,
		ToolCalls: validToolCalls,
	}
	messages = append(messages, assistantMsg)

	for _, toolCall := range validToolCalls {
		if toolCall.Function.Name != "" {
			var args map[string]any
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				log.Errorf("Failed to parse tool arguments for %s: %v, arguments: %s", toolCall.Function.Name, err, toolCall.Function.Arguments)
				errorResult := fmt.Sprintf("Error parsing tool arguments: %v", err)
				toolMessage := openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    errorResult,
					ToolCallID: toolCall.ID,
				}
				messages = append(messages, toolMessage)
				continue
			}

			result, err := c.callMCPTool(ctx, toolCall.Function.Name, args)
			if err != nil {
				log.Errorf("Failed to call MCP tool %s: %v", toolCall.Function.Name, err)
				result = fmt.Sprintf("Error calling tool %s: %v", toolCall.Function.Name, err)
			}

			toolMessage := openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			}
			messages = append(messages, toolMessage)
		}
	}

	return messages
}

// sendErrorResponse send error response in stream
func (c *AIController) sendErrorResponse(w http.ResponseWriter, id, model, errorMsg string) {
	errorResponse := StreamResponse{
		ChatCompletionID: id,
		Object:           "chat.completion.chunk",
		Created:          time.Now().Unix(),
		Model:            model,
		Choices: []StreamChoice{
			{
				Index: 0,
				Delta: Delta{
					Content: fmt.Sprintf("Error: %s", errorMsg),
				},
				FinishReason: nil,
			},
		},
	}
	sendStreamData(w, errorResponse)
}

// getMCPTools
func (c *AIController) getMCPTools() []openai.Tool {
	openaiTools := make([]openai.Tool, 0)
	for _, mcpTool := range mcp_tools.MCPToolsList {
		openaiTool := c.convertMCPToolToOpenAI(mcpTool)
		openaiTools = append(openaiTools, openaiTool)
	}

	return openaiTools
}

// convertMCPToolToOpenAI
func (c *AIController) convertMCPToolToOpenAI(mcpTool mcp.Tool) openai.Tool {
	properties := make(map[string]any)
	required := make([]string, 0)

	maps.Copy(properties, mcpTool.InputSchema.Properties)

	required = append(required, mcpTool.InputSchema.Required...)

	parameters := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		parameters["required"] = required
	}

	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        mcpTool.Name,
			Description: mcpTool.Description,
			Parameters:  parameters,
		},
	}
}

// callMCPTool
func (c *AIController) callMCPTool(ctx context.Context, toolName string, arguments map[string]any) (string, error) {
	request := mcp.CallToolRequest{
		Request: mcp.Request{},
		Params: struct {
			Name      string    `json:"name"`
			Arguments any       `json:"arguments,omitempty"`
			Meta      *mcp.Meta `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: arguments,
		},
	}

	var result *mcp.CallToolResult
	var err error

	log.Debugf("Calling MCP tool: %s with arguments: %v", toolName, arguments)

	switch toolName {
	case "get_questions":
		result, err = c.mcpController.MCPQuestionsHandler()(ctx, request)
	case "get_answers_by_question_id":
		result, err = c.mcpController.MCPAnswersHandler()(ctx, request)
	case "get_comments":
		result, err = c.mcpController.MCPCommentsHandler()(ctx, request)
	case "get_tags":
		result, err = c.mcpController.MCPTagsHandler()(ctx, request)
	case "get_tag_detail":
		result, err = c.mcpController.MCPTagDetailsHandler()(ctx, request)
	case "get_user":
		result, err = c.mcpController.MCPUserDetailsHandler()(ctx, request)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	if err != nil {
		return "", err
	}

	data, _ := json.Marshal(result)
	log.Debugf("MCP tool %s called successfully, result: %v", toolName, string(data))

	if result != nil && len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			return textContent.Text, nil
		}
	}

	return "No result found", nil
}
