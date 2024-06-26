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

package notificationcommon

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/incubator-answer/internal/base/translator"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	"github.com/apache/incubator-answer/internal/service/user_external_login"
	"github.com/apache/incubator-answer/pkg/display"

	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/internal/base/data"
	"github.com/apache/incubator-answer/internal/base/reason"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/internal/service/activity_common"
	"github.com/apache/incubator-answer/internal/service/notice_queue"
	"github.com/apache/incubator-answer/internal/service/object_info"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	"github.com/apache/incubator-answer/pkg/uid"
	"github.com/apache/incubator-answer/plugin"
	"github.com/goccy/go-json"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type NotificationRepo interface {
	AddNotification(ctx context.Context, notification *entity.Notification) (err error)
	GetNotificationPage(ctx context.Context, search *schema.NotificationSearch) ([]*entity.Notification, int64, error)
	ClearUnRead(ctx context.Context, userID string, notificationType int) (err error)
	ClearIDUnRead(ctx context.Context, userID string, id string) (err error)
	GetByUserIdObjectIdTypeId(ctx context.Context, userID, objectID string, notificationType int) (*entity.Notification, bool, error)
	UpdateNotificationContent(ctx context.Context, notification *entity.Notification) (err error)
	GetById(ctx context.Context, id string) (*entity.Notification, bool, error)
}

type NotificationCommon struct {
	data                     *data.Data
	notificationRepo         NotificationRepo
	activityRepo             activity_common.ActivityRepo
	followRepo               activity_common.FollowRepo
	userCommon               *usercommon.UserCommon
	objectInfoService        *object_info.ObjService
	notificationQueueService notice_queue.NotificationQueueService
	userExternalLoginRepo    user_external_login.UserExternalLoginRepo
	siteInfoService          siteinfo_common.SiteInfoCommonService
}

func NewNotificationCommon(
	data *data.Data,
	notificationRepo NotificationRepo,
	userCommon *usercommon.UserCommon,
	activityRepo activity_common.ActivityRepo,
	followRepo activity_common.FollowRepo,
	objectInfoService *object_info.ObjService,
	notificationQueueService notice_queue.NotificationQueueService,
	userExternalLoginRepo user_external_login.UserExternalLoginRepo,
	siteInfoService siteinfo_common.SiteInfoCommonService,
) *NotificationCommon {
	notification := &NotificationCommon{
		data:                     data,
		notificationRepo:         notificationRepo,
		activityRepo:             activityRepo,
		followRepo:               followRepo,
		userCommon:               userCommon,
		objectInfoService:        objectInfoService,
		notificationQueueService: notificationQueueService,
		userExternalLoginRepo:    userExternalLoginRepo,
		siteInfoService:          siteInfoService,
	}
	notificationQueueService.RegisterHandler(notification.AddNotification)
	return notification
}

// AddNotification
// need set
// LoginUserID
// Type  1 inbox 2 achievement
// [inbox] Activity
// [achievement] Rank
// ObjectInfo.Title
// ObjectInfo.ObjectID
// ObjectInfo.ObjectType
func (ns *NotificationCommon) AddNotification(ctx context.Context, msg *schema.NotificationMsg) error {
	if msg.Type == schema.NotificationTypeAchievement && plugin.RankAgentEnabled() {
		return nil
	}
	req := &schema.NotificationContent{
		TriggerUserID:  msg.TriggerUserID,
		ReceiverUserID: msg.ReceiverUserID,
		ObjectInfo: schema.ObjectInfo{
			Title:      msg.Title,
			ObjectID:   uid.DeShortID(msg.ObjectID),
			ObjectType: msg.ObjectType,
		},
		NotificationAction: msg.NotificationAction,
		Type:               msg.Type,
	}
	var questionID string // just for notify all followers
	objInfo, err := ns.objectInfoService.GetInfo(ctx, req.ObjectInfo.ObjectID)
	if err != nil {
		log.Error(err)
	} else {
		req.ObjectInfo.Title = objInfo.Title
		questionID = objInfo.QuestionID
		objectMap := make(map[string]string)
		objectMap["question"] = uid.DeShortID(objInfo.QuestionID)
		objectMap["answer"] = uid.DeShortID(objInfo.AnswerID)
		objectMap["comment"] = objInfo.CommentID
		req.ObjectInfo.ObjectMap = objectMap
	}

	if msg.Type == schema.NotificationTypeAchievement {
		notificationInfo, exist, err := ns.notificationRepo.GetByUserIdObjectIdTypeId(ctx, req.ReceiverUserID, req.ObjectInfo.ObjectID, req.Type)
		if err != nil {
			return fmt.Errorf("get by user id object id type id error: %w", err)
		}
		rank, err := ns.activityRepo.GetUserIDObjectIDActivitySum(ctx, req.ReceiverUserID, req.ObjectInfo.ObjectID)
		if err != nil {
			return fmt.Errorf("get user id object id activity sum error: %w", err)
		}
		req.Rank = rank
		if exist {
			//modify notification
			updateContent := &schema.NotificationContent{}
			err := json.Unmarshal([]byte(notificationInfo.Content), updateContent)
			if err != nil {
				return fmt.Errorf("unmarshal notification content error: %w", err)
			}
			updateContent.Rank = rank
			content, _ := json.Marshal(updateContent)
			notificationInfo.Content = string(content)
			err = ns.notificationRepo.UpdateNotificationContent(ctx, notificationInfo)
			if err != nil {
				return fmt.Errorf("update notification content error: %w", err)
			}
			return nil
		}
	}

	info := &entity.Notification{}
	now := time.Now()
	info.UserID = req.ReceiverUserID
	info.Type = req.Type
	info.IsRead = schema.NotificationNotRead
	info.Status = schema.NotificationStatusNormal
	info.CreatedAt = now
	info.UpdatedAt = now
	info.ObjectID = req.ObjectInfo.ObjectID

	userBasicInfo, exist, err := ns.userCommon.GetUserBasicInfoByID(ctx, req.TriggerUserID)
	if err != nil {
		return fmt.Errorf("get user basic info error: %w", err)
	}
	if !exist {
		return fmt.Errorf("user not exist: %s", req.TriggerUserID)
	}
	req.UserInfo = userBasicInfo
	content, _ := json.Marshal(req)
	_, ok := constant.NotificationMsgTypeMapping[req.NotificationAction]
	if ok {
		info.MsgType = constant.NotificationMsgTypeMapping[req.NotificationAction]
	}
	info.Content = string(content)
	err = ns.notificationRepo.AddNotification(ctx, info)
	if err != nil {
		return fmt.Errorf("add notification error: %w", err)
	}
	err = ns.addRedDot(ctx, info.UserID, info.Type)
	if err != nil {
		log.Error("addRedDot Error", err.Error())
	}

	go ns.SendNotificationToAllFollower(ctx, msg, questionID)

	if msg.Type == schema.NotificationTypeInbox {
		ns.syncNotificationToPlugin(ctx, objInfo, msg)
	}
	return nil
}

func (ns *NotificationCommon) addRedDot(ctx context.Context, userID string, botType int) error {
	key := fmt.Sprintf("answer_RedDot_%d_%s", botType, userID)
	err := ns.data.Cache.SetInt64(ctx, key, 1, 30*24*time.Hour) //Expiration time is one month.
	if err != nil {
		return errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	return nil
}

// SendNotificationToAllFollower send notification to all followers
func (ns *NotificationCommon) SendNotificationToAllFollower(ctx context.Context, msg *schema.NotificationMsg,
	questionID string) {
	if msg.NoNeedPushAllFollow {
		return
	}
	if msg.NotificationAction != constant.NotificationUpdateQuestion &&
		msg.NotificationAction != constant.NotificationAnswerTheQuestion &&
		msg.NotificationAction != constant.NotificationUpdateAnswer &&
		msg.NotificationAction != constant.NotificationAcceptAnswer {
		return
	}
	condObjectID := msg.ObjectID
	if len(questionID) > 0 {
		condObjectID = uid.DeShortID(questionID)
	}
	userIDs, err := ns.followRepo.GetFollowUserIDs(ctx, condObjectID)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("send notification to all followers: %s %d", condObjectID, len(userIDs))
	for _, userID := range userIDs {
		t := &schema.NotificationMsg{}
		_ = copier.Copy(t, msg)
		t.ReceiverUserID = userID
		t.TriggerUserID = msg.TriggerUserID
		t.NoNeedPushAllFollow = true
		ns.notificationQueueService.Send(ctx, t)
	}
}

func (ns *NotificationCommon) syncNotificationToPlugin(ctx context.Context, objInfo *schema.SimpleObjectInfo,
	msg *schema.NotificationMsg) {
	siteInfo, err := ns.siteInfoService.GetSiteGeneral(ctx)
	if err != nil {
		log.Errorf("get site general info failed: %v", err)
		return
	}
	seoInfo, err := ns.siteInfoService.GetSiteSeo(ctx)
	if err != nil {
		log.Errorf("get site seo info failed: %v", err)
		return
	}
	interfaceInfo, err := ns.siteInfoService.GetSiteInterface(ctx)
	if err != nil {
		log.Errorf("get site interface info failed: %v", err)
		return
	}

	objInfo.QuestionID = uid.DeShortID(objInfo.QuestionID)
	objInfo.AnswerID = uid.DeShortID(objInfo.AnswerID)
	pluginNotificationMsg := plugin.NotificationMessage{
		Type:           plugin.NotificationType(msg.NotificationAction),
		ReceiverUserID: msg.ReceiverUserID,
		TriggerUserID:  msg.TriggerUserID,
		QuestionTitle:  objInfo.Title,
	}

	if len(objInfo.QuestionID) > 0 {
		pluginNotificationMsg.QuestionUrl =
			display.QuestionURL(seoInfo.Permalink, siteInfo.SiteUrl, objInfo.QuestionID, objInfo.Title)
	}
	if len(objInfo.AnswerID) > 0 {
		pluginNotificationMsg.AnswerUrl =
			display.AnswerURL(seoInfo.Permalink, siteInfo.SiteUrl, objInfo.QuestionID, objInfo.Title, objInfo.AnswerID)
	}
	if len(objInfo.CommentID) > 0 {
		pluginNotificationMsg.CommentUrl =
			display.CommentURL(seoInfo.Permalink, siteInfo.SiteUrl, objInfo.QuestionID, objInfo.Title, objInfo.AnswerID, objInfo.CommentID)
	}

	if len(msg.TriggerUserID) > 0 {
		triggerUser, exist, err := ns.userCommon.GetUserBasicInfoByID(ctx, msg.TriggerUserID)
		if err != nil {
			log.Errorf("get trigger user basic info failed: %v", err)
			return
		}
		if exist {
			pluginNotificationMsg.TriggerUserID = triggerUser.ID
			pluginNotificationMsg.TriggerUserDisplayName = triggerUser.DisplayName
			pluginNotificationMsg.TriggerUserUrl = display.UserURL(siteInfo.SiteUrl, triggerUser.Username)
		}
	}

	if len(pluginNotificationMsg.ReceiverLang) == 0 && len(msg.ReceiverUserID) > 0 {
		userInfo, _, _ := ns.userCommon.GetUserBasicInfoByID(ctx, msg.ReceiverUserID)
		if userInfo != nil {
			pluginNotificationMsg.ReceiverLang = userInfo.Language
		}
		// If receiver not set language, use site default language.
		if len(pluginNotificationMsg.ReceiverLang) == 0 || pluginNotificationMsg.ReceiverLang == translator.DefaultLangOption {
			pluginNotificationMsg.ReceiverLang = interfaceInfo.Language
		}
	}

	_ = plugin.CallNotification(func(fn plugin.Notification) error {
		userInfo, exist, err := ns.userExternalLoginRepo.GetByUserID(ctx, fn.Info().SlugName, msg.ReceiverUserID)
		if err != nil {
			log.Errorf("get user external login info failed: %v", err)
			return nil
		}
		if exist {
			pluginNotificationMsg.ReceiverExternalID = userInfo.ExternalID
		}
		fn.Notify(pluginNotificationMsg)
		return nil
	})
}
