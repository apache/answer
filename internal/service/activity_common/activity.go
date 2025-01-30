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
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity_queue"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

type ActivityRepo interface {
	GetActivityTypeByObjID(ctx context.Context, objectId string, action string) (activityType, rank int, hasRank int, err error)
	GetActivityTypeByObjectType(ctx context.Context, objectKey, action string) (activityType int, err error)
	GetActivity(ctx context.Context, session *xorm.Session, objectID, userID string, activityType int) (
		existsActivity *entity.Activity, exist bool, err error)
	GetUserActivitiesByActivityType(ctx context.Context, userID string, activityType int) (activityList []*entity.Activity, err error)
	GetUserIDObjectIDActivitySum(ctx context.Context, userID, objectID string) (int, error)
	GetActivityTypeByConfigKey(ctx context.Context, configKey string) (activityType int, err error)
	AddActivity(ctx context.Context, activity *entity.Activity) (err error)
	GetUsersWhoHasGainedTheMostReputation(
		ctx context.Context, startTime, endTime time.Time, limit int) (rankStat []*entity.ActivityUserRankStat, err error)
	GetUsersWhoHasVoteMost(
		ctx context.Context, startTime, endTime time.Time, limit int) (voteStat []*entity.ActivityUserVoteStat, err error)
}

type ActivityCommon struct {
	activityRepo         ActivityRepo
	activityQueueService activity_queue.ActivityQueueService
}

// NewActivityCommon new activity common
func NewActivityCommon(
	activityRepo ActivityRepo,
	activityQueueService activity_queue.ActivityQueueService,
) *ActivityCommon {
	activity := &ActivityCommon{
		activityRepo:         activityRepo,
		activityQueueService: activityQueueService,
	}
	activity.activityQueueService.RegisterHandler(activity.HandleActivity)
	return activity
}

// HandleActivity handle activity message
func (ac *ActivityCommon) HandleActivity(ctx context.Context, msg *schema.ActivityMsg) error {
	activityType, err := ac.activityRepo.GetActivityTypeByConfigKey(ctx, string(msg.ActivityTypeKey))
	if err != nil {
		log.Errorf("error getting activity type %s, activity type is %d", err, activityType)
		return err
	}

	act := &entity.Activity{
		UserID:           msg.UserID,
		TriggerUserID:    msg.TriggerUserID,
		ObjectID:         uid.DeShortID(msg.ObjectID),
		OriginalObjectID: uid.DeShortID(msg.OriginalObjectID),
		ActivityType:     activityType,
		Cancelled:        entity.ActivityAvailable,
	}
	if len(msg.RevisionID) > 0 {
		act.RevisionID = converter.StringToInt64(msg.RevisionID)
	}
	if err := ac.activityRepo.AddActivity(ctx, act); err != nil {
		return err
	}
	return nil
}
