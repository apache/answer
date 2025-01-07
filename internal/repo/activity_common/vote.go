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
	"github.com/apache/answer/pkg/uid"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity_common"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// VoteRepo activity repository
type VoteRepo struct {
	data         *data.Data
	activityRepo activity_common.ActivityRepo
}

// NewVoteRepo new repository
func NewVoteRepo(data *data.Data, activityRepo activity_common.ActivityRepo) activity_common.VoteRepo {
	return &VoteRepo{
		data:         data,
		activityRepo: activityRepo,
	}
}

func (vr *VoteRepo) GetVoteStatus(ctx context.Context, objectID, userID string) (status string) {
	objectID = uid.DeShortID(objectID)
	for _, action := range []string{"vote_up", "vote_down"} {
		activityType, _, _, err := vr.activityRepo.GetActivityTypeByObjID(ctx, objectID, action)
		if err != nil {
			return ""
		}
		at := &entity.Activity{}
		has, err := vr.data.DB.Context(ctx).Where("object_id = ? AND cancelled = 0 AND activity_type = ? AND user_id = ?",
			objectID, activityType, userID).Get(at)
		if err != nil {
			log.Error(err)
			return ""
		}
		if has {
			return action
		}
	}
	return ""
}

func (vr *VoteRepo) GetVoteCount(ctx context.Context, activityTypes []int) (count int64, err error) {
	list := make([]*entity.Activity, 0)
	count, err = vr.data.DB.Context(ctx).Where("cancelled =0").In("activity_type", activityTypes).FindAndCount(&list)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
