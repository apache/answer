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

package user_anonymity_config

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/user_anonymity_config"
	"github.com/segmentfault/pacman/errors"
)

type userAnonymityConfigRepo struct {
	data *data.Data
}

func NewUserAnonymityConfigRepo(data *data.Data) user_anonymity_config.UserAnonymityConfigRepo {
	return &userAnonymityConfigRepo{
		data: data,
	}
}

func (ur *userAnonymityConfigRepo) Add(ctx context.Context, userIDs []string, enabled bool) (err error) {
	var configs []*entity.UserAnonymityConfig
	for _, userID := range userIDs {
		configs = append(configs, &entity.UserAnonymityConfig{
			UserID:  userID,
			Enabled: enabled,
		})
	}
	_, err = ur.data.DB.Context(ctx).Insert(configs)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *userAnonymityConfigRepo) Save(ctx context.Context, uc *entity.UserAnonymityConfig) (err error) {
	old := &entity.UserAnonymityConfig{UserID: uc.UserID}
	exist, err := ur.data.DB.Context(ctx).Get(old)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		old.Enabled = uc.Enabled
		_, err = ur.data.DB.Context(ctx).ID(old.ID).UseBool("enabled").Cols("enabled").Update(old)
	} else {
		_, err = ur.data.DB.Context(ctx).Insert(uc)
	}
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// GetByUserID get anonymity config by user id
func (ur *userAnonymityConfigRepo) GetByUserID(ctx context.Context, userID string) (
	uc *entity.UserAnonymityConfig, exist bool, err error) {
	uc = &entity.UserAnonymityConfig{UserID: userID}
	exist, err = ur.data.DB.Context(ctx).Get(uc)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
