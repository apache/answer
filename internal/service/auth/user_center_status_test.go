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

package auth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	authrepo "github.com/apache/answer/internal/repo/auth"
	authservice "github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/plugin"
)

func TestGetUserCacheInfoChecksUserCenterStatus(t *testing.T) {
	cache, cleanup, err := data.NewCache(&data.CacheConf{})
	if err != nil {
		t.Fatalf("create cache: %v", err)
	}
	t.Cleanup(cleanup)

	previousCallUserCenter := plugin.CallUserCenter
	testUserCenter := &authTestUserCenter{}
	plugin.CallUserCenter = plugin.CallFn[plugin.UserCenter](func(fn plugin.Caller[plugin.UserCenter]) error {
		return fn(testUserCenter)
	})
	t.Cleanup(func() { plugin.CallUserCenter = previousCallUserCenter })

	service := authservice.NewAuthService(authrepo.NewAuthRepo(&data.Data{Cache: cache}), nil)
	for _, expectedStatus := range []plugin.UserStatus{plugin.UserStatusSuspended, plugin.UserStatusDeleted} {
		t.Run(fmt.Sprintf("status_%d", expectedStatus), func(t *testing.T) {
			testUserCenter.status = expectedStatus
			testUserCenter.externalID = ""
			accessToken, _, err := service.SetUserCacheInfo(context.Background(), &entity.UserCacheInfo{
				UserID:      "user-center-user",
				UserStatus:  entity.UserStatusAvailable,
				EmailStatus: entity.EmailStatusAvailable,
				ExternalID:  "central-user-1001",
			})
			if err != nil {
				t.Fatalf("cache user session: %v", err)
			}

			userInfo, err := service.GetUserCacheInfo(context.Background(), accessToken)
			if err != nil {
				t.Fatalf("get user cache info: %v", err)
			}
			if userInfo.UserStatus != int(expectedStatus) {
				t.Fatalf("user status = %d, want %d", userInfo.UserStatus, expectedStatus)
			}
			if testUserCenter.externalID != "central-user-1001" {
				t.Fatalf("user center checked external ID = %q, want %q", testUserCenter.externalID, "central-user-1001")
			}
		})
	}
}

type authTestUserCenter struct {
	externalID string
	status     plugin.UserStatus
}

func (*authTestUserCenter) Info() plugin.Info { return plugin.Info{SlugName: "auth-test-user-center"} }

func (*authTestUserCenter) Description() plugin.UserCenterDesc { return plugin.UserCenterDesc{} }

func (*authTestUserCenter) ControlCenterItems() []plugin.ControlCenter { return nil }

func (*authTestUserCenter) LoginCallback(*plugin.GinContext) (*plugin.UserCenterBasicUserInfo, error) {
	return nil, nil
}

func (*authTestUserCenter) SignUpCallback(*plugin.GinContext) (*plugin.UserCenterBasicUserInfo, error) {
	return nil, nil
}

func (*authTestUserCenter) UserInfo(string) (*plugin.UserCenterBasicUserInfo, error) { return nil, nil }

func (u *authTestUserCenter) UserStatus(externalID string) plugin.UserStatus {
	u.externalID = externalID
	return u.status
}

func (*authTestUserCenter) UserList([]string) ([]*plugin.UserCenterBasicUserInfo, error) {
	return nil, nil
}

func (*authTestUserCenter) UserSettings(string) (*plugin.SettingInfo, error) { return nil, nil }

func (*authTestUserCenter) PersonalBranding(string) []*plugin.PersonalBranding { return nil }

func (*authTestUserCenter) AfterLogin(string, string) {}
