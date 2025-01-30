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

package badge_award

import (
	"context"
	"fmt"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/badge"
	"github.com/apache/answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

type badgeAwardRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

func NewBadgeAwardRepo(data *data.Data, uniqueIDRepo unique.UniqueIDRepo) badge.BadgeAwardRepo {
	return &badgeAwardRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AwardBadgeForUser award badge for user
func (r *badgeAwardRepo) AwardBadgeForUser(ctx context.Context, badgeAward *entity.BadgeAward) (err error) {
	badgeAward.ID, err = r.uniqueIDRepo.GenUniqueIDStr(ctx, entity.BadgeAward{}.TableName())
	if err != nil {
		return err
	}

	_, err = r.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)

		badgeInfo := &entity.Badge{}
		exist, err := session.ID(badgeAward.BadgeID).ForUpdate().Get(badgeInfo)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, fmt.Errorf("badge not exist")
		}

		old := &entity.BadgeAward{
			UserID:         badgeAward.UserID,
			BadgeID:        badgeAward.BadgeID,
			IsBadgeDeleted: entity.IsBadgeNotDeleted,
		}
		if badgeInfo.Single != entity.BadgeSingleAward {
			old.AwardKey = badgeAward.AwardKey
		}
		exist, err = session.Get(old)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, fmt.Errorf("badge already awarded")
		}

		_, err = session.Insert(badgeAward)
		if err != nil {
			return nil, err
		}

		return session.ID(badgeInfo.ID).Incr("award_count", 1).Update(&entity.Badge{})
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// CheckIsAward check this badge is awarded for this user or not
func (r *badgeAwardRepo) CheckIsAward(ctx context.Context, badgeID, userID, awardKey string, singleOrMulti int8) (
	isAward bool, err error) {
	if singleOrMulti == entity.BadgeSingleAward {
		_, isAward, err = r.GetByUserIdAndBadgeId(ctx, userID, badgeID)
	} else {
		_, isAward, err = r.GetByUserIdAndBadgeIdAndAwardKey(ctx, userID, badgeID, awardKey)
	}
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return isAward, err
}

func (r *badgeAwardRepo) CountByUserIdAndBadgeId(ctx context.Context, userID string, badgeID string) (awardCount int64) {
	awardCount, err := r.data.DB.Context(ctx).Where("user_id = ? AND badge_id = ?", userID, badgeID).Count(&entity.BadgeAward{})
	if err != nil {
		return 0
	}
	return
}

func (r *badgeAwardRepo) CountByBadgeID(ctx context.Context, badgeID string) (awardCount int64, err error) {
	awardCount, err = r.data.DB.Context(ctx).Count(&entity.BadgeAward{BadgeID: badgeID})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (r *badgeAwardRepo) SumUserEarnedGroupByBadgeID(ctx context.Context, userID string) (earnedCounts []*entity.BadgeEarnedCount, err error) {
	err = r.data.DB.Context(ctx).Select("badge_id, count(`id`) AS earned_count").Where("user_id = ?", userID).GroupBy("badge_id").Find(&earnedCounts)
	return
}

// ListPagedByBadgeId list badge awards by badge id
func (r *badgeAwardRepo) ListPagedByBadgeId(ctx context.Context, badgeID string, page int, pageSize int) (badgeAwardList []*entity.BadgeAward, total int64, err error) {
	session := r.data.DB.Context(ctx)
	session.Where("badge_id = ?", badgeID)
	total, err = pager.Help(page, pageSize, &badgeAwardList, &entity.BadgeAward{}, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// ListPagedByBadgeIdAndUserId list badge awards by badge id and user id
func (r *badgeAwardRepo) ListPagedByBadgeIdAndUserId(ctx context.Context, badgeID string, userID string, page int, pageSize int) (badgeAwardList []*entity.BadgeAward, total int64, err error) {
	session := r.data.DB.Context(ctx)
	session.Where("badge_id = ? AND user_id = ?", badgeID, userID)
	total, err = pager.Help(page, pageSize, &badgeAwardList, &entity.Question{}, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// ListNewestEarned list newest earned badge awards
func (r *badgeAwardRepo) ListNewestEarned(ctx context.Context, userID string, limit int) (badgeAwards []*entity.BadgeAwardRecent, err error) {
	badgeAwards = make([]*entity.BadgeAwardRecent, 0)
	err = r.data.DB.Context(ctx).
		Select("badge_id, max(created_at) created,count(*) earned_count").
		Where("user_id = ? AND is_badge_deleted = ? ", userID, entity.IsBadgeNotDeleted).
		GroupBy("badge_id").
		OrderBy("created desc").
		Limit(limit).Find(&badgeAwards)
	return
}

// GetByUserIdAndBadgeId get badge award by user id and badge id
func (r *badgeAwardRepo) GetByUserIdAndBadgeId(ctx context.Context, userID string, badgeID string) (
	badgeAward *entity.BadgeAward, exists bool, err error) {
	badgeAward = &entity.BadgeAward{}
	exists, err = r.data.DB.Context(ctx).
		Where("user_id = ? AND badge_id = ? AND is_badge_deleted = 0", userID, badgeID).Get(badgeAward)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByUserIdAndBadgeIdAndAwardKey get badge award by user id and badge id and award key
func (r *badgeAwardRepo) GetByUserIdAndBadgeIdAndAwardKey(ctx context.Context, userID string, badgeID string, awardKey string) (
	badgeAward *entity.BadgeAward, exists bool, err error) {
	badgeAward = &entity.BadgeAward{}
	exists, err = r.data.DB.Context(ctx).
		Where("user_id = ? AND badge_id = ? AND award_key = ? AND is_badge_deleted = 0", userID, badgeID, awardKey).Get(badgeAward)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// DeleteUserBadgeAward delete user badge award
func (r *badgeAwardRepo) DeleteUserBadgeAward(ctx context.Context, userID string) (err error) {
	_, err = r.data.DB.Context(ctx).Where("user_id = ?", userID).Delete(&entity.BadgeAward{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
