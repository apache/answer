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

package fake_username

import (
	"context"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/user_anonymity_config"
	"github.com/segmentfault/pacman/log"
)

type FakeUsernameRepo interface {
	Add(ctx context.Context, userID, questionID string, fakeName string) (err error)
	GetByUserIDAndQuestionID(ctx context.Context, userID, questionID string) (fu *entity.FakeUsername, exist bool, err error)
	BatchGetByUserIDs(ctx context.Context, userIDs []string, questionID string) ([]*entity.FakeUsername, error)
}

type FakeUsernameService struct {
	fakeUsernameGenerator   *FakeUsernameGenerator
	fakeUsernameRepo        FakeUsernameRepo
	userAnonymityConfigRepo user_anonymity_config.UserAnonymityConfigRepo
}

func NewFakeUsernameService(
	fakeUsernameGenerator *FakeUsernameGenerator,
	fakeUsernameRepo FakeUsernameRepo,
	userAnonymityConfigRepo user_anonymity_config.UserAnonymityConfigRepo,
) *FakeUsernameService {
	return &FakeUsernameService{
		fakeUsernameGenerator:   fakeUsernameGenerator,
		fakeUsernameRepo:        fakeUsernameRepo,
		userAnonymityConfigRepo: userAnonymityConfigRepo,
	}
}

// func (fs *FakeUsernameService) AddFakeUsernameFor(ctx context.Context, userID, questionID string) (err error) {
// 	return fs.fakeUsernameRepo.Add(ctx, userID, questionID, fs.fakeUsernameGenerator.GenerateFakeName())
// }

func (fs *FakeUsernameService) AddFakeUsernameIfNeeded(ctx context.Context, userID, questionID string) (err error) {
	userAnonymityConfig, _, err := fs.userAnonymityConfigRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to get user anonymity config: %w", err)
	}

	if userAnonymityConfig.Enabled {
		return fs.fakeUsernameRepo.Add(ctx, userID, questionID, fs.fakeUsernameGenerator.GenerateFakeName())
	}

	return
}
