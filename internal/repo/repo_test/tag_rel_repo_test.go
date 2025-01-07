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
	"log"
	"sync"
	"testing"

	"github.com/apache/answer/internal/repo/unique"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/tag"
	"github.com/stretchr/testify/assert"
)

var (
	tagRelOnce     sync.Once
	testTagRelList = []*entity.TagRel{
		{
			ObjectID: "10010000000000101",
			TagID:    "10030000000000101",
			Status:   entity.TagRelStatusAvailable,
		},
		{
			ObjectID: "10010000000000202",
			TagID:    "10030000000000202",
			Status:   entity.TagRelStatusAvailable,
		},
	}
)

func addTagRelList() {
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	err := tagRelRepo.AddTagRelList(context.TODO(), testTagRelList)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Test_tagListRepo_BatchGetObjectTagRelList(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	relList, err :=
		tagRelRepo.BatchGetObjectTagRelList(context.TODO(), []string{testTagRelList[0].ObjectID, testTagRelList[1].ObjectID})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(relList))
}

func Test_tagListRepo_CountTagRelByTagID(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	count, err := tagRelRepo.CountTagRelByTagID(context.TODO(), "10030000000000101")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func Test_tagListRepo_GetObjectTagRelList(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	relList, err :=
		tagRelRepo.GetObjectTagRelList(context.TODO(), testTagRelList[0].ObjectID)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(relList))
}

func Test_tagListRepo_GetObjectTagRelWithoutStatus(t *testing.T) {
	tagRelOnce.Do(addTagRelList)
	tagRelRepo := tag.NewTagRelRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	relList, err :=
		tagRelRepo.BatchGetObjectTagRelList(context.TODO(), []string{testTagRelList[0].ObjectID, testTagRelList[1].ObjectID})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(relList))

	ids := []int64{relList[0].ID, relList[1].ID}
	err = tagRelRepo.RemoveTagRelListByIDs(context.TODO(), ids)
	assert.NoError(t, err)

	count, err := tagRelRepo.CountTagRelByTagID(context.TODO(), "10030000000000101")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	_, exist, err := tagRelRepo.GetObjectTagRelWithoutStatus(context.TODO(), relList[0].ObjectID, relList[0].TagID)
	assert.NoError(t, err)
	assert.True(t, exist)

	err = tagRelRepo.EnableTagRelByIDs(context.TODO(), ids, false)
	assert.NoError(t, err)

	count, err = tagRelRepo.CountTagRelByTagID(context.TODO(), "10030000000000101")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
