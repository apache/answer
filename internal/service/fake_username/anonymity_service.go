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

	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/checker"
)

type AnonymityService struct {
	fakeUsernameRepo FakeUsernameRepo
}

func NewAnonymityService(fakeUsernameRepo FakeUsernameRepo) *AnonymityService {
	return &AnonymityService{
		fakeUsernameRepo: fakeUsernameRepo,
	}
}

func (s *AnonymityService) AnonymizeUserData(
	ctx context.Context, userIDs []string, questionID, forUserID string,
) (anonymizeInfo map[string]*schema.UserBasicInfo, err error) {
	anonymizeInfo = map[string]*schema.UserBasicInfo{}

	userIDs = checker.FilterEmptyString(userIDs)
	if len(userIDs) == 0 {
		return
	}

	filteredUserIDs := make([]string, 0, len(userIDs))
	if forUserID != "" {
		for _, id := range userIDs {
			if id == forUserID {
				continue
			}
			filteredUserIDs = append(filteredUserIDs, id)
		}
	} else {
		filteredUserIDs = userIDs
	}

	fakeUsernames, err := s.fakeUsernameRepo.BatchGetByUserIDs(ctx, filteredUserIDs, questionID)
	if err != nil {
		return
	}

	for _, item := range fakeUsernames {
		anonymizeInfo[item.UserID] = &schema.UserBasicInfo{DisplayName: item.FakeName}
	}

	return
}
