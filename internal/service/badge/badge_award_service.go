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

package badge

import (
	"context"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/notice_queue"
	"github.com/apache/answer/internal/service/object_info"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type BadgeAwardRepo interface {
	CheckIsAward(ctx context.Context, badgeID string, userID string, awardKey string, singleOrMulti int8) (isAward bool, err error)
	AwardBadgeForUser(ctx context.Context, badgeAward *entity.BadgeAward) (err error)

	CountByUserIdAndBadgeId(ctx context.Context, userID string, badgeID string) (awardCount int64)
	CountByBadgeID(ctx context.Context, badgeID string) (awardCount int64, err error)

	SumUserEarnedGroupByBadgeID(ctx context.Context, userID string) (earnedCounts []*entity.BadgeEarnedCount, err error)

	ListPagedByBadgeId(ctx context.Context, badgeID string, page int, pageSize int) (badgeAwardList []*entity.BadgeAward, total int64, err error)
	ListPagedByBadgeIdAndUserId(ctx context.Context, badgeID string, userID string, page int, pageSize int) (badgeAwards []*entity.BadgeAward, total int64, err error)
	ListNewestEarned(ctx context.Context, userID string, limit int) (badgeAwards []*entity.BadgeAwardRecent, err error)

	GetByUserIdAndBadgeId(ctx context.Context, userID string, badgeID string) (badgeAward *entity.BadgeAward, exists bool, err error)
	GetByUserIdAndBadgeIdAndAwardKey(ctx context.Context, userID string, badgeID string, awardKey string) (badgeAward *entity.BadgeAward, exists bool, err error)

	DeleteUserBadgeAward(ctx context.Context, userID string) (err error)
}

type BadgeAwardService struct {
	badgeAwardRepo           BadgeAwardRepo
	badgeRepo                BadgeRepo
	userCommon               *usercommon.UserCommon
	objectInfoService        *object_info.ObjService
	notificationQueueService notice_queue.NotificationQueueService
}

func NewBadgeAwardService(
	badgeAwardRepo BadgeAwardRepo,
	badgeRepo BadgeRepo,
	userCommon *usercommon.UserCommon,
	objectInfoService *object_info.ObjService,
	notificationQueueService notice_queue.NotificationQueueService,
) *BadgeAwardService {
	return &BadgeAwardService{
		badgeAwardRepo:           badgeAwardRepo,
		badgeRepo:                badgeRepo,
		userCommon:               userCommon,
		objectInfoService:        objectInfoService,
		notificationQueueService: notificationQueueService,
	}
}

// GetBadgeAwardList get badge award list
func (bs *BadgeAwardService) GetBadgeAwardList(
	ctx context.Context,
	req *schema.GetBadgeAwardWithPageReq,
) (resp []*schema.GetBadgeAwardWithPageResp, total int64, err error) {
	var (
		badgeAwardList []*entity.BadgeAward
	)

	req.UserID, err = bs.validateUserByUsername(ctx, req.Username)
	if err != nil {
		badgeAwardList, total, err = bs.badgeAwardRepo.ListPagedByBadgeId(ctx, req.BadgeID, req.Page, req.PageSize)
	} else {
		badgeAwardList, total, err = bs.badgeAwardRepo.ListPagedByBadgeIdAndUserId(ctx, req.BadgeID, req.UserID, req.Page, req.PageSize)
	}

	if err != nil {
		return
	}

	resp = make([]*schema.GetBadgeAwardWithPageResp, len(badgeAwardList))

	for i, badgeAward := range badgeAwardList {
		var (
			objectID, questionID, answerID, commentID, objectType, urlTitle string
		)

		// if exist object info
		objInfo, e := bs.objectInfoService.GetInfo(ctx, badgeAward.AwardKey)
		if e == nil && !objInfo.IsDeleted() {
			objectID = objInfo.ObjectID
			questionID = objInfo.QuestionID
			answerID = objInfo.AnswerID
			commentID = objInfo.CommentID
			objectType = objInfo.ObjectType
			urlTitle = objInfo.Title
		}

		row := &schema.GetBadgeAwardWithPageResp{
			CreatedAt:      badgeAward.CreatedAt.Unix(),
			ObjectID:       objectID,
			QuestionID:     questionID,
			AnswerID:       answerID,
			CommentID:      commentID,
			ObjectType:     objectType,
			UrlTitle:       urlTitle,
			AuthorUserInfo: schema.UserBasicInfo{},
		}

		// get user info
		userInfo, exists, e := bs.userCommon.GetUserBasicInfoByID(ctx, badgeAward.UserID)
		if e != nil {
			log.Errorf("user not found by id: %s, err: %v", badgeAward.UserID, e)
		}
		if exists {
			_ = copier.Copy(&row.AuthorUserInfo, userInfo)
		}

		resp[i] = row
	}

	return
}

// Award award badge
func (bs *BadgeAwardService) Award(ctx context.Context, badgeID string, userID string, awardKey string) (err error) {
	badgeData, exists, err := bs.badgeRepo.GetByID(ctx, badgeID)
	if err != nil {
		return err
	}

	if !exists || badgeData.Status == entity.BadgeStatusInactive {
		return errors.BadRequest(reason.BadgeObjectNotFound)
	}

	alreadyAwarded, err := bs.badgeAwardRepo.CheckIsAward(ctx, badgeID, userID, awardKey, badgeData.Single)
	if err != nil {
		return err
	}
	if alreadyAwarded {
		return nil
	}

	badgeAward := &entity.BadgeAward{
		UserID:         userID,
		BadgeID:        badgeID,
		AwardKey:       awardKey,
		BadgeGroupID:   badgeData.BadgeGroupID,
		IsBadgeDeleted: entity.IsBadgeNotDeleted,
	}
	err = bs.badgeAwardRepo.AwardBadgeForUser(ctx, badgeAward)
	if err != nil {
		return err
	}

	msg := &schema.NotificationMsg{
		TriggerUserID:      badgeAward.UserID,
		ReceiverUserID:     badgeAward.UserID,
		Type:               schema.NotificationTypeAchievement,
		ObjectID:           badgeAward.ID,
		ObjectType:         constant.BadgeAwardObjectType,
		Title:              badgeData.Name,
		ExtraInfo:          map[string]string{"badge_id": badgeData.ID},
		NotificationAction: constant.NotificationEarnedBadge,
	}
	bs.notificationQueueService.Send(ctx, msg)
	return nil
}

// GetUserBadgeAwardList get user badge award list
func (bs *BadgeAwardService) GetUserBadgeAwardList(
	ctx *gin.Context,
	req *schema.GetUserBadgeAwardListReq,
) (
	resp []*schema.GetUserBadgeAwardListResp,
	total int64,
	err error,
) {
	var (
		earnedCounts []*entity.BadgeEarnedCount
	)

	req.UserID, err = bs.validateUserByUsername(ctx, req.Username)
	if err != nil {
		return
	}

	earnedCounts, err = bs.badgeAwardRepo.SumUserEarnedGroupByBadgeID(ctx, req.UserID)
	if err != nil {
		return
	}
	total = int64(len(earnedCounts))
	resp = make([]*schema.GetUserBadgeAwardListResp, total)

	for i, earnedCount := range earnedCounts {
		badge, exists, e := bs.badgeRepo.GetByID(ctx, earnedCount.BadgeID)
		if e != nil {
			err = e
			return
		}
		if !exists {
			continue
		}
		resp[i] = &schema.GetUserBadgeAwardListResp{
			ID:          uid.EnShortID(badge.ID),
			Name:        translator.Tr(handler.GetLangByCtx(ctx), badge.Name),
			Icon:        badge.Icon,
			EarnedCount: earnedCount.EarnedCount,
			Level:       badge.Level,
		}
	}

	return
}

// GetUserRecentBadgeAwardList get user badge award list
func (bs *BadgeAwardService) GetUserRecentBadgeAwardList(ctx *gin.Context, req *schema.GetUserBadgeAwardListReq) (
	resp []*schema.GetUserBadgeAwardListResp, total int64, err error) {
	var (
		earnedCounts []*entity.BadgeAwardRecent
	)

	req.UserID, err = bs.validateUserByUsername(ctx, req.Username)
	if err != nil {
		return
	}

	earnedCounts, err = bs.badgeAwardRepo.ListNewestEarned(ctx, req.UserID, req.Limit)
	if err != nil {
		return
	}

	total = int64(len(earnedCounts))
	resp = make([]*schema.GetUserBadgeAwardListResp, total)

	for i, earnedCount := range earnedCounts {
		badge, exists, e := bs.badgeRepo.GetByID(ctx, earnedCount.BadgeID)
		if e != nil {
			err = e
			return
		}
		if !exists {
			continue
		}
		resp[i] = &schema.GetUserBadgeAwardListResp{
			ID:          uid.EnShortID(badge.ID),
			Name:        translator.Tr(handler.GetLangByCtx(ctx), badge.Name),
			Icon:        badge.Icon,
			EarnedCount: earnedCount.EarnedCount,
			Level:       badge.Level,
		}
	}

	return
}

func (bs *BadgeAwardService) validateUserByUsername(ctx context.Context, userName string) (userID string, err error) {
	var (
		userInfo *schema.UserBasicInfo
		exist    bool
	)
	// validate user exists or not
	if len(userName) > 0 {
		userInfo, exist, err = bs.userCommon.GetUserBasicInfoByUserName(ctx, userName)
		if err != nil {
			return
		}
		if !exist {
			err = errors.BadRequest(reason.UserNotFound)
			return
		}
		userID = userInfo.ID
	}
	if len(userID) == 0 {
		err = errors.BadRequest(reason.UserNotFound)
		return
	}
	return
}
