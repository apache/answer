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
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/tag"
	"github.com/apache/answer/internal/repo/tag_common"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/apache/answer/pkg/converter"
	"github.com/stretchr/testify/assert"
)

var (
	tagOnce     sync.Once
	testTagList = []*entity.Tag{
		{
			SlugName:     "go",
			DisplayName:  "Golang",
			OriginalText: "golang",
			ParsedText:   "<p>golang</p>",
			Status:       entity.TagStatusAvailable,
		},
		{
			SlugName:     "js",
			DisplayName:  "javascript",
			OriginalText: "javascript",
			ParsedText:   "<p>javascript</p>",
			Status:       entity.TagStatusAvailable,
		},
		{
			SlugName:     "go2",
			DisplayName:  "Golang2",
			OriginalText: "golang2",
			ParsedText:   "<p>golang2</p>",
			Status:       entity.TagStatusAvailable,
		},
	}
)

func addTagList() {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, uniqueIDRepo)
	err := tagCommonRepo.AddTagList(context.TODO(), testTagList)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Test_tagRepo_GetTagByID(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[0].ID, true)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].SlugName, gotTag.SlugName)
}

func Test_tagRepo_GetTagBySlugName(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagBySlugName(context.TODO(), testTagList[0].SlugName)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].SlugName, gotTag.SlugName)
}

func Test_tagRepo_GetTagList(t *testing.T) {
	tagOnce.Do(addTagList)
	tagRepo := tag.NewTagRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagRepo.GetTagList(context.TODO(), &entity.Tag{ID: testTagList[0].ID})
	assert.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagListByIDs(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagCommonRepo.GetTagListByIDs(context.TODO(), []string{testTagList[0].ID})
	assert.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagListByName(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagCommonRepo.GetTagListByName(context.TODO(), testTagList[0].SlugName, false, false)
	assert.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagListByNames(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagCommonRepo.GetTagListByNames(context.TODO(), []string{testTagList[0].SlugName})
	assert.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagPage(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, _, err := tagCommonRepo.GetTagPage(context.TODO(), 1, 1, &entity.Tag{SlugName: testTagList[0].SlugName}, "")
	assert.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_RemoveTag(t *testing.T) {
	tagOnce.Do(addTagList)
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagRepo := tag.NewTagRepo(testDataSource, uniqueIDRepo)
	err := tagRepo.RemoveTag(context.TODO(), testTagList[1].ID)
	assert.NoError(t, err)

	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	_, exist, err := tagCommonRepo.GetTagBySlugName(context.TODO(), testTagList[1].SlugName)
	assert.NoError(t, err)
	assert.False(t, exist)
}

func Test_tagRepo_UpdateTag(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagRepo := tag.NewTagRepo(testDataSource, uniqueIDRepo)

	testTagList[0].DisplayName = "golang"
	err := tagRepo.UpdateTag(context.TODO(), testTagList[0])
	assert.NoError(t, err)

	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[0].ID, true)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].DisplayName, gotTag.DisplayName)
}

func Test_tagRepo_UpdateTagQuestionCount(t *testing.T) {
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	testTagList[0].DisplayName = "golang"
	err := tagCommonRepo.UpdateTagQuestionCount(context.TODO(), testTagList[0].ID, 100)
	assert.NoError(t, err)

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[0].ID, true)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, 100, gotTag.QuestionCount)
}

func Test_tagRepo_UpdateTagSynonym(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagRepo := tag.NewTagRepo(testDataSource, uniqueIDRepo)

	testTagList[0].DisplayName = "golang"
	err := tagRepo.UpdateTag(context.TODO(), testTagList[0])
	assert.NoError(t, err)

	err = tagRepo.UpdateTagSynonym(context.TODO(), []string{testTagList[2].SlugName},
		converter.StringToInt64(testTagList[0].ID), testTagList[0].SlugName)
	assert.NoError(t, err)

	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[2].ID, true)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].ID, fmt.Sprintf("%d", gotTag.MainTagID))
}
