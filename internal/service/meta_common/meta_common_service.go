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

package metacommon

import (
	"context"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	myErrors "github.com/segmentfault/pacman/errors"
)

// MetaRepo meta repository
type MetaRepo interface {
	AddMeta(ctx context.Context, meta *entity.Meta) (err error)
	RemoveMeta(ctx context.Context, id int) (err error)
	UpdateMeta(ctx context.Context, meta *entity.Meta) (err error)
	AddOrUpdateMetaByObjectIdAndKey(ctx context.Context, objectId, key string, f func(*entity.Meta, bool) (*entity.Meta, error)) error
	GetMetaByObjectIdAndKey(ctx context.Context, objectId, key string) (meta *entity.Meta, exist bool, err error)
	GetMetaList(ctx context.Context, meta *entity.Meta) (metas []*entity.Meta, err error)
}

// MetaCommonService user service
type MetaCommonService struct {
	metaRepo MetaRepo
}

func NewMetaCommonService(metaRepo MetaRepo) *MetaCommonService {
	return &MetaCommonService{
		metaRepo: metaRepo,
	}
}

// AddMeta add meta
func (ms *MetaCommonService) AddMeta(ctx context.Context, objID, key, value string) (err error) {
	meta := &entity.Meta{
		ObjectID: objID,
		Key:      key,
		Value:    value,
	}
	return ms.metaRepo.AddMeta(ctx, meta)
}

// RemoveMeta delete meta
func (ms *MetaCommonService) RemoveMeta(ctx context.Context, id int) (err error) {
	return ms.metaRepo.RemoveMeta(ctx, id)
}

// UpdateMeta update meta
func (ms *MetaCommonService) UpdateMeta(ctx context.Context, metaID int, key, value string) (err error) {
	meta := &entity.Meta{
		ID:    metaID,
		Key:   key,
		Value: value,
	}
	return ms.metaRepo.UpdateMeta(ctx, meta)
}

func (ms *MetaCommonService) AddOrUpdateMetaByObjectIdAndKey(ctx context.Context, objID, key string, f func(*entity.Meta, bool) (*entity.Meta, error)) (err error) {
	return ms.metaRepo.AddOrUpdateMetaByObjectIdAndKey(ctx, objID, key, f)
}

// GetMetaByObjectIdAndKey get meta one
func (ms *MetaCommonService) GetMetaByObjectIdAndKey(ctx context.Context, objectID, key string) (meta *entity.Meta, err error) {
	meta, exist, err := ms.metaRepo.GetMetaByObjectIdAndKey(ctx, objectID, key)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, myErrors.BadRequest(reason.MetaObjectNotFound)
	}
	return meta, nil
}

// GetMetaList get meta list all
func (ms *MetaCommonService) GetMetaList(ctx context.Context, objID string) (metas []*entity.Meta, err error) {
	metas, err = ms.metaRepo.GetMetaList(ctx, &entity.Meta{ObjectID: objID})
	if err != nil {
		return nil, err
	}
	return metas, err
}
