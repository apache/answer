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
	"path/filepath"
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	authrepo "github.com/apache/answer/internal/repo/auth"
	authservice "github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/plugin"
	"github.com/stretchr/testify/require"
)

func TestGetUserCacheInfoByUserIDUsesActiveUserCenterBinding(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		bindings   []entity.UserExternalLogin
		externalID string
	}{
		{name: "no user center", bindings: []entity.UserExternalLogin{{Provider: "github", ExternalID: "github-user"}}},
		{name: "no binding", enabled: true},
		{name: "github only", enabled: true, bindings: []entity.UserExternalLogin{{Provider: "github", ExternalID: "github-user"}}},
		{name: "inactive user center", enabled: true, bindings: []entity.UserExternalLogin{{Provider: "inactive-user-center", ExternalID: "inactive-user"}}},
		{name: "active user center", enabled: true, bindings: []entity.UserExternalLogin{{Provider: "auth-test-user-center", ExternalID: "central-user"}}, externalID: "central-user"},
		{name: "multiple providers", enabled: true, bindings: []entity.UserExternalLogin{
			{Provider: "github", ExternalID: "github-user"},
			{Provider: "auth-test-user-center", ExternalID: "central-user"},
		}, externalID: "central-user"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := data.NewDB(false, &data.Database{Driver: "sqlite3", Connection: filepath.Join(t.TempDir(), "auth.db")})
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			require.NoError(t, db.Sync2(new(entity.User), new(entity.UserRoleRel), new(entity.UserExternalLogin)))
			_, err = db.Insert(&entity.User{ID: "1", Status: entity.UserStatusAvailable, MailStatus: entity.EmailStatusAvailable})
			require.NoError(t, err)
			// A matching provider binding belonging to another user must also be ignored.
			_, err = db.Insert(&entity.UserExternalLogin{UserID: "2", Provider: "auth-test-user-center", ExternalID: "another-user"})
			require.NoError(t, err)
			for _, binding := range tt.bindings {
				binding.UserID = "1"
				_, err = db.Insert(&binding)
				require.NoError(t, err)
			}
			previous := plugin.CallUserCenter
			center := &authTestUserCenter{status: plugin.UserStatusSuspended}
			plugin.CallUserCenter = func(call plugin.Caller[plugin.UserCenter]) error {
				if tt.enabled {
					return call(center)
				}
				return nil
			}
			t.Cleanup(func() { plugin.CallUserCenter = previous })
			service := authservice.NewAuthService(authrepo.NewAuthRepo(&data.Data{DB: db}), nil)
			user, err := service.GetUserCacheInfoByUserID(t.Context(), "1")
			require.NoError(t, err)
			require.NotNil(t, user)
			require.Equal(t, tt.externalID, user.ExternalID)
			require.Equal(t, tt.externalID, center.externalID)
			expectedStatus := entity.UserStatusAvailable
			if tt.externalID != "" {
				expectedStatus = entity.UserStatusSuspended
			}
			require.Equal(t, expectedStatus, user.UserStatus)
		})
	}
}

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
