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

package user_notification_config

import (
	"context"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	usercommon "github.com/apache/answer/internal/service/user_common"
)

type UserNotificationConfigRepo interface {
	Add(ctx context.Context, userIDs []string, source, channels string) (err error)
	Save(ctx context.Context, uc *entity.UserNotificationConfig) (err error)
	GetByUserID(ctx context.Context, userID string) ([]*entity.UserNotificationConfig, error)
	GetBySource(ctx context.Context, source constant.NotificationSource) ([]*entity.UserNotificationConfig, error)
	GetByUserIDAndSource(ctx context.Context, userID string, source constant.NotificationSource) (
		conf *entity.UserNotificationConfig, exist bool, err error)
	GetByUsersAndSource(ctx context.Context, userIDs []string, source constant.NotificationSource) (
		[]*entity.UserNotificationConfig, error)
}

type UserNotificationConfigService struct {
	userRepo                   usercommon.UserRepo
	userNotificationConfigRepo UserNotificationConfigRepo
}

func NewUserNotificationConfigService(
	userRepo usercommon.UserRepo,
	userNotificationConfigRepo UserNotificationConfigRepo,
) *UserNotificationConfigService {
	return &UserNotificationConfigService{
		userRepo:                   userRepo,
		userNotificationConfigRepo: userNotificationConfigRepo,
	}
}

func (us *UserNotificationConfigService) GetUserNotificationConfig(ctx context.Context, userID string) (
	resp *schema.GetUserNotificationConfigResp, err error) {
	notificationConfigs, err := us.userNotificationConfigRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp = &schema.GetUserNotificationConfigResp{}
	resp.NotificationConfig = schema.NewNotificationConfig(notificationConfigs)
	resp.Format()
	return resp, nil
}

func (us *UserNotificationConfigService) UpdateUserNotificationConfig(
	ctx context.Context, req *schema.UpdateUserNotificationConfigReq) (err error) {
	req.NotificationConfig.Format()

	err = us.userNotificationConfigRepo.Save(ctx,
		us.convertToEntity(ctx, req.UserID, constant.InboxSource, req.NotificationConfig.Inbox))
	if err != nil {
		return err
	}
	err = us.userNotificationConfigRepo.Save(ctx,
		us.convertToEntity(ctx, req.UserID, constant.AllNewQuestionSource, req.NotificationConfig.AllNewQuestion))
	if err != nil {
		return err
	}
	err = us.userNotificationConfigRepo.Save(ctx,
		us.convertToEntity(ctx, req.UserID, constant.AllNewQuestionForFollowingTagsSource,
			req.NotificationConfig.AllNewQuestionForFollowingTags))
	if err != nil {
		return err
	}
	return nil
}

// SetDefaultUserNotificationConfig set default user notification config for user register
func (us *UserNotificationConfigService) SetDefaultUserNotificationConfig(ctx context.Context, userIDs []string) (
	err error) {
	return us.userNotificationConfigRepo.Add(ctx, userIDs,
		string(constant.InboxSource), `[{"key":"email","enable":true}]`)
}

func (us *UserNotificationConfigService) convertToEntity(ctx context.Context, userID string,
	source constant.NotificationSource, channel schema.NotificationChannelConfig) (c *entity.UserNotificationConfig) {
	var channels schema.NotificationChannels
	channels = append(channels, &channel)
	c = &entity.UserNotificationConfig{
		UserID:   userID,
		Source:   string(source),
		Channels: channels.ToJsonString(),
	}
	for _, ch := range channels {
		if ch.Enable {
			c.Enabled = true
			break
		}
	}
	return c
}
