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
	"fmt"
	"xorm.io/builder"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/rank"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/xorm"
)

// UserActiveActivityRepo answer accepted
type UserActiveActivityRepo struct {
	data          *data.Data
	activityRepo  activity_common.ActivityRepo
	userRankRepo  rank.UserRankRepo
	configService *config.ConfigService
}

const (
	UserActivated = "user.activated"
)

// NewUserActiveActivityRepo new repository
func NewUserActiveActivityRepo(
	data *data.Data,
	activityRepo activity_common.ActivityRepo,
	userRankRepo rank.UserRankRepo,
	configService *config.ConfigService,
) activity.UserActiveActivityRepo {
	return &UserActiveActivityRepo{
		data:          data,
		activityRepo:  activityRepo,
		userRankRepo:  userRankRepo,
		configService: configService,
	}
}

// UserActive user active
func (ar *UserActiveActivityRepo) UserActive(ctx context.Context, userID string) (err error) {
	cfg, err := ar.configService.GetConfigByKey(ctx, UserActivated)
	if err != nil {
		return err
	}
	addActivity := &entity.Activity{
		UserID:           userID,
		ObjectID:         "0",
		OriginalObjectID: "0",
		ActivityType:     cfg.ID,
		Rank:             cfg.GetIntValue(),
		HasRank:          1,
	}

	_, err = ar.data.DB.Transaction(func(session *xorm.Session) (result any, err error) {
		session = session.Context(ctx)

		user := &entity.User{}
		exist, err := session.ID(userID).ForUpdate().Get(user)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, fmt.Errorf("user not exist")
		}

		existsActivity := &entity.Activity{}
		exist, err = session.
			And(builder.Eq{"user_id": addActivity.UserID}).
			And(builder.Eq{"activity_type": addActivity.ActivityType}).
			Get(existsActivity)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, nil
		}

		err = ar.userRankRepo.ChangeUserRank(ctx, session, addActivity.UserID, user.Rank, addActivity.Rank)
		if err != nil {
			return nil, err
		}

		_, err = session.Insert(addActivity)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}
