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
	"fmt"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/badge"
	notficationcommon "github.com/apache/answer/internal/service/notification_common"
	"github.com/apache/answer/internal/service/report_common"
	"github.com/apache/answer/internal/service/review"
	"github.com/apache/answer/internal/service/revision_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/uid"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/log"
)

// NotificationService user service
type NotificationService struct {
	data               *data.Data
	notificationRepo   notficationcommon.NotificationRepo
	notificationCommon *notficationcommon.NotificationCommon
	revisionService    *revision_common.RevisionService
	reportRepo         report_common.ReportRepo
	reviewService      *review.ReviewService
	userRepo           usercommon.UserRepo
	badgeRepo          badge.BadgeRepo
}

func NewNotificationService(
	data *data.Data,
	notificationRepo notficationcommon.NotificationRepo,
	notificationCommon *notficationcommon.NotificationCommon,
	revisionService *revision_common.RevisionService,
	userRepo usercommon.UserRepo,
	reportRepo report_common.ReportRepo,
	reviewService *review.ReviewService,
	badgeRepo badge.BadgeRepo,
) *NotificationService {
	return &NotificationService{
		data:               data,
		notificationRepo:   notificationRepo,
		notificationCommon: notificationCommon,
		revisionService:    revisionService,
		userRepo:           userRepo,
		reportRepo:         reportRepo,
		reviewService:      reviewService,
		badgeRepo:          badgeRepo,
	}
}

func (ns *NotificationService) GetRedDot(ctx context.Context, req *schema.GetRedDot) (resp *schema.RedDot, err error) {
	inboxKey := fmt.Sprintf(constant.RedDotCacheKey, constant.NotificationTypeInbox, req.UserID)
	achievementKey := fmt.Sprintf(constant.RedDotCacheKey, constant.NotificationTypeAchievement, req.UserID)

	redBot := &schema.RedDot{}
	redBot.Inbox, _, err = ns.data.Cache.GetInt64(ctx, inboxKey)
	redBot.Achievement, _, err = ns.data.Cache.GetInt64(ctx, achievementKey)

	// get review amount
	if req.CanReviewAnswer || req.CanReviewQuestion || req.CanReviewTag {
		redBot.CanRevision = true
		redBot.Revision = ns.countAllReviewAmount(ctx, req)
	}

	// get badge award
	redBot.BadgeAward = ns.getBadgeAward(ctx, req.UserID)
	return redBot, nil
}

func (ns *NotificationService) getBadgeAward(ctx context.Context, userID string) (badgeAward *schema.RedDotBadgeAward) {
	key := fmt.Sprintf(constant.RedDotCacheKey, constant.NotificationTypeBadgeAchievement, userID)
	cacheData, exist, err := ns.data.Cache.GetString(ctx, key)
	if err != nil {
		log.Errorf("get badge award failed: %v", err)
		return nil
	}
	if !exist {
		return nil
	}

	c := schema.NewRedDotBadgeAwardCache()
	c.FromJSON(cacheData)
	award := c.GetBadgeAward()
	if award == nil {
		return nil
	}
	badgeInfo, exists, err := ns.badgeRepo.GetByID(ctx, award.BadgeID)
	if err != nil {
		log.Errorf("get badge info failed: %v", err)
		return nil
	}
	if !exists {
		return nil
	}
	award.Name = translator.Tr(handler.GetLangByCtx(ctx), badgeInfo.Name)
	award.Icon = badgeInfo.Icon
	award.Level = badgeInfo.Level
	return award
}

func (ns *NotificationService) countAllReviewAmount(ctx context.Context, req *schema.GetRedDot) (amount int64) {
	// get queue amount
	if req.IsAdmin {
		reviewCount, err := ns.reviewService.GetReviewPendingCount(ctx)
		if err != nil {
			log.Errorf("get report count failed: %v", err)
		} else {
			amount += reviewCount
		}
	}

	// get flag amount
	if req.IsAdmin {
		reportCount, err := ns.reportRepo.GetReportCount(ctx)
		if err != nil {
			log.Errorf("get report count failed: %v", err)
		} else {
			amount += reportCount
		}
	}

	// get suggestion amount
	countUnreviewedRevision, err := ns.revisionService.GetUnreviewedRevisionCount(ctx, &schema.RevisionSearch{
		CanReviewQuestion: req.CanReviewQuestion,
		CanReviewAnswer:   req.CanReviewAnswer,
		CanReviewTag:      req.CanReviewTag,
		UserID:            req.UserID,
	})
	if err != nil {
		log.Errorf("get unreviewed revision count failed: %v", err)
	} else {
		amount += countUnreviewedRevision
	}
	return amount
}

func (ns *NotificationService) ClearRedDot(ctx context.Context, req *schema.NotificationClearRequest) (*schema.RedDot, error) {
	_ = ns.notificationCommon.DeleteRedDot(ctx, req.UserID, schema.NotificationType[req.NotificationType])
	resp := &schema.GetRedDot{}
	_ = copier.Copy(resp, req)
	return ns.GetRedDot(ctx, resp)
}

func (ns *NotificationService) ClearUnRead(ctx context.Context, userID string, notificationType string) error {
	botType, ok := schema.NotificationType[notificationType]
	if ok {
		err := ns.notificationRepo.ClearUnRead(ctx, userID, botType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ns *NotificationService) ClearIDUnRead(ctx context.Context, userID string, id string) error {
	notificationInfo, exist, err := ns.notificationRepo.GetById(ctx, id)
	if err != nil {
		log.Errorf("get notification failed: %v", err)
		return nil
	}
	if !exist || notificationInfo.UserID != userID {
		return nil
	}
	if notificationInfo.IsRead == schema.NotificationNotRead {
		err := ns.notificationRepo.ClearIDUnRead(ctx, userID, id)
		if err != nil {
			return err
		}
	}

	err = ns.notificationCommon.RemoveBadgeAwardAlertCache(ctx, userID, id)
	if err != nil {
		log.Errorf("remove badge award alert cache failed: %v", err)
	}

	_ = ns.notificationCommon.DecreaseRedDot(ctx, userID, notificationInfo.Type)
	return nil
}

func (ns *NotificationService) GetNotificationPage(ctx context.Context, searchCond *schema.NotificationSearch) (
	pageModel *pager.PageModel, err error) {
	resp := make([]*schema.NotificationContent, 0)
	searchType, ok := schema.NotificationType[searchCond.TypeStr]
	if !ok {
		return pager.NewPageModel(0, resp), nil
	}
	searchInboxType := schema.NotificationInboxTypeAll
	if searchType == schema.NotificationTypeInbox {
		_, ok = schema.NotificationInboxType[searchCond.InboxTypeStr]
		if ok {
			searchInboxType = schema.NotificationInboxType[searchCond.InboxTypeStr]
		}
	}
	searchCond.Type = searchType
	searchCond.InboxType = searchInboxType
	notifications, total, err := ns.notificationRepo.GetNotificationPage(ctx, searchCond)
	if err != nil {
		return nil, err
	}
	resp, err = ns.formatNotificationPage(ctx, notifications)
	if err != nil {
		return nil, err
	}
	return pager.NewPageModel(total, resp), nil
}

func (ns *NotificationService) formatNotificationPage(ctx context.Context, notifications []*entity.Notification) (
	resp []*schema.NotificationContent, err error) {
	lang := handler.GetLangByCtx(ctx)
	enableShortID := handler.GetEnableShortID(ctx)
	userIDs := make([]string, 0)
	userMapping := make(map[string]bool)
	for _, notificationInfo := range notifications {
		item := &schema.NotificationContent{}
		if err := json.Unmarshal([]byte(notificationInfo.Content), item); err != nil {
			log.Error("NotificationContent Unmarshal Error", err.Error())
			continue
		}
		// If notification is downvote, the user info is not needed.
		if item.NotificationAction == constant.NotificationDownVotedTheQuestion ||
			item.NotificationAction == constant.NotificationDownVotedTheAnswer {
			item.UserInfo = nil
		}
		// If notification is badge, the user info is not needed and the title need to be translated.
		if item.ObjectInfo.ObjectType == constant.BadgeAwardObjectType {
			badgeName := translator.Tr(lang, item.ObjectInfo.Title)
			item.ObjectInfo.Title = translator.TrWithData(lang, constant.NotificationEarnedBadge, struct {
				BadgeName string
			}{BadgeName: badgeName})
			item.UserInfo = nil
		}

		item.ID = notificationInfo.ID
		item.NotificationAction = translator.Tr(lang, item.NotificationAction)
		item.UpdateTime = notificationInfo.UpdatedAt.Unix()
		item.IsRead = notificationInfo.IsRead == schema.NotificationRead

		if enableShortID {
			if answerID, ok := item.ObjectInfo.ObjectMap["answer"]; ok {
				if item.ObjectInfo.ObjectID == answerID {
					item.ObjectInfo.ObjectID = uid.EnShortID(item.ObjectInfo.ObjectMap["answer"])
				}
				item.ObjectInfo.ObjectMap["answer"] = uid.EnShortID(item.ObjectInfo.ObjectMap["answer"])
			}
			if questionID, ok := item.ObjectInfo.ObjectMap["question"]; ok {
				if item.ObjectInfo.ObjectID == questionID {
					item.ObjectInfo.ObjectID = uid.EnShortID(item.ObjectInfo.ObjectMap["question"])
				}
				item.ObjectInfo.ObjectMap["question"] = uid.EnShortID(item.ObjectInfo.ObjectMap["question"])
			}
		}

		if item.UserInfo != nil && !userMapping[item.UserInfo.ID] {
			userIDs = append(userIDs, item.UserInfo.ID)
			userMapping[item.UserInfo.ID] = true
		}
		resp = append(resp, item)
	}

	if len(userIDs) == 0 {
		return resp, nil
	}

	users, err := ns.userRepo.BatchGetByID(ctx, userIDs)
	if err != nil {
		log.Error(err)
		return resp, nil
	}
	userIDMapping := make(map[string]*entity.User, len(users))
	for _, user := range users {
		userIDMapping[user.ID] = user
	}
	for _, item := range resp {
		if item.UserInfo == nil {
			continue
		}
		userInfo, ok := userIDMapping[item.UserInfo.ID]
		if !ok {
			continue
		}
		if userInfo.Status == entity.UserStatusDeleted {
			item.UserInfo = &schema.UserBasicInfo{
				DisplayName: "user" + converter.DeleteUserDisplay(userInfo.ID),
				Status:      constant.UserDeleted,
			}
		}
	}
	return resp, nil
}
