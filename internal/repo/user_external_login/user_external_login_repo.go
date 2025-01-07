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
	"encoding/json"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/user_external_login"
	"github.com/segmentfault/pacman/errors"
)

type userExternalLoginRepo struct {
	data *data.Data
}

// NewUserExternalLoginRepo new repository
func NewUserExternalLoginRepo(data *data.Data) user_external_login.UserExternalLoginRepo {
	return &userExternalLoginRepo{
		data: data,
	}
}

// AddUserExternalLogin add external login information
func (ur *userExternalLoginRepo) AddUserExternalLogin(ctx context.Context, user *entity.UserExternalLogin) (err error) {
	_, err = ur.data.DB.Context(ctx).Insert(user)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateInfo update user info
func (ur *userExternalLoginRepo) UpdateInfo(ctx context.Context, userInfo *entity.UserExternalLogin) (err error) {
	_, err = ur.data.DB.Context(ctx).ID(userInfo.ID).Update(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByExternalID get by external ID
func (ur *userExternalLoginRepo) GetByExternalID(ctx context.Context, provider, externalID string) (
	userInfo *entity.UserExternalLogin, exist bool, err error) {
	userInfo = &entity.UserExternalLogin{}
	exist, err = ur.data.DB.Context(ctx).Where("external_id = ?", externalID).Where("provider = ?", provider).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByUserID get by user ID
func (ur *userExternalLoginRepo) GetByUserID(ctx context.Context, provider, userID string) (
	userInfo *entity.UserExternalLogin, exist bool, err error) {
	userInfo = &entity.UserExternalLogin{}
	exist, err = ur.data.DB.Context(ctx).Where("user_id = ?", userID).Where("provider = ?", provider).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserExternalLoginList get by external ID
func (ur *userExternalLoginRepo) GetUserExternalLoginList(ctx context.Context, userID string) (
	resp []*entity.UserExternalLogin, err error) {
	resp = make([]*entity.UserExternalLogin, 0)
	err = ur.data.DB.Context(ctx).Where("user_id = ?", userID).Find(&resp)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// DeleteUserExternalLogin delete external user login info
func (ur *userExternalLoginRepo) DeleteUserExternalLogin(ctx context.Context, userID, externalID string) (err error) {
	cond := &entity.UserExternalLogin{}
	_, err = ur.data.DB.Context(ctx).Where("user_id = ? AND external_id = ?", userID, externalID).Delete(cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// DeleteUserExternalLoginByUserID delete external user login info by user ID
func (ur *userExternalLoginRepo) DeleteUserExternalLoginByUserID(ctx context.Context, userID string) (err error) {
	cond := &entity.UserExternalLogin{}
	_, err = ur.data.DB.Context(ctx).Where("user_id = ?", userID).Delete(cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SetCacheUserExternalLoginInfo cache user info for external login
func (ur *userExternalLoginRepo) SetCacheUserExternalLoginInfo(
	ctx context.Context, key string, info *schema.ExternalLoginUserInfoCache) (err error) {
	cacheData, _ := json.Marshal(info)
	return ur.data.Cache.SetString(ctx, constant.ConnectorUserExternalInfoCacheKey+key,
		string(cacheData), constant.ConnectorUserExternalInfoCacheTime)
}

// GetCacheUserExternalLoginInfo cache user info for external login
func (ur *userExternalLoginRepo) GetCacheUserExternalLoginInfo(
	ctx context.Context, key string) (info *schema.ExternalLoginUserInfoCache, err error) {
	res, exist, err := ur.data.Cache.GetString(ctx, constant.ConnectorUserExternalInfoCacheKey+key)
	if err != nil {
		return info, err
	}
	if !exist {
		return nil, nil
	}
	info = &schema.ExternalLoginUserInfoCache{}
	_ = json.Unmarshal([]byte(res), &info)
	return info, nil
}
