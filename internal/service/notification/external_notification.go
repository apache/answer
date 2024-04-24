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

	"github.com/apache/incubator-answer/internal/base/data"
	"github.com/apache/incubator-answer/internal/base/translator"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/activity_common"
	"github.com/apache/incubator-answer/internal/service/export"
	"github.com/apache/incubator-answer/internal/service/notice_queue"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	"github.com/apache/incubator-answer/internal/service/user_external_login"
	"github.com/apache/incubator-answer/internal/service/user_notification_config"
	"github.com/segmentfault/pacman/log"
)

type ExternalNotificationService struct {
	data                       *data.Data
	userNotificationConfigRepo user_notification_config.UserNotificationConfigRepo
	followRepo                 activity_common.FollowRepo
	emailService               *export.EmailService
	userRepo                   usercommon.UserRepo
	notificationQueueService   notice_queue.ExternalNotificationQueueService
	userExternalLoginRepo      user_external_login.UserExternalLoginRepo
	siteInfoService            siteinfo_common.SiteInfoCommonService
}

func NewExternalNotificationService(
	data *data.Data,
	userNotificationConfigRepo user_notification_config.UserNotificationConfigRepo,
	followRepo activity_common.FollowRepo,
	emailService *export.EmailService,
	userRepo usercommon.UserRepo,
	notificationQueueService notice_queue.ExternalNotificationQueueService,
	userExternalLoginRepo user_external_login.UserExternalLoginRepo,
	siteInfoService siteinfo_common.SiteInfoCommonService,
) *ExternalNotificationService {
	n := &ExternalNotificationService{
		data:                       data,
		userNotificationConfigRepo: userNotificationConfigRepo,
		followRepo:                 followRepo,
		emailService:               emailService,
		userRepo:                   userRepo,
		notificationQueueService:   notificationQueueService,
		userExternalLoginRepo:      userExternalLoginRepo,
		siteInfoService:            siteInfoService,
	}
	notificationQueueService.RegisterHandler(n.Handler)
	return n
}

func (ns *ExternalNotificationService) Handler(ctx context.Context, msg *schema.ExternalNotificationMsg) error {
	log.Debugf("try to send external notification %+v", msg)

	// If receiver not set language, use site default language.
	if len(msg.ReceiverLang) == 0 || msg.ReceiverLang == translator.DefaultLangOption {
		if interfaceInfo, _ := ns.siteInfoService.GetSiteInterface(ctx); interfaceInfo != nil {
			msg.ReceiverLang = interfaceInfo.Language
		}
	}
	if msg.NewQuestionTemplateRawData != nil {
		return ns.handleNewQuestionNotification(ctx, msg)
	}
	if msg.NewCommentTemplateRawData != nil {
		return ns.handleNewCommentNotification(ctx, msg)
	}
	if msg.NewAnswerTemplateRawData != nil {
		return ns.handleNewAnswerNotification(ctx, msg)
	}
	if msg.NewInviteAnswerTemplateRawData != nil {
		return ns.handleInviteAnswerNotification(ctx, msg)
	}
	log.Errorf("unknown notification message: %+v", msg)
	return nil
}
