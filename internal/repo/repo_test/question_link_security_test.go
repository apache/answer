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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/stretchr/testify/require"
)

func TestQuestionRepoGetQuestionLinkRespectsPendingVisibility(t *testing.T) {
	ctx := context.Background()
	questionRepo := question.NewQuestionRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	newQuestion := func(userID, title string, status, show int) *entity.Question {
		q := &entity.Question{
			UserID:       userID,
			Title:        title,
			OriginalText: title,
			ParsedText:   title,
			Status:       status,
			Show:         show,
		}
		require.NoError(t, questionRepo.AddQuestion(ctx, q))
		return q
	}

	target := newQuestion("link-target-owner", "link target", entity.QuestionStatusAvailable, entity.QuestionShow)
	available := newQuestion("link-author", "available", entity.QuestionStatusAvailable, entity.QuestionShow)
	pendingOwner := newQuestion("link-author", "pending owner", entity.QuestionStatusPending, entity.QuestionShow)
	pendingOther := newQuestion("link-other", "pending other", entity.QuestionStatusPending, entity.QuestionShow)
	hidden := newQuestion("link-author", "hidden", entity.QuestionStatusAvailable, entity.QuestionHide)
	questions := []*entity.Question{target, available, pendingOwner, pendingOther, hidden}

	for _, from := range questions[1:] {
		_, err := testDataSource.DB.Context(ctx).Insert(&entity.QuestionLink{
			FromQuestionID: from.ID,
			ToQuestionID:   target.ID,
			Status:         entity.QuestionLinkStatusAvailable,
		})
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		_, _ = testDataSource.DB.Context(ctx).Where("to_question_id = ?", target.ID).Delete(&entity.QuestionLink{})
		for _, q := range questions {
			_, _ = testDataSource.DB.Context(ctx).ID(q.ID).Delete(&entity.Question{})
		}
	})

	assertLinkedIDs := func(loginUserID string, isAdminModerator bool, want ...string) {
		got, _, err := questionRepo.GetQuestionLink(ctx, 1, 20, target.ID, loginUserID, isAdminModerator, "newest", 0)
		require.NoError(t, err)
		gotIDs := make([]string, 0, len(got))
		for _, q := range got {
			gotIDs = append(gotIDs, q.ID)
		}
		require.ElementsMatch(t, want, gotIDs)
	}

	assertLinkedIDs("", false, available.ID)
	assertLinkedIDs("link-author", false, available.ID, pendingOwner.ID)
	assertLinkedIDs("link-other", false, available.ID, pendingOther.ID)
	assertLinkedIDs("link-moderator", true, available.ID, pendingOwner.ID, pendingOther.ID)
}
