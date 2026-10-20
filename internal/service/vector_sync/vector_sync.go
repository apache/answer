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

package vector_sync

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/queue"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/vector_search_sync"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"
)

const (
	ActionUpsert = "upsert"
	ActionDelete = "delete"
	// ActionRestoreQuestion rebuilds a public question and its available
	// answers after the question itself was restored or shown again.
	ActionRestoreQuestion = "restore_question"

	ObjectTypeQuestion = "question"
	ObjectTypeAnswer   = "answer"
)

const maxRetry = 3

const cleanupPageSize = 500

type Task struct {
	Action     string
	ObjectType string
	ObjectID   string
}

type Service queue.Service[*Task]

func NewService(data *data.Data) Service {
	q := queue.New[*Task]("vector_sync", 128)
	q.RegisterHandler(func(ctx context.Context, msg *Task) error {
		return handle(ctx, data, msg)
	})
	return q
}

func handle(ctx context.Context, data *data.Data, msg *Task) error {
	if msg == nil || msg.ObjectID == "" {
		return nil
	}

	var vectorSearch plugin.VectorSearch
	_ = plugin.CallVectorSearch(func(vs plugin.VectorSearch) error {
		vectorSearch = vs
		return nil
	})
	if vectorSearch == nil {
		return nil
	}

	objectID := uid.DeShortID(msg.ObjectID)
	var lastErr error
	for attempt := 1; attempt <= maxRetry; attempt++ {
		err := handleOnce(ctx, data, vectorSearch, msg.Action, msg.ObjectType, objectID)
		if err == nil {
			return nil
		}
		lastErr = err
		log.Warnf("vector sync failed: action=%s object_type=%s object_id=%s attempt=%d err=%v",
			msg.Action, msg.ObjectType, objectID, attempt, err)
	}
	return lastErr
}

func handleOnce(ctx context.Context, data *data.Data, vectorSearch plugin.VectorSearch,
	action, objectType, objectID string) error {
	if action == ActionDelete {
		if objectType == ObjectTypeQuestion {
			return deleteQuestionAndAnswerContents(ctx, data, vectorSearch, objectID)
		}
		return vectorSearch.DeleteContent(ctx, objectID)
	}
	if action != ActionUpsert {
		if action == ActionRestoreQuestion && objectType == ObjectTypeQuestion {
			return restoreQuestionAndAnswerContents(ctx, data, vectorSearch, objectID)
		}
		return nil
	}

	var (
		content *plugin.VectorSearchContent
		err     error
	)
	switch objectType {
	case ObjectTypeQuestion:
		content, err = vector_search_sync.BuildQuestionContentByID(ctx, data, objectID)
	case ObjectTypeAnswer:
		content, err = vector_search_sync.BuildAnswerContentByID(ctx, data, objectID)
	default:
		return nil
	}
	if err != nil {
		return err
	}
	if content == nil {
		if objectType == ObjectTypeQuestion {
			return deleteQuestionAndAnswerContents(ctx, data, vectorSearch, objectID)
		}
		return vectorSearch.DeleteContent(ctx, objectID)
	}
	return vectorSearch.UpdateContent(ctx, content)
}

func restoreQuestionAndAnswerContents(ctx context.Context, data *data.Data, vectorSearch plugin.VectorSearch, questionID string) error {
	questionID = uid.DeShortID(questionID)
	content, err := vector_search_sync.BuildQuestionContentByID(ctx, data, questionID)
	if err != nil {
		return err
	}
	if content == nil {
		return deleteQuestionAndAnswerContents(ctx, data, vectorSearch, questionID)
	}
	if err := vectorSearch.UpdateContent(ctx, content); err != nil {
		return err
	}
	answers := make([]*entity.Answer, 0)
	if err := data.DB.Context(ctx).Where("question_id = ? AND status = ?", questionID,
		entity.AnswerStatusAvailable).Find(&answers); err != nil {
		return fmt.Errorf("list answer vectors for restored question %s: %w", questionID, err)
	}
	for _, answer := range answers {
		answerContent, err := vector_search_sync.BuildAnswerContentByID(ctx, data, answer.ID)
		if err != nil {
			return err
		}
		if answerContent == nil {
			continue
		}
		if err := vectorSearch.UpdateContent(ctx, answerContent); err != nil {
			return fmt.Errorf("restore answer vector %s: %w", answer.ID, err)
		}
	}
	return nil
}

func deleteQuestionAndAnswerContents(ctx context.Context, data *data.Data, vectorSearch plugin.VectorSearch, questionID string) error {
	questionID = uid.DeShortID(questionID)
	if err := vectorSearch.DeleteContent(ctx, questionID); err != nil {
		return err
	}
	answers := make([]*entity.Answer, 0)
	if err := data.DB.Context(ctx).Where("question_id = ?", questionID).Find(&answers); err != nil {
		return fmt.Errorf("list answer vectors for question %s: %w", questionID, err)
	}
	for _, answer := range answers {
		if err := vectorSearch.DeleteContent(ctx, uid.DeShortID(answer.ID)); err != nil {
			return fmt.Errorf("delete answer vector %s: %w", answer.ID, err)
		}
	}
	return nil
}

// RemoveRestrictedContents removes documents that were indexed before visibility
// filtering was introduced. It deliberately uses the existing VectorSearch
// interface, so installed plugins remain binary-compatible.
func RemoveRestrictedContents(ctx context.Context, data *data.Data, vectorSearch plugin.VectorSearch) error {
	if data == nil || data.DB == nil || vectorSearch == nil {
		return nil
	}

	questions := make([]*entity.Question, 0, cleanupPageSize)
	for offset := 0; ; offset += cleanupPageSize {
		questions = questions[:0]
		if err := data.DB.Context(ctx).OrderBy("id ASC").Limit(cleanupPageSize, offset).Find(&questions); err != nil {
			return fmt.Errorf("list questions for vector cleanup: %w", err)
		}
		for _, question := range questions {
			if question.Show == entity.QuestionShow &&
				(question.Status == entity.QuestionStatusAvailable || question.Status == entity.QuestionStatusClosed) {
				hasRestricted, err := hasRestrictedQuestionAggregate(ctx, data, question.ID)
				if err != nil {
					return err
				}
				if hasRestricted {
					// The following plugin bulk sync rebuilds this document from the
					// new filtered source. Delete the old aggregate first so deleted
					// answers/comments cannot remain in the vector store.
					if err := vectorSearch.DeleteContent(ctx, uid.DeShortID(question.ID)); err != nil {
						return fmt.Errorf("delete stale question vector %s: %w", question.ID, err)
					}
				}
				continue
			}
			if err := deleteQuestionAndAnswerContents(ctx, data, vectorSearch, question.ID); err != nil {
				return fmt.Errorf("delete vectors for restricted question %s: %w", question.ID, err)
			}
		}
		if len(questions) < cleanupPageSize {
			break
		}
	}

	answers := make([]*entity.Answer, 0, cleanupPageSize)
	for offset := 0; ; offset += cleanupPageSize {
		answers = answers[:0]
		if err := data.DB.Context(ctx).Where("status <> ?", entity.AnswerStatusAvailable).
			OrderBy("id ASC").Limit(cleanupPageSize, offset).Find(&answers); err != nil {
			return fmt.Errorf("list restricted answers for vector cleanup: %w", err)
		}
		for _, answer := range answers {
			if err := vectorSearch.DeleteContent(ctx, uid.DeShortID(answer.ID)); err != nil {
				return fmt.Errorf("delete restricted answer vector %s: %w", answer.ID, err)
			}
		}
		if len(answers) < cleanupPageSize {
			break
		}
	}

	for offset := 0; ; offset += cleanupPageSize {
		answers = answers[:0]
		if err := data.DB.Context(ctx).Where("status = ?", entity.AnswerStatusAvailable).
			OrderBy("id ASC").Limit(cleanupPageSize, offset).Find(&answers); err != nil {
			return fmt.Errorf("list available answers for vector cleanup: %w", err)
		}
		for _, answer := range answers {
			hasRestricted, err := hasRestrictedAnswerAggregate(ctx, data, answer.ID)
			if err != nil {
				return err
			}
			if hasRestricted {
				if err := vectorSearch.DeleteContent(ctx, uid.DeShortID(answer.ID)); err != nil {
					return fmt.Errorf("delete stale answer vector %s: %w", answer.ID, err)
				}
			}
		}
		if len(answers) < cleanupPageSize {
			break
		}
	}
	return nil
}

func hasRestrictedQuestionAggregate(ctx context.Context, data *data.Data, questionID string) (bool, error) {
	answerCount, err := data.DB.Context(ctx).Where("question_id = ? AND status <> ?", questionID,
		entity.AnswerStatusAvailable).Count(new(entity.Answer))
	if err != nil {
		return false, fmt.Errorf("count restricted answers for question %s: %w", questionID, err)
	}
	commentCount, err := data.DB.Context(ctx).Where("question_id = ? AND status <> ?", questionID,
		entity.CommentStatusAvailable).Count(new(entity.Comment))
	if err != nil {
		return false, fmt.Errorf("count restricted comments for question %s: %w", questionID, err)
	}
	return answerCount > 0 || commentCount > 0, nil
}

func hasRestrictedAnswerAggregate(ctx context.Context, data *data.Data, answerID string) (bool, error) {
	count, err := data.DB.Context(ctx).Where("object_id = ? AND status <> ?", answerID,
		entity.CommentStatusAvailable).Count(new(entity.Comment))
	if err != nil {
		return false, fmt.Errorf("count restricted comments for answer %s: %w", answerID, err)
	}
	return count > 0, nil
}
