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

package badge

import (
	"context"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/badge"
	"github.com/apache/answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

type badgeRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

// NewBadgeRepo creates a new badge repository
func NewBadgeRepo(data *data.Data, uniqueIDRepo unique.UniqueIDRepo) badge.BadgeRepo {
	return &badgeRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

func (r *badgeRepo) GetByID(ctx context.Context, id string) (badge *entity.Badge, exists bool, err error) {
	badge = &entity.Badge{}
	exists, err = r.data.DB.Context(ctx).Where("id = ?", id).Get(badge)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (r *badgeRepo) GetByIDs(ctx context.Context, ids []string) (badges []*entity.Badge, err error) {
	badges = make([]*entity.Badge, 0)
	err = r.data.DB.Context(ctx).In("id", ids).Find(&badges)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// ListPaged returns a list of activated badges
func (r *badgeRepo) ListPaged(ctx context.Context, page int, pageSize int) (badges []*entity.Badge, total int64, err error) {
	badges = make([]*entity.Badge, 0)
	total = 0

	session := r.data.DB.Context(ctx).Where("status <> ?", entity.BadgeStatusDeleted)
	if page == 0 || pageSize == 0 {
		err = session.Find(&badges)
	} else {
		total, err = pager.Help(page, pageSize, &badges, &entity.Badge{}, session)
	}

	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// ListActivated returns a list of activated badges
func (r *badgeRepo) ListActivated(ctx context.Context, page int, pageSize int) (badges []*entity.Badge, total int64, err error) {
	badges = make([]*entity.Badge, 0)
	total = 0

	session := r.data.DB.Context(ctx).Where("status = ?", entity.BadgeStatusActive)
	if page == 0 || pageSize == 0 {
		err = session.Find(&badges)
	} else {
		total, err = pager.Help(page, pageSize, &badges, &entity.Badge{}, session)
	}

	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// ListInactivated returns a list of inactivated badges
func (r *badgeRepo) ListInactivated(ctx context.Context, page int, pageSize int) (badges []*entity.Badge, total int64, err error) {
	badges = make([]*entity.Badge, 0)
	total = 0

	session := r.data.DB.Context(ctx).Where("status = ?", entity.BadgeStatusInactive)
	if page == 0 || pageSize == 0 {
		err = session.Find(&badges)
	} else {
		total, err = pager.Help(page, pageSize, &badges, &entity.Badge{}, session)
	}

	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateStatus updates the award count of a badge
func (r *badgeRepo) UpdateStatus(ctx context.Context, id string, status int8) (err error) {
	_, err = r.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		_, err = session.ID(id).Update(&entity.Badge{
			Status: status,
		})
		if err != nil {
			err = errors.InternalServer(reason.DatabaseError).WithError(session.Rollback()).WithStack()
			return
		}
		if status >= entity.BadgeStatusDeleted {
			_, err = session.Where("badge_id = ?", id).Cols("is_badge_deleted").Update(&entity.BadgeAward{
				IsBadgeDeleted: entity.IsBadgeDeleted,
			})
		} else {
			_, err = session.Where("badge_id = ?", id).Cols("is_badge_deleted").Update(&entity.BadgeAward{
				IsBadgeDeleted: entity.IsBadgeNotDeleted,
			})
		}
		return
	})

	return
}

// UpdateAwardCount updates the award count of a badge
func (r *badgeRepo) UpdateAwardCount(ctx context.Context, badgeID string, awardCount int) (err error) {
	_, err = r.data.DB.Context(ctx).ID(badgeID).Cols("award_count").Update(&entity.Badge{AwardCount: awardCount})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
