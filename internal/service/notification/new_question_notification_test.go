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

package notification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/constant"
	basedata "github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/export"
	"github.com/apache/answer/internal/service/mock"
	"go.uber.org/mock/gomock"
)

func TestNewQuestionNotificationEmailSendInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  time.Duration
	}{
		{
			name: "unset",
			want: 0,
		},
		{
			name:  "empty",
			value: "",
			set:   true,
			want:  0,
		},
		{
			name:  "positive integer",
			value: "5",
			set:   true,
			want:  5 * time.Second,
		},
		{
			name:  "positive integer with whitespace",
			value: " 5 ",
			set:   true,
			want:  5 * time.Second,
		},
		{
			name:  "invalid",
			value: "not-a-number",
			set:   true,
			want:  0,
		},
		{
			name:  "negative",
			value: "-1",
			set:   true,
			want:  0,
		},
		{
			name:  "duration overflow",
			value: "9223372037",
			set:   true,
			want:  0,
		},
		{
			name:  "parse int overflow",
			value: "9223372036854775808",
			set:   true,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setNewQuestionNotificationEmailSendIntervalEnv(t, tt.value, tt.set)

			got := newQuestionNotificationEmailSendInterval()
			if got != tt.want {
				t.Fatalf("newQuestionNotificationEmailSendInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendNewQuestionNotificationEmailsWithInterval(t *testing.T) {
	rawData := &schema.NewQuestionTemplateRawData{
		QuestionTitle: "question",
		QuestionID:    "1",
		Tags:          []string{"go"},
		TagIDs:        []string{"tag-1"},
	}
	interval := 3 * time.Second

	tests := []struct {
		name        string
		interval    time.Duration
		subscribers []*NewQuestionSubscriber
		wantSends   []string
		wantSleeps  []time.Duration
		wantEvents  []string
	}{
		{
			name:     "interval 0",
			interval: 0,
			subscribers: []*NewQuestionSubscriber{
				newQuestionSubscriber("user-1", newQuestionEmailChannel(true)),
				newQuestionSubscriber("user-2", newQuestionEmailChannel(true)),
			},
			wantSends: []string{"user-1", "user-2"},
			wantEvents: []string{
				"send:user-1",
				"send:user-2",
			},
		},
		{
			name:     "0 enabled email attempts",
			interval: interval,
			subscribers: []*NewQuestionSubscriber{
				newQuestionSubscriber("user-1", newQuestionEmailChannel(false)),
				newQuestionSubscriber("user-2", newQuestionNonEmailChannel(true)),
			},
		},
		{
			name:     "1 enabled email attempt",
			interval: interval,
			subscribers: []*NewQuestionSubscriber{
				newQuestionSubscriber("user-1", newQuestionEmailChannel(true)),
			},
			wantSends: []string{"user-1"},
			wantEvents: []string{
				"send:user-1",
			},
		},
		{
			name:     "N enabled email attempts",
			interval: interval,
			subscribers: []*NewQuestionSubscriber{
				newQuestionSubscriber("user-1", newQuestionEmailChannel(true)),
				newQuestionSubscriber("user-2", newQuestionEmailChannel(true)),
				newQuestionSubscriber("user-3", newQuestionEmailChannel(true)),
			},
			wantSends:  []string{"user-1", "user-2", "user-3"},
			wantSleeps: []time.Duration{interval, interval},
			wantEvents: []string{
				"send:user-1",
				"sleep:3s",
				"send:user-2",
				"sleep:3s",
				"send:user-3",
			},
		},
		{
			name:     "disabled email channel does not add delay",
			interval: interval,
			subscribers: []*NewQuestionSubscriber{
				newQuestionSubscriber("user-1", newQuestionEmailChannel(true)),
				newQuestionSubscriber("user-2", newQuestionEmailChannel(false)),
				newQuestionSubscriber("user-3", newQuestionEmailChannel(true)),
			},
			wantSends:  []string{"user-1", "user-3"},
			wantSleeps: []time.Duration{interval},
			wantEvents: []string{
				"send:user-1",
				"sleep:3s",
				"send:user-3",
			},
		},
		{
			name:     "non-email channel does not add delay",
			interval: interval,
			subscribers: []*NewQuestionSubscriber{
				newQuestionSubscriber("user-1", newQuestionEmailChannel(true)),
				newQuestionSubscriber("user-2", newQuestionNonEmailChannel(true)),
				newQuestionSubscriber("user-3", newQuestionEmailChannel(true)),
			},
			wantSends:  []string{"user-1", "user-3"},
			wantSleeps: []time.Duration{interval},
			wantEvents: []string{
				"send:user-1",
				"sleep:3s",
				"send:user-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotEvents []string
			var gotSleeps []time.Duration
			sleep := func(duration time.Duration) {
				gotSleeps = append(gotSleeps, duration)
				gotEvents = append(gotEvents, "sleep:"+duration.String())
			}

			var gotSends []string
			var gotCodes []string
			send := func(_ context.Context, userID string, rawData *schema.NewQuestionTemplateRawData) {
				gotSends = append(gotSends, userID)
				gotEvents = append(gotEvents, "send:"+userID)
				if rawData.UnsubscribeCode == "" {
					t.Fatalf("expected unsubscribe code for %s", userID)
				}
				gotCodes = append(gotCodes, rawData.UnsubscribeCode)
			}

			sendNewQuestionNotificationEmailsWithInterval(
				context.Background(), tt.subscribers, rawData, tt.interval, sleep, send)

			if !reflect.DeepEqual(gotSends, tt.wantSends) {
				t.Fatalf("send calls = %v, want %v", gotSends, tt.wantSends)
			}
			if !reflect.DeepEqual(gotSleeps, tt.wantSleeps) {
				t.Fatalf("sleep calls = %v, want %v", gotSleeps, tt.wantSleeps)
			}
			if !reflect.DeepEqual(gotEvents, tt.wantEvents) {
				t.Fatalf("events = %v, want %v", gotEvents, tt.wantEvents)
			}
			assertUniqueNewQuestionUnsubscribeCodes(t, gotCodes)
		})
	}
}

func TestHandleNewQuestionNotificationSendsEmailsThroughFanOut(t *testing.T) {
	setNewQuestionNotificationEmailSendIntervalEnv(t, "0", true)

	cache, cleanup, err := basedata.NewCache(&basedata.CacheConf{})
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	t.Cleanup(cleanup)

	ctrl := gomock.NewController(t)
	siteInfoService := mock.NewMockSiteInfoCommonService(ctrl)
	siteInfoService.EXPECT().GetSiteGeneral(gomock.Any()).Return(&schema.SiteGeneralResp{
		Name:         "Answer",
		SiteUrl:      "https://answer.test",
		ContactEmail: "support@answer.test",
	}, nil).AnyTimes()
	siteInfoService.EXPECT().GetSiteSeo(gomock.Any()).Return(&schema.SiteSeoResp{
		Permalink: constant.PermalinkQuestionIDAndTitle,
	}, nil).AnyTimes()

	emailRepo := &newQuestionNotificationTestEmailRepo{
		codesByUserID: make(map[string][]string),
	}
	notificationConfigRepo := &newQuestionNotificationTestUserNotificationConfigRepo{
		followedTagConfigs: map[string]*entity.UserNotificationConfig{
			"tag-user": newQuestionNotificationConfig(
				"tag-user", constant.AllNewQuestionForFollowingTagsSource, true),
			"dup-user": newQuestionNotificationConfig(
				"dup-user", constant.AllNewQuestionForFollowingTagsSource, true),
			"author": newQuestionNotificationConfig(
				"author", constant.AllNewQuestionForFollowingTagsSource, true),
		},
		allQuestionConfigs: []*entity.UserNotificationConfig{
			newQuestionNotificationConfig("all-user", constant.AllNewQuestionSource, true),
			newQuestionNotificationConfig("dup-user", constant.AllNewQuestionSource, true),
			newQuestionNotificationConfig("author", constant.AllNewQuestionSource, true),
		},
	}
	service := &ExternalNotificationService{
		data: &basedata.Data{
			Cache: cache,
		},
		userNotificationConfigRepo: notificationConfigRepo,
		followRepo: &newQuestionNotificationTestFollowRepo{
			followersByObjectID: map[string][]string{
				"tag-1": {"tag-user", "dup-user", "author"},
			},
		},
		emailService: export.NewEmailService(
			config.NewConfigService(newQuestionNotificationTestConfigRepo{}),
			emailRepo,
			siteInfoService,
		),
		userRepo: &newQuestionNotificationTestUserRepo{
			users: map[string]*entity.User{
				"tag-user": newQuestionNotificationTestUser("tag-user"),
				"dup-user": newQuestionNotificationTestUser("dup-user"),
				"all-user": newQuestionNotificationTestUser("all-user"),
				"author":   newQuestionNotificationTestUser("author"),
			},
		},
		siteInfoService: siteInfoService,
	}

	err = service.handleNewQuestionNotification(context.Background(), &schema.ExternalNotificationMsg{
		NewQuestionTemplateRawData: &schema.NewQuestionTemplateRawData{
			QuestionTitle:        "New question",
			QuestionID:           "1",
			QuestionAuthorUserID: "author",
			Tags:                 []string{"go"},
			TagIDs:               []string{"tag-1"},
		},
	})
	if err != nil {
		t.Fatalf("handleNewQuestionNotification() error = %v", err)
	}

	wantUsers := []string{"all-user", "dup-user", "tag-user"}
	assertStringSet(t, emailRepo.userIDs(), wantUsers)
	for _, userID := range wantUsers {
		codes := emailRepo.codesByUserID[userID]
		if len(codes) != 1 {
			t.Fatalf("saved codes for %s = %v, want exactly one code", userID, codes)
		}
		if codes[0] == "" {
			t.Fatalf("saved empty code for %s", userID)
		}
	}
	if codes := emailRepo.codesByUserID["author"]; len(codes) > 0 {
		t.Fatalf("question author received notification codes: %v", codes)
	}
}

func assertUniqueNewQuestionUnsubscribeCodes(t *testing.T, codes []string) {
	t.Helper()

	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Fatalf("duplicate unsubscribe code %q", code)
		}
		seen[code] = true
	}
}

func setNewQuestionNotificationEmailSendIntervalEnv(t *testing.T, value string, set bool) {
	t.Helper()

	oldValue, oldSet := os.LookupEnv(newQuestionNotificationEmailSendIntervalEnv)
	if set {
		if err := os.Setenv(newQuestionNotificationEmailSendIntervalEnv, value); err != nil {
			t.Fatalf("set env: %v", err)
		}
	} else {
		if err := os.Unsetenv(newQuestionNotificationEmailSendIntervalEnv); err != nil {
			t.Fatalf("unset env: %v", err)
		}
	}
	t.Cleanup(func() {
		if oldSet {
			_ = os.Setenv(newQuestionNotificationEmailSendIntervalEnv, oldValue)
		} else {
			_ = os.Unsetenv(newQuestionNotificationEmailSendIntervalEnv)
		}
	})
}

func newQuestionSubscriber(userID string, channels ...*schema.NotificationChannelConfig) *NewQuestionSubscriber {
	return &NewQuestionSubscriber{
		UserID:   userID,
		Channels: channels,
	}
}

func newQuestionEmailChannel(enable bool) *schema.NotificationChannelConfig {
	return &schema.NotificationChannelConfig{
		Key:    constant.EmailChannel,
		Enable: enable,
	}
}

func newQuestionNonEmailChannel(enable bool) *schema.NotificationChannelConfig {
	return &schema.NotificationChannelConfig{
		Key:    constant.NotificationChannelKey("inbox"),
		Enable: enable,
	}
}

func newQuestionNotificationConfig(
	userID string, source constant.NotificationSource, emailEnabled bool) *entity.UserNotificationConfig {
	channels := schema.NotificationChannels{
		newQuestionEmailChannel(emailEnabled),
	}
	return &entity.UserNotificationConfig{
		UserID:   userID,
		Source:   string(source),
		Channels: channels.ToJsonString(),
		Enabled:  emailEnabled,
	}
}

func newQuestionNotificationTestUser(userID string) *entity.User {
	return &entity.User{
		ID:          userID,
		Username:    userID,
		DisplayName: userID,
		EMail:       userID + "@example.com",
		Status:      entity.UserStatusAvailable,
		MailStatus:  entity.EmailStatusAvailable,
	}
}

func assertStringSet(t *testing.T, got, want []string) {
	t.Helper()

	gotSet := make(map[string]bool)
	for _, value := range got {
		gotSet[value] = true
	}
	wantSet := make(map[string]bool)
	for _, value := range want {
		wantSet[value] = true
	}
	if !reflect.DeepEqual(gotSet, wantSet) {
		t.Fatalf("values = %v, want %v", got, want)
	}
}

type newQuestionNotificationTestFollowRepo struct {
	followersByObjectID map[string][]string
}

func (r *newQuestionNotificationTestFollowRepo) GetFollowIDs(
	context.Context, string, string) ([]string, error) {
	return nil, nil
}

func (r *newQuestionNotificationTestFollowRepo) GetFollowAmount(context.Context, string) (int, error) {
	return 0, nil
}

func (r *newQuestionNotificationTestFollowRepo) GetFollowUserIDs(
	_ context.Context, objectID string) ([]string, error) {
	return r.followersByObjectID[objectID], nil
}

func (r *newQuestionNotificationTestFollowRepo) IsFollowed(context.Context, string, string) (bool, error) {
	return false, nil
}

func (r *newQuestionNotificationTestFollowRepo) MigrateFollowers(
	context.Context, string, string, string) error {
	return nil
}

type newQuestionNotificationTestUserNotificationConfigRepo struct {
	followedTagConfigs map[string]*entity.UserNotificationConfig
	allQuestionConfigs []*entity.UserNotificationConfig
}

func (r *newQuestionNotificationTestUserNotificationConfigRepo) Add(
	context.Context, []string, string, string) error {
	return nil
}

func (r *newQuestionNotificationTestUserNotificationConfigRepo) Save(
	context.Context, *entity.UserNotificationConfig) error {
	return nil
}

func (r *newQuestionNotificationTestUserNotificationConfigRepo) GetByUserID(
	context.Context, string) ([]*entity.UserNotificationConfig, error) {
	return nil, nil
}

func (r *newQuestionNotificationTestUserNotificationConfigRepo) GetBySource(
	_ context.Context, source constant.NotificationSource) ([]*entity.UserNotificationConfig, error) {
	if source == constant.AllNewQuestionSource {
		return r.allQuestionConfigs, nil
	}
	return nil, nil
}

func (r *newQuestionNotificationTestUserNotificationConfigRepo) GetByUserIDAndSource(
	context.Context, string, constant.NotificationSource) (*entity.UserNotificationConfig, bool, error) {
	return nil, false, nil
}

func (r *newQuestionNotificationTestUserNotificationConfigRepo) GetByUsersAndSource(
	_ context.Context, userIDs []string, source constant.NotificationSource) (
	[]*entity.UserNotificationConfig, error) {
	if source != constant.AllNewQuestionForFollowingTagsSource {
		return nil, nil
	}
	configs := make([]*entity.UserNotificationConfig, 0, len(userIDs))
	for _, userID := range userIDs {
		if config, ok := r.followedTagConfigs[userID]; ok {
			configs = append(configs, config)
		}
	}
	return configs, nil
}

type newQuestionNotificationTestUserRepo struct {
	users map[string]*entity.User
}

func (r *newQuestionNotificationTestUserRepo) AddUser(context.Context, *entity.User) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) IncreaseAnswerCount(context.Context, string, int) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) IncreaseQuestionCount(context.Context, string, int) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateQuestionCount(context.Context, string, int64) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateAnswerCount(context.Context, string, int) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateLastLoginDate(context.Context, string) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateEmailStatus(context.Context, string, int) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateNoticeStatus(context.Context, string, int) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateEmail(context.Context, string, string) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateUserInterface(
	context.Context, string, string, string) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdatePass(context.Context, string, string) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateInfo(context.Context, *entity.User) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) UpdateUserProfile(context.Context, *entity.User) error {
	return nil
}

func (r *newQuestionNotificationTestUserRepo) GetByUserID(
	_ context.Context, userID string) (*entity.User, bool, error) {
	user, ok := r.users[userID]
	return user, ok, nil
}

func (r *newQuestionNotificationTestUserRepo) BatchGetByID(
	context.Context, []string) ([]*entity.User, error) {
	return nil, nil
}

func (r *newQuestionNotificationTestUserRepo) GetByUsername(
	context.Context, string) (*entity.User, bool, error) {
	return nil, false, nil
}

func (r *newQuestionNotificationTestUserRepo) GetByUsernames(
	context.Context, []string) ([]*entity.User, error) {
	return nil, nil
}

func (r *newQuestionNotificationTestUserRepo) GetByEmail(
	context.Context, string) (*entity.User, bool, error) {
	return nil, false, nil
}

func (r *newQuestionNotificationTestUserRepo) GetUserCount(context.Context) (int64, error) {
	return 0, nil
}

func (r *newQuestionNotificationTestUserRepo) SearchUserListByName(
	context.Context, string, int, bool) ([]*entity.User, error) {
	return nil, nil
}

func (r *newQuestionNotificationTestUserRepo) IsAvatarFileUsed(context.Context, string) (bool, error) {
	return false, nil
}

type newQuestionNotificationTestConfigRepo struct{}

func (newQuestionNotificationTestConfigRepo) GetConfigByID(
	context.Context, int) (*entity.Config, error) {
	return nil, nil
}

func (newQuestionNotificationTestConfigRepo) GetConfigByKey(
	context.Context, string) (*entity.Config, error) {
	config := export.EmailConfig{
		FromEmail: "noreply@answer.test",
		FromName:  "Answer",
	}
	value, _ := json.Marshal(config)
	return &entity.Config{
		Value: string(value),
	}, nil
}

func (newQuestionNotificationTestConfigRepo) GetConfigByKeyFromDB(
	context.Context, string) (*entity.Config, error) {
	return nil, nil
}

func (newQuestionNotificationTestConfigRepo) UpdateConfig(context.Context, string, string) error {
	return nil
}

type newQuestionNotificationTestEmailRepo struct {
	codesByUserID map[string][]string
}

func (r *newQuestionNotificationTestEmailRepo) SetCode(
	_ context.Context, userID, code, _ string, _ time.Duration) error {
	r.codesByUserID[userID] = append(r.codesByUserID[userID], code)
	return nil
}

func (r *newQuestionNotificationTestEmailRepo) VerifyCode(context.Context, string) (string, error) {
	return "", nil
}

func (r *newQuestionNotificationTestEmailRepo) userIDs() []string {
	userIDs := make([]string, 0, len(r.codesByUserID))
	for userID := range r.codesByUserID {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
