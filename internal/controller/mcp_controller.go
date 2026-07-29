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
	"strings"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/embedding"
	"github.com/apache/answer/internal/service/feature_toggle"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/siteinfo_common"
	tagcommonser "github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/plugin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/segmentfault/pacman/log"
)

type MCPController struct {
	searchService    *content.SearchService
	siteInfoService  siteinfo_common.SiteInfoCommonService
	tagCommonService *tagcommonser.TagCommonService
	questioncommon   *questioncommon.QuestionCommon
	commentRepo      comment.CommentRepo
	userCommon       *usercommon.UserCommon
	answerRepo       answercommon.AnswerRepo
	featureToggleSvc *feature_toggle.FeatureToggleService
	embeddingService *embedding.EmbeddingService
}

// NewMCPController new site info controller.
func NewMCPController(
	searchService *content.SearchService,
	siteInfoService siteinfo_common.SiteInfoCommonService,
	tagCommonService *tagcommonser.TagCommonService,
	questioncommon *questioncommon.QuestionCommon,
	commentRepo comment.CommentRepo,
	userCommon *usercommon.UserCommon,
	answerRepo answercommon.AnswerRepo,
	featureToggleSvc *feature_toggle.FeatureToggleService,
	embeddingService *embedding.EmbeddingService,
) *MCPController {
	return &MCPController{
		searchService:    searchService,
		siteInfoService:  siteInfoService,
		tagCommonService: tagCommonService,
		questioncommon:   questioncommon,
		commentRepo:      commentRepo,
		userCommon:       userCommon,
		answerRepo:       answerRepo,
		featureToggleSvc: featureToggleSvc,
		embeddingService: embeddingService,
	}
}

func (c *MCPController) ensureMCPEnabled(ctx context.Context) error {
	if c.featureToggleSvc == nil {
		return nil
	}
	return c.featureToggleSvc.EnsureEnabled(ctx, feature_toggle.FeatureMCP)
}

func (c *MCPController) MCPQuestionsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		searchResp, err := c.searchService.Search(ctx, &schema.SearchDTO{
			Query: cond.ToQueryString() + " is:question",
			Page:  1,
			Size:  5,
			Order: "newest",
		})
		if err != nil {
			return nil, err
		}

		resp := make([]*schema.MCPSearchQuestionInfoResp, 0)
		for _, question := range searchResp.SearchResults {
			t := &schema.MCPSearchQuestionInfoResp{
				QuestionID: question.Object.QuestionID,
				Title:      question.Object.Title,
				Content:    question.Object.Excerpt,
				Link:       fmt.Sprintf("%s/questions/%s", siteGeneral.SiteUrl, question.Object.QuestionID),
			}
			resp = append(resp, t)
		}

		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPQuestionDetailHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchQuestionDetail(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		question, err := c.questioncommon.Info(ctx, cond.QuestionID, "")
		if err != nil {
			log.Errorf("get question failed: %v", err)
			return mcp.NewToolResultText("No question found."), nil
		}

		resp := &schema.MCPSearchQuestionInfoResp{
			QuestionID: question.ID,
			Title:      question.Title,
			Content:    question.Content,
			Link:       fmt.Sprintf("%s/questions/%s", siteGeneral.SiteUrl, question.ID),
		}
		res, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(res)), nil
	}
}

func (c *MCPController) MCPAnswersHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchAnswerCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		if len(cond.QuestionID) > 0 {
			answerList, err := c.answerRepo.GetAnswerList(ctx, &entity.Answer{QuestionID: cond.QuestionID})
			if err != nil {
				log.Errorf("get answers failed: %v", err)
				return nil, err
			}
			resp := make([]*schema.MCPSearchAnswerInfoResp, 0)
			for _, answer := range answerList {
				if answer.Status != entity.AnswerStatusAvailable {
					continue
				}
				t := &schema.MCPSearchAnswerInfoResp{
					QuestionID:    answer.QuestionID,
					AnswerID:      answer.ID,
					AnswerContent: answer.OriginalText,
					Link:          fmt.Sprintf("%s/questions/%s/answers/%s", siteGeneral.SiteUrl, answer.QuestionID, answer.ID),
				}
				resp = append(resp, t)
			}
			data, _ := json.Marshal(resp)
			return mcp.NewToolResultText(string(data)), nil
		}

		answerList, err := c.answerRepo.GetAnswerList(ctx, &entity.Answer{QuestionID: cond.QuestionID})
		if err != nil {
			log.Errorf("get answers failed: %v", err)
			return nil, err
		}
		resp := make([]*schema.MCPSearchAnswerInfoResp, 0)
		for _, answer := range answerList {
			if answer.Status != entity.AnswerStatusAvailable {
				continue
			}
			t := &schema.MCPSearchAnswerInfoResp{
				QuestionID:    answer.QuestionID,
				AnswerID:      answer.ID,
				AnswerContent: answer.OriginalText,
				Link:          fmt.Sprintf("%s/questions/%s/answers/%s", siteGeneral.SiteUrl, answer.QuestionID, answer.ID),
			}
			resp = append(resp, t)
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPCommentsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchCommentCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		dto := &comment.CommentQuery{
			PageCond:  pager.PageCond{Page: 1, PageSize: 5},
			QueryCond: "newest",
			ObjectID:  cond.ObjectID,
		}
		commentList, total, err := c.commentRepo.GetCommentPage(ctx, dto)
		if err != nil {
			return nil, err
		}
		if total == 0 {
			return mcp.NewToolResultText("No comments found."), nil
		}

		resp := make([]*schema.MCPSearchCommentInfoResp, 0)
		for _, comment := range commentList {
			t := &schema.MCPSearchCommentInfoResp{
				CommentID: comment.ID,
				Content:   comment.OriginalText,
				ObjectID:  comment.ObjectID,
				Link:      fmt.Sprintf("%s/comments/%s", siteGeneral.SiteUrl, comment.ID),
			}
			resp = append(resp, t)
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPTagsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchTagCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		tags, total, err := c.tagCommonService.GetTagPage(ctx, 1, 10, &entity.Tag{DisplayName: cond.TagName}, "newest")
		if err != nil {
			log.Errorf("get tags failed: %v", err)
			return nil, err
		}

		if total == 0 {
			res := strings.Builder{}
			res.WriteString("No tags found.\n")
			return mcp.NewToolResultText(res.String()), nil
		}

		resp := make([]*schema.MCPSearchTagResp, 0)
		for _, tag := range tags {
			t := &schema.MCPSearchTagResp{
				TagName:     tag.SlugName,
				DisplayName: tag.DisplayName,
				Description: tag.OriginalText,
				Link:        fmt.Sprintf("%s/tags/%s", siteGeneral.SiteUrl, tag.SlugName),
			}
			resp = append(resp, t)
		}
		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (c *MCPController) MCPTagDetailsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchTagCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		tag, exist, err := c.tagCommonService.GetTagBySlugName(ctx, cond.TagName)
		if err != nil {
			log.Errorf("get tag failed: %v", err)
			return nil, err
		}
		if !exist {
			return mcp.NewToolResultText("Tag not found."), nil
		}

		resp := &schema.MCPSearchTagResp{
			TagName:     tag.SlugName,
			DisplayName: tag.DisplayName,
			Description: tag.OriginalText,
			Link:        fmt.Sprintf("%s/tags/%s", siteGeneral.SiteUrl, tag.SlugName),
		}
		res, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(res)), nil
	}
}

func (c *MCPController) MCPUserDetailsHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSearchUserCond(request)

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		user, exist, err := c.userCommon.GetUserBasicInfoByUserName(ctx, cond.Username)
		if err != nil {
			log.Errorf("get user failed: %v", err)
			return nil, err
		}
		if !exist {
			return mcp.NewToolResultText("User not found."), nil
		}

		resp := &schema.MCPSearchUserInfoResp{
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
			Link:        fmt.Sprintf("%s/users/%s", siteGeneral.SiteUrl, user.Username),
		}
		res, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(res)), nil
	}
}

func (c *MCPController) MCPSemanticSearchHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := c.ensureMCPEnabled(ctx); err != nil {
			return nil, err
		}
		cond := schema.NewMCPSemanticSearchCond(request)
		if len(cond.Query) == 0 {
			return mcp.NewToolResultText("Query is required for semantic search."), nil
		}

		siteGeneral, err := c.siteInfoService.GetSiteGeneral(ctx)
		if err != nil {
			log.Errorf("get site general info failed: %v", err)
			return nil, err
		}

		results, err := c.embeddingService.SearchSimilar(ctx, cond.Query, cond.TopK)
		if err != nil {
			log.Errorf("semantic search failed: %v", err)
			return mcp.NewToolResultText("Semantic search is not available. Embedding may not be configured."), nil
		}
		if len(results) == 0 {
			return mcp.NewToolResultText("No semantically similar content found."), nil
		}

		resp := make([]*schema.MCPSemanticSearchResp, 0, len(results))
		for _, r := range results {
			var meta plugin.VectorSearchMetadata
			_ = json.Unmarshal([]byte(r.Metadata), &meta)

			item := &schema.MCPSemanticSearchResp{
				ObjectID:   r.ObjectID,
				ObjectType: r.ObjectType,
				Score:      r.Score,
			}

			// Compose link from metadata
			if r.ObjectType == "answer" && meta.AnswerID != "" {
				item.Link = fmt.Sprintf("%s/questions/%s/%s", siteGeneral.SiteUrl, meta.QuestionID, meta.AnswerID)
			} else {
				item.Link = fmt.Sprintf("%s/questions/%s", siteGeneral.SiteUrl, meta.QuestionID)
			}

			// Query content from DB using IDs stored in metadata
			if r.ObjectType == "question" {
				question, qErr := c.questioncommon.Info(ctx, meta.QuestionID, "")
				if qErr != nil {
					log.Warnf("get question %s for semantic search failed: %v", meta.QuestionID, qErr)
				} else {
					item.Title = question.Title
					item.Content = question.Content
				}

				// Fetch answers by ID from metadata
				for _, a := range meta.Answers {
					answerEntity, exist, aErr := c.answerRepo.GetAnswer(ctx, a.AnswerID)
					if aErr != nil || !exist {
						continue
					}
					answerItem := &schema.MCPSemanticSearchAnswer{
						AnswerID: a.AnswerID,
						Content:  answerEntity.OriginalText,
					}
					// Fetch comments on this answer from DB
					for _, ac := range a.Comments {
						cmt, cExist, cErr := c.commentRepo.GetComment(ctx, ac.CommentID)
						if cErr == nil && cExist {
							answerItem.Comments = append(answerItem.Comments, &schema.MCPSemanticSearchComment{
								CommentID: ac.CommentID,
								Content:   cmt.OriginalText,
							})
						}
					}
					item.Answers = append(item.Answers, answerItem)
				}

				// Fetch question comments from DB
				for _, qc := range meta.Comments {
					cmt, cExist, cErr := c.commentRepo.GetComment(ctx, qc.CommentID)
					if cErr == nil && cExist {
						item.Comments = append(item.Comments, &schema.MCPSemanticSearchComment{
							CommentID: qc.CommentID,
							Content:   cmt.OriginalText,
						})
					}
				}
			} else if r.ObjectType == "answer" {
				// Fetch question title for context
				question, qErr := c.questioncommon.Info(ctx, meta.QuestionID, "")
				if qErr == nil {
					item.Title = question.Title
				}

				// Fetch answer content from DB
				if meta.AnswerID != "" {
					answerEntity, exist, aErr := c.answerRepo.GetAnswer(ctx, meta.AnswerID)
					if aErr == nil && exist {
						item.Content = answerEntity.OriginalText
					}
				} else if len(meta.Answers) > 0 {
					answerEntity, exist, aErr := c.answerRepo.GetAnswer(ctx, meta.Answers[0].AnswerID)
					if aErr == nil && exist {
						item.Content = answerEntity.OriginalText
					}
				}

				// Fetch answer comments from DB
				if len(meta.Answers) > 0 {
					for _, ac := range meta.Answers[0].Comments {
						cmt, cExist, cErr := c.commentRepo.GetComment(ctx, ac.CommentID)
						if cErr == nil && cExist {
							item.Comments = append(item.Comments, &schema.MCPSemanticSearchComment{
								CommentID: ac.CommentID,
								Content:   cmt.OriginalText,
							})
						}
					}
				}
			}

			resp = append(resp, item)
		}

		data, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(data)), nil
	}
}
