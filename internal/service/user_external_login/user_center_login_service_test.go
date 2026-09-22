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

package user_external_login

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/plugin"
)

func TestUserCenterFirstLoginCachesExternalID(t *testing.T) {
	userRepo := &userCenterLoginTestUserRepo{}
	externalLoginRepo := &userCenterLoginTestExternalLoginRepo{}
	userCommonService := &userCenterLoginTestUserCommonService{}
	service := &UserCenterLoginService{
		userRepo:              userRepo,
		userExternalLoginRepo: externalLoginRepo,
		userCommonService:     userCommonService,
		userActivity:          userCenterLoginTestActivity{},
	}

	const externalID = "central-user-1001"
	resp, err := service.ExternalLogin(context.Background(), userCenterLoginTestUserCenter{},
		&plugin.UserCenterBasicUserInfo{ExternalID: externalID, Username: "central-user"})
	if err != nil {
		t.Fatalf("first user center login failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatal("first user center login did not issue an access token")
	}
	if userCommonService.externalID != externalID {
		t.Fatalf("cached external ID = %q, want %q", userCommonService.externalID, externalID)
	}
	if externalLoginRepo.added == nil || externalLoginRepo.added.ExternalID != externalID {
		t.Fatalf("persisted external ID = %#v, want %q", externalLoginRepo.added, externalID)
	}
}

type userCenterLoginTestUserRepo struct{}

func (*userCenterLoginTestUserRepo) AddUser(_ context.Context, user *entity.User) error {
	user.ID = "answer-user-1"
	return nil
}

func (*userCenterLoginTestUserRepo) GetByUserID(context.Context, string) (*entity.User, bool, error) {
	return nil, false, nil
}

func (*userCenterLoginTestUserRepo) GetByUsername(context.Context, string) (*entity.User, bool, error) {
	return nil, false, nil
}

func (*userCenterLoginTestUserRepo) UpdateLastLoginDate(context.Context, string) error { return nil }

type userCenterLoginTestExternalLoginRepo struct {
	added *entity.UserExternalLogin
}

func (*userCenterLoginTestExternalLoginRepo) GetByExternalID(context.Context, string, string) (*entity.UserExternalLogin, bool, error) {
	return &entity.UserExternalLogin{}, false, nil
}

func (r *userCenterLoginTestExternalLoginRepo) AddUserExternalLogin(_ context.Context, user *entity.UserExternalLogin) error {
	r.added = user
	return nil
}

func (*userCenterLoginTestExternalLoginRepo) GetUserExternalLoginList(context.Context, string) ([]*entity.UserExternalLogin, error) {
	return nil, nil
}

type userCenterLoginTestUserCommonService struct {
	externalID string
}

func (*userCenterLoginTestUserCommonService) MakeUsername(_ context.Context, username string) (string, error) {
	return username, nil
}

func (s *userCenterLoginTestUserCommonService) CacheLoginUserInfo(
	_ context.Context, _ string, _, _ int, externalID string,
) (string, *entity.UserCacheInfo, error) {
	s.externalID = externalID
	return "access-token", &entity.UserCacheInfo{ExternalID: externalID}, nil
}

type userCenterLoginTestActivity struct{}

func (userCenterLoginTestActivity) UserActive(context.Context, string) error { return nil }

type userCenterLoginTestUserCenter struct{}

func (userCenterLoginTestUserCenter) Info() plugin.Info {
	return plugin.Info{SlugName: "test-user-center"}
}

func (userCenterLoginTestUserCenter) Description() plugin.UserCenterDesc {
	return plugin.UserCenterDesc{}
}

func (userCenterLoginTestUserCenter) ControlCenterItems() []plugin.ControlCenter { return nil }

func (userCenterLoginTestUserCenter) LoginCallback(*plugin.GinContext) (*plugin.UserCenterBasicUserInfo, error) {
	return nil, nil
}

func (userCenterLoginTestUserCenter) SignUpCallback(*plugin.GinContext) (*plugin.UserCenterBasicUserInfo, error) {
	return nil, nil
}

func (userCenterLoginTestUserCenter) UserInfo(string) (*plugin.UserCenterBasicUserInfo, error) {
	return nil, nil
}

func (userCenterLoginTestUserCenter) UserStatus(string) plugin.UserStatus {
	return plugin.UserStatusAvailable
}

func (userCenterLoginTestUserCenter) UserList([]string) ([]*plugin.UserCenterBasicUserInfo, error) {
	return nil, nil
}

func (userCenterLoginTestUserCenter) UserSettings(string) (*plugin.SettingInfo, error) {
	return nil, nil
}

func (userCenterLoginTestUserCenter) PersonalBranding(string) []*plugin.PersonalBranding { return nil }

func (userCenterLoginTestUserCenter) AfterLogin(string, string) {}
