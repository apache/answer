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

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity"
	"github.com/apache/answer/internal/service/activity_type"
	"github.com/apache/answer/internal/service/config"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// activityRepo activity repository
type activityRepo struct {
	data          *data.Data
	configService *config.ConfigService
}

// NewActivityRepo new repository
func NewActivityRepo(
	data *data.Data,
	configService *config.ConfigService,
) activity.ActivityRepo {
	return &activityRepo{
		data:          data,
		configService: configService,
	}
}

func (ar *activityRepo) GetObjectAllActivity(ctx context.Context, objectID string, showVote bool) (
	activityList []*entity.Activity, err error) {
	activityList = make([]*entity.Activity, 0)
	session := ar.data.DB.Context(ctx).Desc("id")

	if !showVote {
		activityTypeNotShown := ar.getAllActivityType(ctx)
		session.NotIn("activity_type", activityTypeNotShown)
	}
	err = session.Find(&activityList, &entity.Activity{OriginalObjectID: objectID})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return activityList, nil
}

func (ar *activityRepo) getAllActivityType(ctx context.Context) (activityTypes []int) {
	var activityTypeNotShown []int
	for _, key := range activity_type.VoteActivityTypeList {
		id, err := ar.configService.GetIDByKey(ctx, key)
		if err != nil {
			log.Errorf("get config id by key [%s] error: %v", key, err)
		} else {
			activityTypeNotShown = append(activityTypeNotShown, id)
		}
	}
	return activityTypeNotShown
}
