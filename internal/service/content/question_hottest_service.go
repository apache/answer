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

package content

import (
	"context"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/log"
	"math"
	"time"
)

func (q *QuestionService) RefreshHottestCron(ctx context.Context) {

	var (
		page     = 1
		pageSize = 100
	)

	for {
		questionList, _, err := q.questionRepo.GetQuestionPage(
			ctx,
			page, pageSize,
			[]string{},
			"", "newest",
			schema.HotInDays,
			false, false)
		if err != nil {
			return
		}

		for _, question := range questionList {
			updatedAt := question.UpdatedAt.Unix()
			if updatedAt < 0 {
				updatedAt = question.CreatedAt.Unix()
			}

			qAgeInHours := (time.Now().Unix() - question.CreatedAt.Unix()) / 3600
			qUpdated := (time.Now().Unix() - updatedAt) / 3600

			aScores, err := q.answerRepo.SumVotesByQuestionID(ctx, question.ID)
			if err != nil {
				aScores = 0
			}

			score := q.getScore(float64(question.ViewCount), float64(question.AnswerCount), float64(question.VoteCount), aScores, float64(qAgeInHours), float64(qUpdated))
			if score < 0 {
				score = 0
			}

			questioninfo := &entity.Question{}
			questioninfo.ID = question.ID
			questioninfo.HotScore = int(math.Ceil(score * 10000))
			err = q.questionRepo.UpdateQuestion(ctx, questioninfo, []string{"hot_score"})
			if err != nil {
				log.Error("update question hot score error,question ID:", question.ID, " error: ", err)
			}
		}

		if len(questionList) < pageSize {
			break
		}
		page++
	}
}

func (q *QuestionService) getScore(qViews, qAnswers, qScore, aScores, qAgeInHours, qUpdated float64) (score float64) {
	score = ((math.Log(qViews) * 4) + ((qAnswers * qScore) / 5) + aScores) /
		math.Pow(((qAgeInHours+1)-((qAgeInHours-qUpdated)/2)), 1.5)
	return score
}
