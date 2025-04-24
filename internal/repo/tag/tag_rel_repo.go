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
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	tagcommon "github.com/apache/answer/internal/service/tag_common"
	"github.com/apache/answer/internal/service/unique"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// tagRelRepo tag rel repository
type tagRelRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

// NewTagRelRepo new repository
func NewTagRelRepo(data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo) tagcommon.TagRelRepo {
	return &tagRelRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddTagRelList add tag list
func (tr *tagRelRepo) AddTagRelList(ctx context.Context, tagList []*entity.TagRel) (err error) {
	for _, item := range tagList {
		item.ObjectID = uid.DeShortID(item.ObjectID)
	}
	_, err = tr.data.DB.Context(ctx).Insert(tagList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range tagList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// RemoveTagRelListByObjectID delete tag list
func (tr *tagRelRepo) RemoveTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.data.DB.Context(ctx).Where("object_id = ?", objectID).Update(&entity.TagRel{Status: entity.TagRelStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RecoverTagRelListByObjectID recover tag list
func (tr *tagRelRepo) RecoverTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.data.DB.Context(ctx).Where("object_id = ?", objectID).Update(&entity.TagRel{Status: entity.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagRelRepo) HideTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.data.DB.Context(ctx).Where("object_id = ?", objectID).And("status = ?", entity.TagRelStatusAvailable).Cols("status").Update(&entity.TagRel{Status: entity.TagRelStatusHide})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (tr *tagRelRepo) ShowTagRelListByObjectID(ctx context.Context, objectID string) (err error) {
	objectID = uid.DeShortID(objectID)
	_, err = tr.data.DB.Context(ctx).Where("object_id = ?", objectID).And("status = ?", entity.TagRelStatusHide).Cols("status").Update(&entity.TagRel{Status: entity.TagRelStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// RemoveTagRelListByIDs delete tag list
func (tr *tagRelRepo) RemoveTagRelListByIDs(ctx context.Context, ids []int64) (err error) {
	_, err = tr.data.DB.Context(ctx).In("id", ids).Update(&entity.TagRel{Status: entity.TagRelStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetObjectTagRelWithoutStatus get object tag relation no matter status
func (tr *tagRelRepo) GetObjectTagRelWithoutStatus(ctx context.Context, objectID, tagID string) (
	tagRel *entity.TagRel, exist bool, err error,
) {
	objectID = uid.DeShortID(objectID)
	tagRel = &entity.TagRel{}
	session := tr.data.DB.Context(ctx).Where("object_id = ?", objectID).And("tag_id = ?", tagID)
	exist, err = session.Get(tagRel)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if handler.GetEnableShortID(ctx) {
		tagRel.ObjectID = uid.EnShortID(tagRel.ObjectID)
	}
	return
}

// EnableTagRelByIDs update tag status to available
func (tr *tagRelRepo) EnableTagRelByIDs(ctx context.Context, ids []int64, hide bool) (err error) {
	status := entity.TagRelStatusAvailable
	if hide {
		status = entity.TagRelStatusHide
	}
	_, err = tr.data.DB.Context(ctx).In("id", ids).Update(&entity.TagRel{Status: status})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetObjectTagRelList get object tag relation list all
func (tr *tagRelRepo) GetObjectTagRelList(ctx context.Context, objectID string) (tagListList []*entity.TagRel, err error) {
	objectID = uid.DeShortID(objectID)
	tagListList = make([]*entity.TagRel, 0)
	session := tr.data.DB.Context(ctx).Where("object_id = ?", objectID)
	session.In("status", []int{entity.TagRelStatusAvailable, entity.TagRelStatusHide})
	err = session.Find(&tagListList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range tagListList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// BatchGetObjectTagRelList get object tag relation list all
func (tr *tagRelRepo) BatchGetObjectTagRelList(ctx context.Context, objectIds []string) (tagListList []*entity.TagRel, err error) {
	for num, item := range objectIds {
		objectIds[num] = uid.DeShortID(item)
	}
	tagListList = make([]*entity.TagRel, 0)
	session := tr.data.DB.Context(ctx).In("object_id", objectIds)
	session.Where("status = ?", entity.TagRelStatusAvailable)
	err = session.Find(&tagListList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if handler.GetEnableShortID(ctx) {
		for _, item := range tagListList {
			item.ObjectID = uid.EnShortID(item.ObjectID)
		}
	}
	return
}

// CountTagRelByTagID count tag relation
func (tr *tagRelRepo) CountTagRelByTagID(ctx context.Context, tagID string) (count int64, err error) {
	count, err = tr.data.DB.Context(ctx).Count(&entity.TagRel{TagID: tagID, Status: entity.AnswerStatusAvailable})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetTagRelDefaultStatusByObjectID get tag rel default status
func (tr *tagRelRepo) GetTagRelDefaultStatusByObjectID(ctx context.Context, objectID string) (status int, err error) {
	question := entity.Question{}
	exist, err := tr.data.DB.Context(ctx).ID(objectID).Cols("show", "status").Get(&question)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist && (question.Show == entity.QuestionHide || question.Status == entity.QuestionStatusDeleted) {
		return entity.TagRelStatusHide, nil
	}
	return entity.TagRelStatusAvailable, nil
}

// MigrateTagObjects migrate tag objects
func (tr *tagRelRepo) MigrateTagObjects(ctx context.Context, sourceTagId, targetTagId string) error {
	_, err := tr.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		// 1. Get all objects related to source tag
		var sourceObjects []entity.TagRel
		err = session.Where("tag_id = ?", sourceTagId).Find(&sourceObjects)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}

		// 2. Get existing target tag relations
		var existingTargets []entity.TagRel
		err = session.Where("tag_id = ?", targetTagId).Find(&existingTargets)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}

		// Create map of existing target objects for quick lookup
		existingMap := make(map[string]bool)
		for _, target := range existingTargets {
			existingMap[target.ObjectID] = true
		}

		// 3. Create new relations for objects not already tagged with target
		newRelations := make([]*entity.TagRel, 0)
		for _, source := range sourceObjects {
			if !existingMap[source.ObjectID] {
				newRelations = append(newRelations, &entity.TagRel{
					TagID:    targetTagId,
					ObjectID: source.ObjectID,
					Status:   source.Status,
				})
			}
		}

		if len(newRelations) > 0 {
			_, err = session.Insert(newRelations)
			if err != nil {
				return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
			}
		}

		// 4. Remove old relations
		_, err = session.Where("tag_id = ?", sourceTagId).Delete(&entity.TagRel{})
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}

		return nil, nil
	})

	return err
}
