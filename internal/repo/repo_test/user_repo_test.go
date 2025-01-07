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
	"github.com/apache/answer/internal/repo/user"
	"github.com/stretchr/testify/assert"
)

func Test_userRepo_AddUser(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	userInfo := &entity.User{
		Username:    "answer",
		Pass:        "answer",
		EMail:       "answer@example.com",
		MailStatus:  entity.EmailStatusAvailable,
		Status:      entity.UserStatusAvailable,
		DisplayName: "answer",
		IsAdmin:     false,
	}
	err := userRepo.AddUser(context.TODO(), userInfo)
	assert.NoError(t, err)
}

func Test_userRepo_BatchGetByID(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	got, err := userRepo.BatchGetByID(context.TODO(), []string{"1"})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(got))
	assert.Equal(t, "admin", got[0].Username)
}

func Test_userRepo_GetByEmail(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	got, exist, err := userRepo.GetByEmail(context.TODO(), "admin@admin.com")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "admin", got.Username)
}

func Test_userRepo_GetByUserID(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	got, exist, err := userRepo.GetByUserID(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "admin", got.Username)
}

func Test_userRepo_GetByUsername(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	got, exist, err := userRepo.GetByUsername(context.TODO(), "admin")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "admin", got.Username)
}

func Test_userRepo_IncreaseAnswerCount(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.IncreaseAnswerCount(context.TODO(), "1", 1)
	assert.NoError(t, err)

	got, exist, err := userRepo.GetByUserID(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, 1, got.AnswerCount)
}

func Test_userRepo_IncreaseQuestionCount(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.IncreaseQuestionCount(context.TODO(), "1", 1)
	assert.NoError(t, err)

	got, exist, err := userRepo.GetByUserID(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, 1, got.AnswerCount)
}

func Test_userRepo_UpdateEmail(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.UpdateEmail(context.TODO(), "1", "admin@admin.com")
	assert.NoError(t, err)
}

func Test_userRepo_UpdateEmailStatus(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.UpdateEmailStatus(context.TODO(), "1", entity.EmailStatusToBeVerified)
	assert.NoError(t, err)
}

func Test_userRepo_UpdateInfo(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.UpdateInfo(context.TODO(), &entity.User{ID: "1", Bio: "test"})
	assert.NoError(t, err)

	got, exist, err := userRepo.GetByUserID(context.TODO(), "1")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "test", got.Bio)
}

func Test_userRepo_UpdateLastLoginDate(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.UpdateLastLoginDate(context.TODO(), "1")
	assert.NoError(t, err)
}

func Test_userRepo_UpdateNoticeStatus(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.UpdateNoticeStatus(context.TODO(), "1", 1)
	assert.NoError(t, err)
}

func Test_userRepo_UpdatePass(t *testing.T) {
	userRepo := user.NewUserRepo(testDataSource)
	err := userRepo.UpdatePass(context.TODO(), "1", "admin")
	assert.NoError(t, err)
}
