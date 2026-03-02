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

package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/entity"
	embeddingRepo "github.com/apache/answer/internal/repo/embedding"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/comment"
	"github.com/apache/answer/internal/service/content"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/robfig/cron/v3"
	"github.com/sashabaranov/go-openai"
	"github.com/segmentfault/pacman/log"
)

const (
	EmbeddingLevelQuestion = "question"
	EmbeddingLevelAnswer   = "answer"
)

// EmbeddingService handles embedding generation, text aggregation, and indexing.
type EmbeddingService struct {
	embeddingRepo   embeddingRepo.EmbeddingRepo
	searchService   *content.SearchService
	answerService   *content.AnswerService
	questionCommon  *questioncommon.QuestionCommon
	commentRepo     comment.CommentRepo
	siteInfoService siteinfo_common.SiteInfoCommonService

	mu       sync.Mutex
	cronJob  *cron.Cron
	cronSpec string
}

// NewEmbeddingService creates a new EmbeddingService.
func NewEmbeddingService(
	embeddingRepo embeddingRepo.EmbeddingRepo,
	searchService *content.SearchService,
	answerService *content.AnswerService,
	questionCommon *questioncommon.QuestionCommon,
	commentRepo comment.CommentRepo,
	siteInfoService siteinfo_common.SiteInfoCommonService,
) *EmbeddingService {
	return &EmbeddingService{
		embeddingRepo:   embeddingRepo,
		searchService:   searchService,
		answerService:   answerService,
		questionCommon:  questionCommon,
		commentRepo:     commentRepo,
		siteInfoService: siteInfoService,
	}
}

// getAIConfig returns the current AI configuration.
func (s *EmbeddingService) getAIConfig(ctx context.Context) (*schema.SiteAIResp, *schema.SiteAIProvider, error) {
	aiConfig, err := s.siteInfoService.GetSiteAI(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get AI config failed: %w", err)
	}
	if !aiConfig.Enabled {
		return nil, nil, fmt.Errorf("AI feature is disabled")
	}
	provider := aiConfig.GetProvider()
	if provider.EmbeddingModel == "" {
		return nil, nil, fmt.Errorf("embedding model not configured")
	}
	return aiConfig, provider, nil
}

// createEmbeddingClient creates an OpenAI-compatible client for embedding requests.
func (s *EmbeddingService) createEmbeddingClient(provider *schema.SiteAIProvider) *openai.Client {
	config := openai.DefaultConfig(provider.APIKey)
	config.BaseURL = provider.APIHost
	if !strings.HasSuffix(config.BaseURL, "/v1") {
		config.BaseURL += "/v1"
	}
	return openai.NewClientWithConfig(config)
}

// GenerateEmbedding generates an embedding vector for the given text.
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	_, provider, err := s.getAIConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := s.createEmbeddingClient(provider)
	resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Input: []string{text},
		Model: openai.EmbeddingModel(provider.EmbeddingModel),
	})
	if err != nil {
		return nil, fmt.Errorf("create embeddings failed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return resp.Data[0].Embedding, nil
}

// ComputeContentHash computes SHA256 of the text.
func ComputeContentHash(text string) string {
	h := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", h)
}

// BuildTextForQuestion aggregates question title + body + all answers + comments into one text.
// Uses SearchService and QuestionCommon to respect the plugin architecture.
func (s *EmbeddingService) BuildTextForQuestion(ctx context.Context, questionID string) (text string, meta *entity.EmbeddingMetadata, err error) {
	// Get question detail via service layer
	question, err := s.questionCommon.Info(ctx, questionID, "")
	if err != nil {
		return "", nil, fmt.Errorf("get question info failed: %w", err)
	}

	meta = &entity.EmbeddingMetadata{
		QuestionID: questionID,
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Question: %s\n%s", question.Title, question.Content))

	// Get answers via AnswerService
	answerInfoList, _, err := s.answerService.SearchList(ctx, &schema.AnswerListReq{
		QuestionID: questionID,
		Page:       1,
		PageSize:   50,
	})
	if err != nil {
		log.Warnf("get answers for question %s failed: %v", questionID, err)
	} else {
		for _, a := range answerInfoList {
			parts = append(parts, fmt.Sprintf("Answer: %s", a.Content))
			answerMeta := entity.EmbeddingMetadataAnswer{
				AnswerID: a.ID,
			}

			// Get comments on this answer
			answerComments, _, err := s.commentRepo.GetCommentPage(ctx, &comment.CommentQuery{
				PageCond:  pager.PageCond{Page: 1, PageSize: 50},
				ObjectID:  a.ID,
				QueryCond: "newest",
			})
			if err != nil {
				log.Warnf("get comments for answer %s failed: %v", a.ID, err)
			} else {
				for _, c := range answerComments {
					parts = append(parts, fmt.Sprintf("Comment on answer: %s", c.OriginalText))
					answerMeta.Comments = append(answerMeta.Comments, entity.EmbeddingMetadataComment{
						CommentID: c.ID,
					})
				}
			}
			meta.Answers = append(meta.Answers, answerMeta)
		}
	}

	// Get comments on the question
	commentList, _, err := s.commentRepo.GetCommentPage(ctx, &comment.CommentQuery{
		PageCond:  pager.PageCond{Page: 1, PageSize: 50},
		ObjectID:  questionID,
		QueryCond: "newest",
	})
	if err != nil {
		log.Warnf("get comments for question %s failed: %v", questionID, err)
	} else {
		for _, c := range commentList {
			parts = append(parts, fmt.Sprintf("Comment: %s", c.OriginalText))
			meta.Comments = append(meta.Comments, entity.EmbeddingMetadataComment{
				CommentID: c.ID,
			})
		}
	}

	return strings.Join(parts, "\n\n"), meta, nil
}

// BuildTextForAnswer aggregates answer body + parent question title + answer comments into one text.
func (s *EmbeddingService) BuildTextForAnswer(ctx context.Context, answerID, questionID string) (text string, meta *entity.EmbeddingMetadata, err error) {
	// Get parent question title
	question, err := s.questionCommon.Info(ctx, questionID, "")
	if err != nil {
		return "", nil, fmt.Errorf("get question info for answer failed: %w", err)
	}

	meta = &entity.EmbeddingMetadata{
		QuestionID: questionID,
		AnswerID:   answerID,
	}

	// Get the specific answer's content via AnswerService
	answerInfo, err := s.answerService.GetDetail(ctx, answerID)
	if err != nil {
		return "", nil, fmt.Errorf("get answer failed: %w", err)
	}

	var answerText string
	if answerInfo != nil {
		answerText = answerInfo.Content
	}

	var parts []string
	parts = append(parts, fmt.Sprintf("Question: %s", question.Title))
	if answerText != "" {
		parts = append(parts, fmt.Sprintf("Answer: %s", answerText))
		meta.Answers = append(meta.Answers, entity.EmbeddingMetadataAnswer{
			AnswerID: answerID,
		})
	}

	// Get comments on the answer
	commentList, _, err := s.commentRepo.GetCommentPage(ctx, &comment.CommentQuery{
		PageCond:  pager.PageCond{Page: 1, PageSize: 50},
		ObjectID:  answerID,
		QueryCond: "newest",
	})
	if err != nil {
		log.Warnf("get comments for answer %s failed: %v", answerID, err)
	} else {
		for _, c := range commentList {
			parts = append(parts, fmt.Sprintf("Comment: %s", c.OriginalText))
			if len(meta.Answers) > 0 {
				meta.Answers[0].Comments = append(meta.Answers[0].Comments, entity.EmbeddingMetadataComment{
					CommentID: c.ID,
				})
			} else {
				meta.Comments = append(meta.Comments, entity.EmbeddingMetadataComment{
					CommentID: c.ID,
				})
			}
		}
	}

	return strings.Join(parts, "\n\n"), meta, nil
}

// IndexQuestion indexes a single question embedding.
func (s *EmbeddingService) IndexQuestion(ctx context.Context, questionID string) error {
	text, meta, err := s.BuildTextForQuestion(ctx, questionID)
	if err != nil {
		return err
	}

	contentHash := ComputeContentHash(text)

	// Check staleness
	existing, exist, _ := s.embeddingRepo.GetByObjectID(ctx, questionID, EmbeddingLevelQuestion)
	if exist && existing.ContentHash == contentHash {
		return nil // already up to date
	}

	vec, err := s.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("generate embedding for question %s failed: %w", questionID, err)
	}

	metaJSON, _ := json.Marshal(meta)
	vecJSON, _ := json.Marshal(vec)

	return s.embeddingRepo.Upsert(ctx, &entity.Embedding{
		ObjectID:    questionID,
		ObjectType:  EmbeddingLevelQuestion,
		ContentHash: contentHash,
		Metadata:    string(metaJSON),
		Embedding:   string(vecJSON),
		Dimensions:  len(vec),
	})
}

// IndexAnswer indexes a single answer embedding.
func (s *EmbeddingService) IndexAnswer(ctx context.Context, answerID, questionID string) error {
	text, meta, err := s.BuildTextForAnswer(ctx, answerID, questionID)
	if err != nil {
		return err
	}

	contentHash := ComputeContentHash(text)

	existing, exist, _ := s.embeddingRepo.GetByObjectID(ctx, answerID, EmbeddingLevelAnswer)
	if exist && existing.ContentHash == contentHash {
		return nil
	}

	vec, err := s.GenerateEmbedding(ctx, text)
	if err != nil {
		return fmt.Errorf("generate embedding for answer %s failed: %w", answerID, err)
	}

	metaJSON, _ := json.Marshal(meta)
	vecJSON, _ := json.Marshal(vec)

	return s.embeddingRepo.Upsert(ctx, &entity.Embedding{
		ObjectID:    answerID,
		ObjectType:  EmbeddingLevelAnswer,
		ContentHash: contentHash,
		Metadata:    string(metaJSON),
		Embedding:   string(vecJSON),
		Dimensions:  len(vec),
	})
}

// SearchSimilar performs semantic search and returns top-K similar results.
// Results below the configured similarity threshold are filtered out.
func (s *EmbeddingService) SearchSimilar(ctx context.Context, query string, topK int) ([]embeddingRepo.SimilarResult, error) {
	vec, err := s.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding failed: %w", err)
	}
	results, err := s.embeddingRepo.SearchSimilar(ctx, vec, topK)
	if err != nil {
		return nil, err
	}

	for _, r := range results {
		log.Debugf("semantic search result: object_id=%s object_type=%s score=%.6f", r.ObjectID, r.ObjectType, r.Score)
	}

	// Apply similarity threshold from config (default 0 means no filtering)
	_, provider, cfgErr := s.getAIConfig(ctx)
	if cfgErr == nil && provider.SimilarityThreshold > 0 {
		filtered := make([]embeddingRepo.SimilarResult, 0, len(results))
		for _, r := range results {
			if r.Score >= provider.SimilarityThreshold {
				filtered = append(filtered, r)
			}
		}
		log.Debugf("semantic search: %d/%d results passed threshold %.4f", len(filtered), len(results), provider.SimilarityThreshold)
		return filtered, nil
	}

	return results, nil
}

// GetEmbeddingCount returns the total number of stored embeddings.
func (s *EmbeddingService) GetEmbeddingCount(ctx context.Context) (int64, error) {
	return s.embeddingRepo.Count(ctx)
}

// RemoveEmbedding removes an embedding by object ID and type.
func (s *EmbeddingService) RemoveEmbedding(ctx context.Context, objectID, objectType string) error {
	return s.embeddingRepo.DeleteByObjectID(ctx, objectID, objectType)
}

// IndexAll indexes all questions (and optionally answers) based on the configured embedding level.
func (s *EmbeddingService) IndexAll(ctx context.Context) error {
	_, provider, err := s.getAIConfig(ctx)
	if err != nil {
		log.Warnf("embedding indexer: %v", err)
		return err
	}

	level := provider.EmbeddingLevel
	if level == "" {
		level = EmbeddingLevelQuestion
	}

	log.Debugf("Starting embedding indexer at level: %s", level)

	page := 1
	totalIndexed := 0
	for {
		searchResp, err := s.searchService.Search(ctx, &schema.SearchDTO{
			Query: "is:question",
			Page:  page,
			Size:  50,
			Order: "newest",
		})
		if err != nil {
			return fmt.Errorf("search questions for indexing failed: %w", err)
		}
		if searchResp == nil || len(searchResp.SearchResults) == 0 {
			break
		}

		for _, result := range searchResp.SearchResults {
			if result.Object == nil {
				continue
			}
			qID := result.Object.QuestionID
			if level == EmbeddingLevelQuestion {
				if err := s.IndexQuestion(ctx, qID); err != nil {
					log.Warnf("index question %s failed: %v", qID, err)
					continue
				}
				totalIndexed++
			} else if level == EmbeddingLevelAnswer {
				// Index each answer for this question via AnswerService
				answerInfoList, _, err := s.answerService.SearchList(ctx, &schema.AnswerListReq{
					QuestionID: qID,
					Page:       1,
					PageSize:   50,
				})
				if err != nil {
					log.Warnf("get answers for question %s failed: %v", qID, err)
					continue
				}
				for _, a := range answerInfoList {
					if err := s.IndexAnswer(ctx, a.ID, qID); err != nil {
						log.Warnf("index answer %s failed: %v", a.ID, err)
						continue
					}
					totalIndexed++
				}
			}
		}

		if int64((page)*50) >= searchResp.Total {
			break
		}
		page++
	}

	log.Infof("Embedding indexer completed: %d items indexed", totalIndexed)
	return nil
}

// StartScheduler starts a cron job to periodically run IndexAll.
func (s *EmbeddingService) StartScheduler(spec string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing cron if running
	if s.cronJob != nil {
		s.cronJob.Stop()
		s.cronJob = nil
		s.cronSpec = ""
	}

	if spec == "" {
		return nil
	}

	c := cron.New()
	_, err := c.AddFunc(spec, func() {
		ctx := context.Background()
		log.Infof("embedding cron triggered (spec=%s)", spec)
		if err := s.IndexAll(ctx); err != nil {
			log.Errorf("embedding cron IndexAll failed: %v", err)
		}
	})
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", spec, err)
	}

	c.Start()
	s.cronJob = c
	s.cronSpec = spec
	log.Infof("embedding scheduler started with cron: %s", spec)
	return nil
}

// StopScheduler stops the embedding cron scheduler.
func (s *EmbeddingService) StopScheduler() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cronJob != nil {
		s.cronJob.Stop()
		s.cronJob = nil
		s.cronSpec = ""
		log.Infof("embedding scheduler stopped")
	}
}

// ApplyConfig reads the current AI config and starts or stops the scheduler accordingly.
func (s *EmbeddingService) ApplyConfig(ctx context.Context) {
	aiConfig, provider, err := s.getAIConfig(ctx)
	if err != nil || aiConfig == nil || provider == nil {
		s.StopScheduler()
		return
	}

	if provider.EmbeddingModel == "" || provider.EmbeddingCrontab == "" {
		s.StopScheduler()
		return
	}

	// Only restart if the cron spec changed
	s.mu.Lock()
	currentSpec := s.cronSpec
	s.mu.Unlock()

	if currentSpec == provider.EmbeddingCrontab {
		return
	}

	if err := s.StartScheduler(provider.EmbeddingCrontab); err != nil {
		log.Errorf("failed to start embedding scheduler: %v", err)
	}
}
