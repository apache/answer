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

package tag

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/tag_common"
	"github.com/apache/answer/internal/service/unique"
	"github.com/apache/answer/pkg/converter"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

// tagRepo tag repository
type tagRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

// NewTagRepo new repository
func NewTagRepo(
	data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo,
) tag_common.TagRepo {
	return &tagRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// RemoveTag delete tag
func (tr *tagRepo) RemoveTag(ctx context.Context, tagID string) (err error) {
	session := tr.data.DB.Context(ctx).Where(builder.Eq{"id": tagID})
	_, err = session.Update(&entity.Tag{Status: entity.TagStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateTag update tag
func (tr *tagRepo) UpdateTag(ctx context.Context, tag *entity.Tag) (err error) {
	_, err = tr.data.DB.Context(ctx).Where(builder.Eq{"id": tag.ID}).Update(tag)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RecoverTag recover deleted tag
func (tr *tagRepo) RecoverTag(ctx context.Context, tagID string) (err error) {
	_, err = tr.data.DB.Context(ctx).ID(tagID).Update(&entity.Tag{Status: entity.TagStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// MustGetTagByNameOrID get tag by name or id
func (tr *tagRepo) MustGetTagByNameOrID(ctx context.Context, tagID, slugName string) (
	tag *entity.Tag, exist bool, err error) {
	if len(tagID) == 0 && len(slugName) == 0 {
		return nil, false, nil
	}
	tag = &entity.Tag{}
	session := tr.data.DB.Context(ctx)
	if len(tagID) > 0 {
		session.ID(tagID)
	}
	if len(slugName) > 0 {
		session.Where(builder.Eq{"slug_name": slugName})
	}
	exist, err = session.Get(tag)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateTagSynonym update synonym tag
func (tr *tagRepo) UpdateTagSynonym(ctx context.Context, tagSlugNameList []string, mainTagID int64,
	mainTagSlugName string,
) (err error) {
	bean := &entity.Tag{MainTagID: mainTagID, MainTagSlugName: mainTagSlugName}
	session := tr.data.DB.Context(ctx).In("slug_name", tagSlugNameList).MustCols("main_tag_id", "main_tag_slug_name")
	_, err = session.Update(bean)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagRepo) GetTagSynonymCount(ctx context.Context, tagID string) (count int64, err error) {
	count, err = tr.data.DB.Context(ctx).Count(&entity.Tag{MainTagID: converter.StringToInt64(tagID), Status: entity.TagStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagRepo) GetIDsByMainTagId(ctx context.Context, mainTagID string) (tagIDs []string, err error) {
	session := tr.data.DB.Context(ctx).Table(entity.Tag{}.TableName()).Where(builder.Eq{"status": entity.TagStatusAvailable, "main_tag_id": converter.StringToInt64(mainTagID)}).Cols("id")
	err = session.Find(&tagIDs)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagList get tag list all
func (tr *tagRepo) GetTagList(ctx context.Context, tag *entity.Tag) (tagList []*entity.Tag, err error) {
	tagList = make([]*entity.Tag, 0)
	session := tr.data.DB.Context(ctx).Where(builder.Eq{"status": entity.TagStatusAvailable})
	err = session.Find(&tagList, tag)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
