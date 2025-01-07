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
	"github.com/apache/answer/internal/repo/meta"
	"github.com/stretchr/testify/assert"
)

func buildMetaEntity() *entity.Meta {
	return &entity.Meta{
		ObjectID: "1",
		Key:      "1",
		Value:    "1",
	}
}

func Test_metaRepo_GetMetaByObjectIdAndKey(t *testing.T) {
	metaRepo := meta.NewMetaRepo(testDataSource)
	metaEnt := buildMetaEntity()

	err := metaRepo.AddMeta(context.TODO(), metaEnt)
	assert.NoError(t, err)

	gotMeta, exist, err := metaRepo.GetMetaByObjectIdAndKey(context.TODO(), metaEnt.ObjectID, metaEnt.Key)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, metaEnt.ID, gotMeta.ID)

	err = metaRepo.RemoveMeta(context.TODO(), metaEnt.ID)
	assert.NoError(t, err)
}

func Test_metaRepo_GetMetaList(t *testing.T) {
	metaRepo := meta.NewMetaRepo(testDataSource)
	metaEnt := buildMetaEntity()

	err := metaRepo.AddMeta(context.TODO(), metaEnt)
	assert.NoError(t, err)

	gotMetaList, err := metaRepo.GetMetaList(context.TODO(), metaEnt)
	assert.NoError(t, err)
	assert.Equal(t, len(gotMetaList), 1)
	assert.Equal(t, gotMetaList[0].ID, metaEnt.ID)

	err = metaRepo.RemoveMeta(context.TODO(), metaEnt.ID)
	assert.NoError(t, err)
}

func Test_metaRepo_GetMetaPage(t *testing.T) {
	metaRepo := meta.NewMetaRepo(testDataSource)
	metaEnt := buildMetaEntity()

	err := metaRepo.AddMeta(context.TODO(), metaEnt)
	assert.NoError(t, err)

	gotMetaList, err := metaRepo.GetMetaList(context.TODO(), metaEnt)
	assert.NoError(t, err)
	assert.Equal(t, len(gotMetaList), 1)
	assert.Equal(t, gotMetaList[0].ID, metaEnt.ID)

	err = metaRepo.RemoveMeta(context.TODO(), metaEnt.ID)
	assert.NoError(t, err)
}

func Test_metaRepo_UpdateMeta(t *testing.T) {
	metaRepo := meta.NewMetaRepo(testDataSource)
	metaEnt := buildMetaEntity()

	err := metaRepo.AddMeta(context.TODO(), metaEnt)
	assert.NoError(t, err)

	metaEnt.Value = "testing"
	err = metaRepo.UpdateMeta(context.TODO(), metaEnt)
	assert.NoError(t, err)

	gotMeta, exist, err := metaRepo.GetMetaByObjectIdAndKey(context.TODO(), metaEnt.ObjectID, metaEnt.Key)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, gotMeta.Value, metaEnt.Value)

	err = metaRepo.RemoveMeta(context.TODO(), metaEnt.ID)
	assert.NoError(t, err)
}
