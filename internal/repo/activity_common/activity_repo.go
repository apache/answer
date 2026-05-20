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
	"fmt"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/activity_type"
	"github.com/apache/answer/pkg/obj"
	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
)

// ActivityRepo activity repository
type ActivityRepo struct {
	data          *data.Data
	uniqueIDRepo  unique.UniqueIDRepo
	configService *config.ConfigService
}

// NewActivityRepo new repository
func NewActivityRepo(
	data *data.Data,
	uniqueIDRepo unique.UniqueIDRepo,
	configService *config.ConfigService,
) activity_common.ActivityRepo {
	return &ActivityRepo{
		data:          data,
		uniqueIDRepo:  uniqueIDRepo,
		configService: configService,
	}
}

func (ar *ActivityRepo) GetActivityTypeByObjID(ctx context.Context, objectID string, action string) (
	activityType, rank, hasRank int, err error) {
	objectType, err := obj.GetObjectTypeStrByObjectID(objectID)
	if err != nil {
		return
	}

	confKey := fmt.Sprintf("%s.%s", objectType, action)
	cfg, err := ar.configService.GetConfigByKey(ctx, confKey)
	if err != nil {
		return
	}
	activityType, rank = cfg.ID, cfg.GetIntValue()
	hasRank = 0
	if rank != 0 {
		hasRank = 1
	}
	return
}

func (ar *ActivityRepo) GetActivityTypeByObjectType(ctx context.Context, objectType, action string) (activityType int, err error) {
	configKey := fmt.Sprintf("%s.%s", objectType, action)
	cfg, err := ar.configService.GetConfigByKey(ctx, configKey)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return cfg.ID, nil
}

func (ar *ActivityRepo) GetActivityTypeByConfigKey(ctx context.Context, configKey string) (activityType int, err error) {
	cfg, err := ar.configService.GetConfigByKey(ctx, configKey)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return cfg.ID, nil
}

func (ar *ActivityRepo) GetActivity(ctx context.Context, session *xorm.Session,
	objectID, userID string, activityType int,
) (existsActivity *entity.Activity, exist bool, err error) {
	existsActivity = &entity.Activity{}
	exist, err = session.
		Where(builder.Eq{"object_id": objectID}).
		And(builder.Eq{"user_id": userID}).
		And(builder.Eq{"activity_type": activityType}).
		Get(existsActivity)
	return
}

func (ar *ActivityRepo) GetUserActivitiesByActivityType(ctx context.Context, userID string, activityType int) (
	activityList []*entity.Activity, err error) {
	activityList = make([]*entity.Activity, 0)
	err = ar.data.DB.Context(ctx).Where("user_id = ?", userID).
		And("activity_type = ?", activityType).
		And("cancelled = 0").
		Find(&activityList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ar *ActivityRepo) GetUserIDObjectIDActivitySum(ctx context.Context, userID, objectID string) (int, error) {
	sum := &entity.ActivityRankSum{}
	_, err := ar.data.DB.Context(ctx).Table(entity.Activity{}.TableName()).
		Select("sum(`rank`) as `rank`").
		Where("user_id =?", userID).
		And("object_id = ?", objectID).
		And("cancelled =0").
		Get(sum)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return 0, err
	}
	return sum.Rank, nil
}

// AddActivity add activity
func (ar *ActivityRepo) AddActivity(ctx context.Context, activity *entity.Activity) (err error) {
	_, err = ar.data.DB.Context(ctx).Insert(activity)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUsersWhoHasGainedTheMostReputation get users who has gained the most reputation over a period of time
func (ar *ActivityRepo) GetUsersWhoHasGainedTheMostReputation(
	ctx context.Context, startTime, endTime time.Time, limit int) (rankStat []*entity.ActivityUserRankStat, err error) {
	rankStat = make([]*entity.ActivityUserRankStat, 0)
	session := ar.data.DB.Context(ctx).Select("user_id, SUM(`rank`) AS rank_amount").Table("activity")
	session.Where("has_rank = 1 AND cancelled = 0")
	session.Where("created_at >= ?", startTime)
	session.Where("created_at <= ?", endTime)
	session.GroupBy("user_id")
	session.Desc("rank_amount")
	session.Limit(limit)
	err = session.Find(&rankStat)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUsersWhoHasVoteMost get users who has vote most
func (ar *ActivityRepo) GetUsersWhoHasVoteMost(
	ctx context.Context, startTime, endTime time.Time, limit int) (voteStat []*entity.ActivityUserVoteStat, err error) {
	voteStat = make([]*entity.ActivityUserVoteStat, 0)

	actIDs := make([]int, 0)
	for _, act := range activity_type.ActivityTypeList {
		cfg, err := ar.configService.GetConfigByKey(ctx, act)
		if err == nil {
			actIDs = append(actIDs, cfg.ID)
		}
	}

	session := ar.data.DB.Context(ctx).Select("user_id, COUNT(*) AS vote_count").Table("activity")
	session.Where("cancelled = 0")
	session.In("activity_type", actIDs)
	session.Where("created_at >= ?", startTime)
	session.Where("created_at <= ?", endTime)
	session.GroupBy("user_id")
	session.Desc("vote_count")
	session.Limit(limit)
	err = session.Find(&voteStat)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
