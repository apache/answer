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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/auth"
	"github.com/apache/answer/internal/repo/user"
	"github.com/stretchr/testify/assert"
)

func Test_userAdminRepo_GetUserInfo(t *testing.T) {
	userAdminRepo := user.NewUserAdminRepo(testDataSource, auth.NewAuthRepo(testDataSource))
	got, exist, err := userAdminRepo.GetUserInfo(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "1", got.ID)
}

func Test_userAdminRepo_GetUserPage(t *testing.T) {
	userAdminRepo := user.NewUserAdminRepo(testDataSource, auth.NewAuthRepo(testDataSource))
	got, total, err := userAdminRepo.GetUserPage(context.TODO(), 1, 1, &entity.User{Username: "admin"}, "", false)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "1", got[0].ID)
}

func Test_userAdminRepo_UpdateUserStatus(t *testing.T) {
	userAdminRepo := user.NewUserAdminRepo(testDataSource, auth.NewAuthRepo(testDataSource))
	got, exist, err := userAdminRepo.GetUserInfo(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, entity.UserStatusAvailable, got.Status)

	err = userAdminRepo.UpdateUserStatus(context.TODO(), "1", entity.UserStatusSuspended, entity.EmailStatusAvailable,
		"admin@admin.com")
	assert.NoError(t, err)

	got, exist, err = userAdminRepo.GetUserInfo(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, entity.UserStatusSuspended, got.Status)

	err = userAdminRepo.UpdateUserStatus(context.TODO(), "1", entity.UserStatusAvailable, entity.EmailStatusAvailable,
		"admin@admin.com")
	assert.NoError(t, err)

	got, exist, err = userAdminRepo.GetUserInfo(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, entity.UserStatusAvailable, got.Status)
}
