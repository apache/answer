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

package activity_common

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/unique"
	"github.com/apache/answer/pkg/obj"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// FollowRepo follow repository
type FollowRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
	activityRepo activity_common.ActivityRepo
}

// NewFollowRepo new repository
func NewFollowRepo(
	data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo,
	activityRepo activity_common.ActivityRepo,
) activity_common.FollowRepo {
	return &FollowRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
		activityRepo: activityRepo,
	}
}

// GetFollowAmount get object id's follows
func (ar *FollowRepo) GetFollowAmount(ctx context.Context, objectID string) (follows int, err error) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return 0, err
	}
	switch objectType {
	case "question":
		model := &entity.Question{}
		_, err = ar.data.DB.Context(ctx).Where("id = ?", objectID).Cols("`follow_count`").Get(model)
		if err == nil {
			follows = int(model.FollowCount)
		}
	case "user":
		model := &entity.User{}
		_, err = ar.data.DB.Context(ctx).Where("id = ?", objectID).Cols("`follow_count`").Get(model)
		if err == nil {
			follows = int(model.FollowCount)
		}
	case "tag":
		model := &entity.Tag{}
		_, err = ar.data.DB.Context(ctx).Where("id = ?", objectID).Cols("`follow_count`").Get(model)
		if err == nil {
			follows = int(model.FollowCount)
		}
	default:
		err = errors.InternalServer(reason.DisallowFollow).WithMsg("this object can't be followed")
	}

	if err != nil {
		return 0, err
	}
	return follows, nil
}

// GetFollowUserIDs get follow userID by objectID
func (ar *FollowRepo) GetFollowUserIDs(ctx context.Context, objectID string) (userIDs []string, err error) {
	objectTypeStr, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return nil, err
	}
	activityType, err := ar.activityRepo.GetActivityTypeByObjectType(ctx, objectTypeStr, "follow")
	if err != nil {
		log.Errorf("can't get activity type by object key: %s", objectTypeStr)
		return nil, err
	}

	userIDs = make([]string, 0)
	session := ar.data.DB.Context(ctx).Select("user_id")
	session.Table(entity.Activity{}.TableName())
	session.Where("object_id = ?", objectID)
	session.Where("activity_type = ?", activityType)
	session.Where("cancelled = 0")
	err = session.Find(&userIDs)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return userIDs, nil
}

// GetFollowIDs get all follow id list
func (ar *FollowRepo) GetFollowIDs(ctx context.Context, userID, objectKey string) (followIDs []string, err error) {
	followIDs = make([]string, 0)
	activityType, err := ar.activityRepo.GetActivityTypeByObjectType(ctx, objectKey, "follow")
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	session := ar.data.DB.Context(ctx).Select("object_id")
	session.Table(entity.Activity{}.TableName())
	session.Where("user_id = ? AND activity_type = ?", userID, activityType)
	session.Where("cancelled = 0")
	err = session.Find(&followIDs)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return followIDs, nil
}

// IsFollowed check user if follow object or not
func (ar *FollowRepo) IsFollowed(ctx context.Context, userID, objectID string) (followed bool, err error) {
	objectKey, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return false, err
	}

	activityType, err := ar.activityRepo.GetActivityTypeByObjectType(ctx, objectKey, "follow")
	if err != nil {
		return false, err
	}

	at := &entity.Activity{}
	has, err := ar.data.DB.Context(ctx).Where("user_id = ? AND object_id = ? AND activity_type = ?", userID, objectID, activityType).Get(at)
	if err != nil {
		return false, err
	}
	if !has {
		return false, nil
	}
	if at.Cancelled == entity.ActivityCancelled {
		return false, nil
	} else {
		return true, nil
	}
}
