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
	"github.com/apache/answer/internal/repo/activity"
	"github.com/apache/answer/internal/repo/activity_common"
	"github.com/apache/answer/internal/repo/config"
	"github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/repo/tag"
	"github.com/apache/answer/internal/repo/tag_common"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/apache/answer/internal/repo/user"
	config2 "github.com/apache/answer/internal/service/config"
	"github.com/stretchr/testify/assert"
)

func Test_questionRepo_GetRecommend(t *testing.T) {
	var (
		uniqueIDRepo       = unique.NewUniqueIDRepo(testDataSource)
		questionRepo       = question.NewQuestionRepo(testDataSource, uniqueIDRepo)
		userRepo           = user.NewUserRepo(testDataSource)
		tagRelRepo         = tag.NewTagRelRepo(testDataSource, uniqueIDRepo)
		tagRepo            = tag.NewTagRepo(testDataSource, uniqueIDRepo)
		tagCommenRepo      = tag_common.NewTagCommonRepo(testDataSource, uniqueIDRepo)
		configRepo         = config.NewConfigRepo(testDataSource)
		configService      = config2.NewConfigService(configRepo)
		activityCommonRepo = activity_common.NewActivityRepo(testDataSource, uniqueIDRepo, configService)
		followRepo         = activity.NewFollowRepo(testDataSource, uniqueIDRepo, activityCommonRepo)
	)

	// create question and user
	user := &entity.User{
		Username:    "ferrischi201",
		Pass:        "ferrischi201",
		EMail:       "ferrischi201@example.com",
		MailStatus:  entity.EmailStatusAvailable,
		Status:      entity.UserStatusAvailable,
		DisplayName: "ferrischi201",
		IsAdmin:     false,
	}
	err := userRepo.AddUser(context.TODO(), user)
	assert.NoError(t, err)
	assert.NotEqual(t, "", user.ID)

	questions := make([]*entity.Question, 0)
	// tag, unjoin, unfollow
	questions = append(questions, &entity.Question{
		UserID:       "1",
		Title:        "Valid recommendation 1",
		OriginalText: "A go question",
		ParsedText:   "Go question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})
	// tag, unjoin, follow
	questions = append(questions, &entity.Question{
		UserID:       "1",
		Title:        "Valid recommendation 2",
		OriginalText: "A go question",
		ParsedText:   "Go question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})
	// tag, join, unfollow
	questions = append(questions, &entity.Question{
		UserID:       user.ID,
		Title:        "Invalid recommendation 1",
		OriginalText: "A go question 1",
		ParsedText:   "Go question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})
	// tag, join, follow
	questions = append(questions, &entity.Question{
		UserID:       user.ID,
		Title:        "Valid recommendation 3",
		OriginalText: "A java question",
		ParsedText:   "Java question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})
	// untag, unjoin, unfollow
	questions = append(questions, &entity.Question{
		UserID:       "1",
		Title:        "Invalid recommendation 2",
		OriginalText: "A go question",
		ParsedText:   "Go question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})
	// untag, unjoin, follow
	questions = append(questions, &entity.Question{
		UserID:       "1",
		Title:        "Valid recommendation 4",
		OriginalText: "A go question",
		ParsedText:   "Go question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})
	// untag, join, unfollow
	questions = append(questions, &entity.Question{
		UserID:       user.ID,
		Title:        "Invalid recommendation 3",
		OriginalText: "A go question 1",
		ParsedText:   "Go question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})
	// untag, join, follow
	questions = append(questions, &entity.Question{
		UserID:       user.ID,
		Title:        "Valid recommendation 5",
		OriginalText: "A java question",
		ParsedText:   "Java question",
		Status:       entity.QuestionStatusAvailable,
		Show:         entity.QuestionShow,
	})

	for _, question := range questions {
		err = questionRepo.AddQuestion(context.TODO(), question)
		assert.NoError(t, err)
		assert.NotEqual(t, "", question.ID)
	}

	tags := []*entity.Tag{
		{
			SlugName:     "go",
			DisplayName:  "Golang",
			OriginalText: "golang",
			ParsedText:   "<p>golang</p>",
			Status:       entity.TagStatusAvailable,
		},
		{
			SlugName:     "java",
			DisplayName:  "Java",
			OriginalText: "java",
			ParsedText:   "<p>java</p>",
			Status:       entity.TagStatusAvailable,
		},
	}
	err = tagCommenRepo.AddTagList(context.TODO(), tags)
	assert.NoError(t, err)

	tagRels := make([]*entity.TagRel, 0)
	for i, question := range questions {
		tagRel := &entity.TagRel{
			TagID:    tags[i/4].ID,
			ObjectID: question.ID,
			Status:   entity.TagRelStatusAvailable,
		}
		tagRels = append(tagRels, tagRel)
	}
	err = tagRelRepo.AddTagRelList(context.TODO(), tagRels)
	assert.NoError(t, err)

	followQuestionIDs := make([]string, 0)
	for i := range questions {
		if i%2 == 0 {
			continue
		}
		err = followRepo.Follow(context.TODO(), questions[i].ID, user.ID)
		assert.NoError(t, err)
		followQuestionIDs = append(followQuestionIDs, questions[i].ID)
	}

	// get recommend
	questionList, total, err := questionRepo.GetRecommendQuestionPageByTags(context.TODO(), user.ID, []string{tags[0].ID}, followQuestionIDs, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Equal(t, 5, len(questionList))

	// recovery
	t.Cleanup(func() {
		tagRelIDs := make([]int64, 0)
		for i, tagRel := range tagRels {
			if i%2 == 1 {
				err = followRepo.FollowCancel(context.TODO(), questions[i].ID, user.ID)
				assert.NoError(t, err)
			}
			tagRelIDs = append(tagRelIDs, tagRel.ID)
		}
		err = tagRelRepo.RemoveTagRelListByIDs(context.TODO(), tagRelIDs)
		assert.NoError(t, err)
		for _, tag := range tags {
			err = tagRepo.RemoveTag(context.TODO(), tag.ID)
			assert.NoError(t, err)
		}
		for _, q := range questions {
			err = questionRepo.RemoveQuestion(context.TODO(), q.ID)
			assert.NoError(t, err)
		}
	})
}
