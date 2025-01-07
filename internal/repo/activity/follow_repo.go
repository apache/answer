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

package activity

import (
	"context"
	"time"

	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/follow"
	"github.com/apache/answer/pkg/obj"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// FollowRepo activity repository
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
) follow.FollowRepo {
	return &FollowRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
		activityRepo: activityRepo,
	}
}

func (ar *FollowRepo) Follow(ctx context.Context, objectID, userID string) error {
	objectTypeStr, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	activityType, err := ar.activityRepo.GetActivityTypeByObjectType(ctx, objectTypeStr, "follow")
	if err != nil {
		return err
	}

	_, err = ar.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)
		var (
			existsActivity entity.Activity
			has            bool
		)
		result = nil

		has, err = session.Where(builder.Eq{"activity_type": activityType}).
			And(builder.Eq{"user_id": userID}).
			And(builder.Eq{"object_id": objectID}).
			Get(&existsActivity)

		if err != nil {
			return
		}

		if has && existsActivity.Cancelled == entity.ActivityAvailable {
			return
		}

		if has {
			_, err = session.Where(builder.Eq{"id": existsActivity.ID}).
				Cols(`cancelled`).
				Update(&entity.Activity{
					Cancelled: entity.ActivityAvailable,
				})
		} else {
			// update existing activity with new user id and u object id
			_, err = session.Insert(&entity.Activity{
				UserID:           userID,
				ObjectID:         objectID,
				OriginalObjectID: objectID,
				ActivityType:     activityType,
				Cancelled:        entity.ActivityAvailable,
				Rank:             0,
				HasRank:          0,
			})
		}

		if err != nil {
			log.Error(err)
			return
		}

		// start update followers when everything is fine
		err = ar.updateFollows(ctx, session, objectID, 1)
		if err != nil {
			log.Error(err)
		}

		return
	})

	return err
}

func (ar *FollowRepo) FollowCancel(ctx context.Context, objectID, userID string) error {
	objectTypeStr, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	activityType, err := ar.activityRepo.GetActivityTypeByObjectType(ctx, objectTypeStr, "follow")
	if err != nil {
		return err
	}

	_, err = ar.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)
		var (
			existsActivity entity.Activity
			has            bool
		)
		result = nil

		has, err = session.Where(builder.Eq{"activity_type": activityType}).
			And(builder.Eq{"user_id": userID}).
			And(builder.Eq{"object_id": objectID}).
			Get(&existsActivity)

		if err != nil || !has {
			return
		}

		if has && existsActivity.Cancelled == entity.ActivityCancelled {
			return
		}
		if _, err = session.Where("id = ?", existsActivity.ID).
			Cols("cancelled").
			Update(&entity.Activity{
				Cancelled:   entity.ActivityCancelled,
				CancelledAt: time.Now(),
			}); err != nil {
			return
		}
		err = ar.updateFollows(ctx, session, objectID, -1)
		return
	})
	return err
}

func (ar *FollowRepo) updateFollows(ctx context.Context, session *xorm.Session, objectID string, follows int) error {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return err
	}
	switch objectType {
	case "question":
		_, err = session.Where("id = ?", objectID).Incr("follow_count", follows).Update(&entity.Question{})
	case "user":
		_, err = session.Where("id = ?", objectID).Incr("follow_count", follows).Update(&entity.User{})
	case "tag":
		_, err = session.Where("id = ?", objectID).Incr("follow_count", follows).Update(&entity.Tag{})
	default:
		err = errors.InternalServer(reason.DisallowFollow).WithMsg("this object can't be followed")
	}
	return err
}
