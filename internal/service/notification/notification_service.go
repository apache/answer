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

	"github.com/apache/incubator-answer/internal/service/report_common"
	"github.com/apache/incubator-answer/internal/service/review"
	usercommon "github.com/apache/incubator-answer/internal/service/user_common"
	"github.com/apache/incubator-answer/pkg/converter"

	"github.com/apache/incubator-answer/internal/base/constant"
	"github.com/apache/incubator-answer/internal/base/data"
	"github.com/apache/incubator-answer/internal/base/handler"
	"github.com/apache/incubator-answer/internal/base/pager"
	"github.com/apache/incubator-answer/internal/base/translator"
	"github.com/apache/incubator-answer/internal/entity"
	"github.com/apache/incubator-answer/internal/schema"
	notficationcommon "github.com/apache/incubator-answer/internal/service/notification_common"
	"github.com/apache/incubator-answer/internal/service/revision_common"
	"github.com/apache/incubator-answer/pkg/uid"
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
}

func NewNotificationService(
	data *data.Data,
	notificationRepo notficationcommon.NotificationRepo,
	notificationCommon *notficationcommon.NotificationCommon,
	revisionService *revision_common.RevisionService,
	userRepo usercommon.UserRepo,
	reportRepo report_common.ReportRepo,
	reviewService *review.ReviewService,
) *NotificationService {
	return &NotificationService{
		data:               data,
		notificationRepo:   notificationRepo,
		notificationCommon: notificationCommon,
		revisionService:    revisionService,
		userRepo:           userRepo,
		reportRepo:         reportRepo,
		reviewService:      reviewService,
	}
}

func (ns *NotificationService) GetRedDot(ctx context.Context, req *schema.GetRedDot) (resp *schema.RedDot, err error) {
	redBot := &schema.RedDot{}
	inboxKey := fmt.Sprintf("answer_RedDot_%d_%s", schema.NotificationTypeInbox, req.UserID)
	achievementKey := fmt.Sprintf("answer_RedDot_%d_%s", schema.NotificationTypeAchievement, req.UserID)
	inboxValue, _, err := ns.data.Cache.GetInt64(ctx, inboxKey)
	if err != nil {
		redBot.Inbox = 0
	} else {
		redBot.Inbox = inboxValue
	}
	achievementValue, _, err := ns.data.Cache.GetInt64(ctx, achievementKey)
	if err != nil {
		redBot.Achievement = 0
	} else {
		redBot.Achievement = achievementValue
	}
	revisionCount := &schema.RevisionSearch{}
	_ = copier.Copy(revisionCount, req)
	if req.CanReviewAnswer || req.CanReviewQuestion || req.CanReviewTag {
		redBot.CanRevision = true
		redBot.Revision = ns.countAllReviewAmount(ctx, req)
	}

	return redBot, nil
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
	botType, ok := schema.NotificationType[req.TypeStr]
	if ok {
		key := fmt.Sprintf("answer_RedDot_%d_%s", botType, req.UserID)
		err := ns.data.Cache.Del(ctx, key)
		if err != nil {
			log.Error("ClearRedDot del cache error", err.Error())
		}
	}
	getRedDotreq := &schema.GetRedDot{}
	_ = copier.Copy(getRedDotreq, req)
	return ns.GetRedDot(ctx, getRedDotreq)
}

func (ns *NotificationService) ClearUnRead(ctx context.Context, userID string, botTypeStr string) error {
	botType, ok := schema.NotificationType[botTypeStr]
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
		log.Error("notificationRepo.GetById error", err.Error())
		return nil
	}
	if !exist {
		return nil
	}
	if notificationInfo.UserID == userID && notificationInfo.IsRead == schema.NotificationNotRead {
		err := ns.notificationRepo.ClearIDUnRead(ctx, userID, id)
		if err != nil {
			return err
		}
	}

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
