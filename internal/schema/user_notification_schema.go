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

package schema

import (
	"encoding/json"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
)

type NotificationChannelConfig struct {
	Key    constant.NotificationChannelKey `json:"key"`
	Enable bool                            `json:"enable"`
}

type NotificationChannels []*NotificationChannelConfig

func NewNotificationChannelsFormJson(jsonStr string) NotificationChannels {
	var list NotificationChannels
	_ = json.Unmarshal([]byte(jsonStr), &list)
	return list
}

func NewNotificationChannelConfigFormJson(jsonStr string) NotificationChannelConfig {
	var list NotificationChannels
	_ = json.Unmarshal([]byte(jsonStr), &list)
	if len(list) > 0 {
		return *list[0]
	}
	return NotificationChannelConfig{}
}

func (n *NotificationChannels) ToJsonString() string {
	data, _ := json.Marshal(n)
	return string(data)
}

type NotificationConfig struct {
	Inbox                          NotificationChannelConfig `json:"inbox"`
	AllNewQuestion                 NotificationChannelConfig `json:"all_new_question"`
	AllNewQuestionForFollowingTags NotificationChannelConfig `json:"all_new_question_for_following_tags"`
}

func NewNotificationConfig(configs []*entity.UserNotificationConfig) NotificationConfig {
	nc := NotificationConfig{}
	for _, item := range configs {
		switch item.Source {
		case string(constant.InboxSource):
			nc.Inbox = NewNotificationChannelConfigFormJson(item.Channels)
		case string(constant.AllNewQuestionSource):
			nc.AllNewQuestion = NewNotificationChannelConfigFormJson(item.Channels)
		case string(constant.AllNewQuestionForFollowingTagsSource):
			nc.AllNewQuestionForFollowingTags = NewNotificationChannelConfigFormJson(item.Channels)
		}
	}
	return nc
}

func (n *NotificationConfig) Format() {
	if n.Inbox.Key == "" {
		n.Inbox.Key = constant.EmailChannel
		n.Inbox.Enable = false
	}
	if n.AllNewQuestion.Key == "" {
		n.AllNewQuestion.Key = constant.EmailChannel
		n.AllNewQuestion.Enable = false
	}
	if n.AllNewQuestionForFollowingTags.Key == "" {
		n.AllNewQuestionForFollowingTags.Key = constant.EmailChannel
		n.AllNewQuestionForFollowingTags.Enable = false
	}
}

// UpdateUserNotificationConfigReq update user notification config request
type UpdateUserNotificationConfigReq struct {
	NotificationConfig
	UserID string `json:"-"`
}

// GetUserNotificationConfigResp get user notification config response
type GetUserNotificationConfigResp struct {
	NotificationConfig
}
