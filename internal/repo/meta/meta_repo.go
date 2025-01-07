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

package meta

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/meta_common"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// metaRepo meta repository
type metaRepo struct {
	data *data.Data
}

// NewMetaRepo new repository
func NewMetaRepo(data *data.Data) metacommon.MetaRepo {
	return &metaRepo{
		data: data,
	}
}

// AddMeta add meta
func (mr *metaRepo) AddMeta(ctx context.Context, meta *entity.Meta) (err error) {
	_, err = mr.data.DB.Context(ctx).Insert(meta)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveMeta delete meta
func (mr *metaRepo) RemoveMeta(ctx context.Context, id int) (err error) {
	_, err = mr.data.DB.Context(ctx).ID(id).Delete(&entity.Meta{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateMeta update meta
func (mr *metaRepo) UpdateMeta(ctx context.Context, meta *entity.Meta) (err error) {
	_, err = mr.data.DB.Context(ctx).ID(meta.ID).Update(meta)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// AddOrUpdateMetaByObjectIdAndKey if exist record with same objectID and key, update it. Or create a new one
func (mr *metaRepo) AddOrUpdateMetaByObjectIdAndKey(ctx context.Context, objectId, key string, f func(*entity.Meta, bool) (*entity.Meta, error)) error {
	_, err := mr.data.DB.Transaction(func(session *xorm.Session) (interface{}, error) {
		session = session.Context(ctx)

		// 1. acquire meta entity with target object id and key
		metaEntity := &entity.Meta{}
		exist, err := session.Where(builder.Eq{"object_id": objectId}.And(builder.Eq{"`key`": key})).ForUpdate().Get(metaEntity)
		if err != nil {
			return nil, err
		}

		meta, err := f(metaEntity, exist)
		if err != nil {
			return nil, err
		}

		// return entity.Meta
		if exist {
			_, err = session.ID(metaEntity.ID).Update(meta)
		} else {
			_, err = session.Insert(meta)
		}

		return nil, err
	})
	return err
}

// GetMetaByObjectIdAndKey get meta one
func (mr *metaRepo) GetMetaByObjectIdAndKey(ctx context.Context, objectID, key string) (
	meta *entity.Meta, exist bool, err error) {
	meta = &entity.Meta{}
	exist, err = mr.data.DB.Context(ctx).Where(builder.Eq{"object_id": objectID}.And(builder.Eq{"`key`": key})).Desc("created_at").Get(meta)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetMetaList get meta list all
func (mr *metaRepo) GetMetaList(ctx context.Context, meta *entity.Meta) (metaList []*entity.Meta, err error) {
	metaList = make([]*entity.Meta, 0)
	err = mr.data.DB.Context(ctx).Find(&metaList, meta)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
