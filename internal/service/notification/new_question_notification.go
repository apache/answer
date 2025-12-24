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
	"strings"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/display"
	"github.com/apache/answer/pkg/token"
	"github.com/apache/answer/plugin"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/i18n"
	"github.com/segmentfault/pacman/log"
)

type NewQuestionSubscriber struct {
	UserID             string                      `json:"user_id"`
	Channels           schema.NotificationChannels `json:"channels"`
	NotificationSource constant.NotificationSource `json:"notification_source"`
}

func (ns *ExternalNotificationService) handleNewQuestionNotification(ctx context.Context,
	msg *schema.ExternalNotificationMsg) error {
	log.Debugf("try to send new question notification %+v", msg)
	subscribers, err := ns.getNewQuestionSubscribers(ctx, msg)
	if err != nil {
		return err
	}
	log.Debugf("get subscribers %d for question %s", len(subscribers), msg.NewQuestionTemplateRawData.QuestionID)

	for _, subscriber := range subscribers {
		for _, channel := range subscriber.Channels {
			if !channel.Enable {
				continue
			}
			if channel.Key == constant.EmailChannel {
				ns.sendNewQuestionNotificationEmail(ctx, subscriber.UserID, &schema.NewQuestionTemplateRawData{
					QuestionTitle:   msg.NewQuestionTemplateRawData.QuestionTitle,
					QuestionID:      msg.NewQuestionTemplateRawData.QuestionID,
					UnsubscribeCode: token.GenerateToken(),
					Tags:            msg.NewQuestionTemplateRawData.Tags,
					TagIDs:          msg.NewQuestionTemplateRawData.TagIDs,
				})
			}
		}
	}

	ns.syncNewQuestionNotificationToPlugin(ctx, msg)
	return nil
}

func (ns *ExternalNotificationService) getNewQuestionSubscribers(ctx context.Context, msg *schema.ExternalNotificationMsg) (
	subscribers []*NewQuestionSubscriber, err error) {
	subscribersMapping := make(map[string]*NewQuestionSubscriber)

	// 1. get all this new question's tags followers
	tagsFollowerIDs := make([]string, 0)
	followerMapping := make(map[string]bool)
	for _, tagID := range msg.NewQuestionTemplateRawData.TagIDs {
		userIDs, err := ns.followRepo.GetFollowUserIDs(ctx, tagID)
		if err != nil {
			log.Error(err)
			continue
		}
		for _, userID := range userIDs {
			if _, ok := followerMapping[userID]; ok {
				continue
			}
			followerMapping[userID] = true
			tagsFollowerIDs = append(tagsFollowerIDs, userID)
		}
	}
	userNotificationConfigs, err := ns.userNotificationConfigRepo.GetByUsersAndSource(
		ctx, tagsFollowerIDs, constant.AllNewQuestionForFollowingTagsSource)
	if err != nil {
		return nil, err
	}
	for _, userNotificationConfig := range userNotificationConfigs {
		if _, ok := subscribersMapping[userNotificationConfig.UserID]; ok {
			continue
		}
		subscribersMapping[userNotificationConfig.UserID] = &NewQuestionSubscriber{
			UserID:             userNotificationConfig.UserID,
			Channels:           schema.NewNotificationChannelsFormJson(userNotificationConfig.Channels),
			NotificationSource: constant.AllNewQuestionForFollowingTagsSource,
		}
	}
	log.Debugf("get %d subscribers from tags", len(subscribersMapping))

	// 2. get all new question's followers
	notificationConfigs, err := ns.userNotificationConfigRepo.GetBySource(ctx, constant.AllNewQuestionSource)
	if err != nil {
		return nil, err
	}
	for _, notificationConfig := range notificationConfigs {
		if _, ok := subscribersMapping[notificationConfig.UserID]; ok {
			continue
		}
		if ns.checkSendNewQuestionNotificationEmailLimit(ctx, notificationConfig.UserID) {
			continue
		}
		subscribersMapping[notificationConfig.UserID] = &NewQuestionSubscriber{
			UserID:             notificationConfig.UserID,
			Channels:           schema.NewNotificationChannelsFormJson(notificationConfig.Channels),
			NotificationSource: constant.AllNewQuestionSource,
		}
	}

	// 3. remove question owner
	delete(subscribersMapping, msg.NewQuestionTemplateRawData.QuestionAuthorUserID)
	for _, subscriber := range subscribersMapping {
		subscribers = append(subscribers, subscriber)
	}
	log.Debugf("get %d subscribers from all new question config", len(subscribers))
	return subscribers, nil
}

func (ns *ExternalNotificationService) checkSendNewQuestionNotificationEmailLimit(ctx context.Context, userID string) bool {
	key := constant.NewQuestionNotificationLimitCacheKeyPrefix + userID
	old, exist, err := ns.data.Cache.GetInt64(ctx, key)
	if err != nil {
		log.Error(err)
		return false
	}
	if exist && old >= constant.NewQuestionNotificationLimitMax {
		log.Debugf("%s user reach new question notification limit", userID)
		return true
	}
	if !exist {
		err = ns.data.Cache.SetInt64(ctx, key, 1, constant.NewQuestionNotificationLimitCacheTime)
	} else {
		_, err = ns.data.Cache.Increase(ctx, key, 1)
	}
	if err != nil {
		log.Error(err)
	}
	return false
}

func (ns *ExternalNotificationService) sendNewQuestionNotificationEmail(ctx context.Context,
	userID string, rawData *schema.NewQuestionTemplateRawData) {
	if unavailable := ns.checkUserStatusBeforeNotification(ctx, userID); unavailable {
		return
	}
	userInfo, exist, err := ns.userRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Error(err)
		return
	}
	if !exist {
		log.Errorf("user %s not exist", userID)
		return
	}
	// If receiver has set language, use it to send email.
	if len(userInfo.Language) > 0 {
		ctx = context.WithValue(ctx, constant.AcceptLanguageContextKey, i18n.Language(userInfo.Language))
	}
	title, body, err := ns.emailService.NewQuestionTemplate(ctx, rawData)
	if err != nil {
		log.Error(err)
		return
	}

	codeContent := &schema.EmailCodeContent{
		SourceType: schema.UnsubscribeSourceType,
		Email:      userInfo.EMail,
		UserID:     userID,
		NotificationSources: []constant.NotificationSource{
			constant.AllNewQuestionSource,
			constant.AllNewQuestionForFollowingTagsSource,
		},
		SkipValidationLatestCode: true,
	}
	ns.emailService.SendAndSaveCodeWithTime(
		ctx, userInfo.ID, userInfo.EMail, title, body, rawData.UnsubscribeCode, codeContent.ToJSONString(), 1*24*time.Hour)
}

func (ns *ExternalNotificationService) syncNewQuestionNotificationToPlugin(ctx context.Context,
	msg *schema.ExternalNotificationMsg) {
	_ = plugin.CallNotification(func(fn plugin.Notification) error {
		// 1. get all this new question's tags followers
		subscribersMapping := make(map[string]plugin.NotificationType)
		for _, tagID := range msg.NewQuestionTemplateRawData.TagIDs {
			userIDs, err := ns.followRepo.GetFollowUserIDs(ctx, tagID)
			if err != nil {
				log.Error(err)
				continue
			}
			for _, userID := range userIDs {
				subscribersMapping[userID] = plugin.NotificationNewQuestionFollowedTag
			}
		}

		// 2. get all new question's followers
		questionSubscribers := fn.GetNewQuestionSubscribers()
		for _, subscriber := range questionSubscribers {
			subscribersMapping[subscriber] = plugin.NotificationNewQuestion
		}

		// 3. remove question owner
		delete(subscribersMapping, msg.NewQuestionTemplateRawData.QuestionAuthorUserID)

		pluginNotificationMsg := ns.newPluginQuestionNotification(ctx, msg)

		// 4. send notification
		for subscriberUserID, notificationType := range subscribersMapping {
			newMsg := plugin.NotificationMessage{}
			_ = copier.Copy(&newMsg, pluginNotificationMsg)
			newMsg.ReceiverUserID = subscriberUserID
			newMsg.Type = notificationType

			if len(subscriberUserID) > 0 {
				userInfo, _, _ := ns.userRepo.GetByUserID(ctx, subscriberUserID)
				if userInfo != nil && len(userInfo.Language) > 0 && userInfo.Language != translator.DefaultLangOption {
					newMsg.ReceiverLang = userInfo.Language
				}
			}

			externalLogins, err := ns.userExternalLoginRepo.GetUserExternalLoginList(ctx, subscriberUserID)
			if err != nil {
				log.Errorf("get user external login list failed for user %s: %v", subscriberUserID, err)
			} else if len(externalLogins) > 0 {
				newMsg.ReceiverExternalID = externalLogins[0].ExternalID
				if len(externalLogins) > 1 {
					log.Debugf("user %s has %d SSO logins, using most recent: provider=%s",
						subscriberUserID, len(externalLogins), externalLogins[0].Provider)
				}
			}

			fn.Notify(newMsg)
		}
		return nil
	})
}

func (ns *ExternalNotificationService) newPluginQuestionNotification(
	ctx context.Context, msg *schema.ExternalNotificationMsg) (raw *plugin.NotificationMessage) {
	raw = &plugin.NotificationMessage{
		ReceiverUserID: msg.ReceiverUserID,
		ReceiverLang:   msg.ReceiverLang,
		QuestionTitle:  msg.NewQuestionTemplateRawData.QuestionTitle,
		QuestionTags:   strings.Join(msg.NewQuestionTemplateRawData.Tags, ","),
	}
	siteInfo, err := ns.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		return raw
	}
	seoInfo, err := ns.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		return raw
	}
	interfaceInfo, err := ns.siteInfoService.GetSiteInterface(ctx)
	if err != nil {
		return raw
	}
	if len(raw.ReceiverLang) == 0 || raw.ReceiverLang == translator.DefaultLangOption {
		raw.ReceiverLang = interfaceInfo.Language
	}
	raw.QuestionUrl = display.QuestionURL(
		seoInfo.Permalink, siteInfo.SiteUrl,
		msg.NewQuestionTemplateRawData.QuestionID, msg.NewQuestionTemplateRawData.QuestionTitle)
	if len(msg.NewQuestionTemplateRawData.QuestionAuthorUserID) > 0 {
		triggerUser, exist, err := ns.userRepo.GetByUserID(ctx, msg.NewQuestionTemplateRawData.QuestionAuthorUserID)
		if err != nil {
			log.Errorf("get trigger user basic info failed: %v", err)
			return
		}
		if exist {
			raw.TriggerUserID = triggerUser.ID
			raw.TriggerUserDisplayName = triggerUser.DisplayName
			raw.TriggerUserUrl = display.UserURL(siteInfo.SiteUrl, triggerUser.Username)
		}
	}
	return raw
}
