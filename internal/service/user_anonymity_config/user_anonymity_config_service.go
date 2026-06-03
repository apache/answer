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

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

type UserAnonymityConfigRepo interface {
	Add(ctx context.Context, userIDs []string, enabled bool) (err error)
	Save(ctx context.Context, uc *entity.UserAnonymityConfig) (err error)
	GetByUserID(ctx context.Context, userID string) (uc *entity.UserAnonymityConfig, exists bool, err error)
}

type UserAnonymityConfigService struct {
	userAnonymityConfigRepo UserAnonymityConfigRepo
}

func NewUserAnonymityConfigService(userAnonymityConfigRepo UserAnonymityConfigRepo) *UserAnonymityConfigService {
	return &UserAnonymityConfigService{
		userAnonymityConfigRepo: userAnonymityConfigRepo,
	}
}

func (us *UserAnonymityConfigService) GetUserAnonymityConfig(ctx context.Context, userID string) (
	resp *schema.GetUserAnonymityConfigResp, err error,
) {
	anonymityConfig, exists, err := us.userAnonymityConfigRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		anonymityConfig = &entity.UserAnonymityConfig{}
	}
	resp = &schema.GetUserAnonymityConfigResp{}
	resp.UserAnonymityConfig = schema.NewUserAnonymityConfig(*anonymityConfig)
	return resp, nil
}

func (us *UserAnonymityConfigService) UpdateUserAnonymityConfig(
	ctx context.Context, req *schema.UpdateUserAnonymityConfigReq) (err error) {
	// req.Format()

	err = us.userAnonymityConfigRepo.Save(ctx, us.convertToEntity(ctx, req.UserID, req.Enabled))
	if err != nil {
		return err
	}
	return nil
}

func (us *UserAnonymityConfigService) SetDefaultUserAnonymityConfig(
	ctx context.Context, userIDs []string,
) (err error) {
	return us.userAnonymityConfigRepo.Add(ctx, userIDs, false)
}

func (us *UserAnonymityConfigService) convertToEntity(
	_ context.Context, userID string, enabled bool,
) *entity.UserAnonymityConfig {
	return &entity.UserAnonymityConfig{
		UserID:  userID,
		Enabled: enabled,
	}
}
