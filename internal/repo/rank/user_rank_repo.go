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

package rank

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/rank"
	"github.com/apache/answer/plugin"
	"github.com/jinzhu/now"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// UserRankRepo user rank repository
type UserRankRepo struct {
	data          *data.Data
	configService *config.ConfigService
}

// NewUserRankRepo new repository
func NewUserRankRepo(data *data.Data, configService *config.ConfigService) rank.UserRankRepo {
	return &UserRankRepo{
		data:          data,
		configService: configService,
	}
}

func (ur *UserRankRepo) GetMaxDailyRank(ctx context.Context) (maxDailyRank int, err error) {
	maxDailyRank, err = ur.configService.GetIntValue(ctx, "daily_rank_limit")
	if err != nil {
		return 0, err
	}
	return maxDailyRank, nil
}

func (ur *UserRankRepo) CheckReachLimit(ctx context.Context, session *xorm.Session,
	userID string, maxDailyRank int) (
	reach bool, err error) {
	session.Where(builder.Eq{"user_id": userID})
	session.Where(builder.Eq{"cancelled": 0})
	session.Where(builder.Between{
		Col:     "updated_at",
		LessVal: now.BeginningOfDay(),
		MoreVal: now.EndOfDay(),
	})

	earned, err := session.SumInt(&entity.Activity{}, "`rank`")
	if err != nil {
		return false, err
	}
	if int(earned) < maxDailyRank {
		return false, nil
	}
	log.Infof("user %s today has rank %d is reach stand %d", userID, earned, maxDailyRank)
	return true, nil
}

// ChangeUserRank change user rank
func (ur *UserRankRepo) ChangeUserRank(
	ctx context.Context, session *xorm.Session, userID string, userCurrentScore, deltaRank int) (err error) {
	// IMPORTANT: If user center enabled the rank agent, then we should not change user rank.
	if plugin.RankAgentEnabled() || deltaRank == 0 {
		return nil
	}

	// If user rank is lower than 1 after this action, then user rank will be set to 1 only.
	if deltaRank < 0 && userCurrentScore+deltaRank < 1 {
		deltaRank = 1 - userCurrentScore
	}

	_, err = session.ID(userID).Incr("`rank`", deltaRank).Update(&entity.User{})
	if err != nil {
		return err
	}
	return nil
}

// TriggerUserRank trigger user rank change
// session is need provider, it means this action must be success or failure
// if outer action is failed then this action is need rollback
func (ur *UserRankRepo) TriggerUserRank(ctx context.Context,
	session *xorm.Session, userID string, deltaRank int, activityType int,
) (isReachStandard bool, err error) {
	// IMPORTANT: If user center enabled the rank agent, then we should not change user rank.
	if plugin.RankAgentEnabled() || deltaRank == 0 {
		return false, nil
	}

	if deltaRank < 0 {
		// if user rank is lower than 1 after this action, then user rank will be set to 1 only.
		var isReachMin bool
		isReachMin, err = ur.checkUserMinRank(ctx, session, userID, deltaRank)
		if err != nil {
			return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		if isReachMin {
			_, err = session.Where(builder.Eq{"id": userID}).Update(&entity.User{Rank: 1})
			if err != nil {
				return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
			}
			return true, nil
		}
	} else {
		isReachStandard, err = ur.checkUserTodayRank(ctx, session, userID, activityType)
		if err != nil {
			return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		if isReachStandard {
			return isReachStandard, nil
		}
	}
	_, err = session.Where(builder.Eq{"id": userID}).Incr("`rank`", deltaRank).Update(&entity.User{})
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return false, nil
}

func (ur *UserRankRepo) checkUserMinRank(ctx context.Context, session *xorm.Session, userID string, deltaRank int) (
	isReachStandard bool, err error,
) {
	bean := &entity.User{ID: userID}
	_, err = session.Select("`rank`").Get(bean)
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if bean.Rank+deltaRank < 1 {
		log.Infof("user %s is rank %d out of range before rank operation", userID, deltaRank)
		return true, nil
	}
	return
}

func (ur *UserRankRepo) checkUserTodayRank(ctx context.Context,
	session *xorm.Session, userID string, activityType int,
) (isReachStandard bool, err error) {
	// exclude daily rank
	exclude, _ := ur.configService.GetArrayStringValue(ctx, "daily_rank_limit.exclude")
	for _, item := range exclude {
		cfg, err := ur.configService.GetConfigByKey(ctx, item)
		if err != nil {
			return false, err
		}
		if activityType == cfg.ID {
			return false, nil
		}
	}

	// get user
	start, end := now.BeginningOfDay(), now.EndOfDay()
	session.Where(builder.Eq{"user_id": userID})
	session.Where(builder.Eq{"cancelled": 0})
	session.Where(builder.Between{
		Col:     "updated_at",
		LessVal: start,
		MoreVal: end,
	})
	earned, err := session.SumInt(&entity.Activity{}, "`rank`")
	if err != nil {
		return false, err
	}

	// max rank
	maxDailyRank, err := ur.configService.GetIntValue(ctx, "daily_rank_limit")
	if err != nil {
		return false, err
	}

	if int(earned) < maxDailyRank {
		return false, nil
	}
	log.Infof("user %s today has rank %d is reach stand %d", userID, earned, maxDailyRank)
	return true, nil
}

func (ur *UserRankRepo) UserRankPage(ctx context.Context, userID string, page, pageSize int) (
	rankPage []*entity.Activity, total int64, err error,
) {
	rankPage = make([]*entity.Activity, 0)

	session := ur.data.DB.Context(ctx).Where(builder.Eq{"has_rank": 1}.And(builder.Eq{"cancelled": 0})).And(builder.Gt{"`rank`": 0})
	session.Desc("created_at")

	cond := &entity.Activity{UserID: userID}
	total, err = pager.Help(page, pageSize, &rankPage, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
