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
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/uid"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"strings"
)

type BadgeRepo interface {
	GetByID(ctx context.Context, id string) (badge *entity.Badge, exists bool, err error)
	GetByIDs(ctx context.Context, ids []string) (badges []*entity.Badge, err error)

	ListPaged(ctx context.Context, page int, pageSize int) (badges []*entity.Badge, total int64, err error)
	ListActivated(ctx context.Context, page int, pageSize int) (badges []*entity.Badge, total int64, err error)
	ListInactivated(ctx context.Context, page int, pageSize int) (badges []*entity.Badge, total int64, err error)

	UpdateStatus(ctx context.Context, id string, status int8) (err error)
	UpdateAwardCount(ctx context.Context, badgeID string, awardCount int) (err error)
}

type BadgeService struct {
	badgeRepo             BadgeRepo
	badgeGroupRepo        BadgeGroupRepo
	badgeAwardRepo        BadgeAwardRepo
	badgeEventService     *BadgeEventService
	siteInfoCommonService siteinfo_common.SiteInfoCommonService
}

func NewBadgeService(
	badgeRepo BadgeRepo,
	badgeGroupRepo BadgeGroupRepo,
	badgeAwardRepo BadgeAwardRepo,
	badgeEventService *BadgeEventService,
	siteInfoCommonService siteinfo_common.SiteInfoCommonService,
) *BadgeService {
	return &BadgeService{
		badgeRepo:             badgeRepo,
		badgeGroupRepo:        badgeGroupRepo,
		badgeAwardRepo:        badgeAwardRepo,
		badgeEventService:     badgeEventService,
		siteInfoCommonService: siteInfoCommonService,
	}
}

// ListByGroup list all badges group by group
func (b *BadgeService) ListByGroup(ctx context.Context, userID string) (resp []*schema.GetBadgeListResp, err error) {
	var (
		groups       []*entity.BadgeGroup
		badges       []*entity.Badge
		earnedCounts []*entity.BadgeEarnedCount

		groupMap  = make(map[int64]string, 0)
		badgesMap = make(map[int64][]*schema.BadgeListInfo, 0)
	)
	resp = make([]*schema.GetBadgeListResp, 0)

	groups, err = b.badgeGroupRepo.ListGroups(ctx)
	if err != nil {
		return
	}
	badges, _, err = b.badgeRepo.ListActivated(ctx, 0, 0)
	if err != nil {
		return
	}

	if len(userID) > 0 {
		earnedCounts, err = b.badgeAwardRepo.SumUserEarnedGroupByBadgeID(ctx, userID)
		if err != nil {
			return
		}
	}

	for _, group := range groups {
		groupMap[converter.StringToInt64(group.ID)] = translator.Tr(handler.GetLangByCtx(ctx), group.Name)
	}

	for _, badge := range badges {
		// check is earned
		var earned int64 = 0
		if len(earnedCounts) > 0 {
			for _, earnedCount := range earnedCounts {
				if badge.ID == earnedCount.BadgeID && earnedCount.EarnedCount > 0 {
					earned = earnedCount.EarnedCount
					break
				}
			}
		}

		badgesMap[badge.BadgeGroupID] = append(badgesMap[badge.BadgeGroupID], &schema.BadgeListInfo{
			ID:          uid.EnShortID(badge.ID),
			Name:        translator.Tr(handler.GetLangByCtx(ctx), badge.Name),
			Icon:        badge.Icon,
			AwardCount:  badge.AwardCount,
			EarnedCount: earned,
			Level:       badge.Level,
		})
	}

	for _, group := range groups {
		resp = append(resp, &schema.GetBadgeListResp{
			GroupName: translator.Tr(handler.GetLangByCtx(ctx), group.Name),
			Badges:    badgesMap[converter.StringToInt64(group.ID)],
		})
	}

	return
}

// ListPaged list all badges by page
func (b *BadgeService) ListPaged(ctx context.Context, req *schema.GetBadgeListPagedReq) (resp []*schema.GetBadgeListPagedResp, total int64, err error) {
	var (
		groups   []*entity.BadgeGroup
		badges   []*entity.Badge
		badge    *entity.Badge
		exists   bool
		groupMap = make(map[int64]string, 0)
	)

	total = 0

	if len(req.Query) > 0 {
		isID := strings.Index(req.Query, "badge:")
		if isID != 0 {
			badges, err = b.searchByName(ctx, req.Query)
			if err != nil {
				return
			}
			// paged result
			count := len(badges)
			total = int64(count)
			start := (req.Page - 1) * req.PageSize
			end := req.Page * req.PageSize
			if start >= count {
				start = count
				end = count
			}
			if end > count {
				end = count
			}
			badges = badges[start:end]
		} else {
			req.Query = strings.TrimSpace(strings.TrimLeft(req.Query, "badge:"))
			id := uid.DeShortID(req.Query)
			if len(id) == 0 {
				return
			}
			badge, exists, err = b.badgeRepo.GetByID(ctx, id)
			if err != nil || !exists {
				return
			}
			badges = append(badges, badge)
		}
	} else {
		switch req.Status {
		case schema.BadgeStatusActive:
			badges, total, err = b.badgeRepo.ListActivated(ctx, req.Page, req.PageSize)
		case schema.BadgeStatusInactive:
			badges, total, err = b.badgeRepo.ListInactivated(ctx, req.Page, req.PageSize)
		default:
			badges, total, err = b.badgeRepo.ListPaged(ctx, req.Page, req.PageSize)
		}
		if err != nil {
			return
		}
	}

	// find all group and build group map
	groups, err = b.badgeGroupRepo.ListGroups(ctx)
	if err != nil {
		return
	}
	for _, group := range groups {
		groupMap[converter.StringToInt64(group.ID)] = translator.Tr(handler.GetLangByCtx(ctx), group.Name)
	}

	resp = make([]*schema.GetBadgeListPagedResp, len(badges))

	general, siteErr := b.siteInfoCommonService.GetSiteGeneral(ctx)
	var baseURL = ""
	if siteErr != nil {
		baseURL = ""
	} else {
		baseURL = general.SiteUrl
	}

	for i, badge := range badges {
		resp[i] = &schema.GetBadgeListPagedResp{
			ID:          uid.EnShortID(badge.ID),
			Name:        translator.Tr(handler.GetLangByCtx(ctx), badge.Name),
			Description: translator.TrWithData(handler.GetLangByCtx(ctx), badge.Description, &schema.BadgeTplData{ProfileURL: baseURL + "/users/settings/profile"}),
			Icon:        badge.Icon,
			AwardCount:  badge.AwardCount,
			Level:       badge.Level,
			GroupName:   groupMap[badge.BadgeGroupID],
			Status:      schema.BadgeStatusMap[badge.Status],
		}
	}
	return
}

// searchByName
func (b *BadgeService) searchByName(ctx context.Context, name string) (result []*entity.Badge, err error) {
	var badges []*entity.Badge
	name = strings.ToLower(name)
	result = make([]*entity.Badge, 0)

	badges, _, err = b.badgeRepo.ListPaged(ctx, 0, 0)
	for _, badge := range badges {
		tn := strings.ToLower(translator.Tr(handler.GetLangByCtx(ctx), badge.Name))
		if strings.Contains(tn, name) {
			result = append(result, badge)
		}
	}
	return
}

// GetBadgeInfo get badge info
func (b *BadgeService) GetBadgeInfo(ctx *gin.Context, id string, userID string) (info *schema.GetBadgeInfoResp, err error) {
	var (
		badge       *entity.Badge
		earnedTotal int64 = 0
		exists            = false
	)

	badge, exists, err = b.badgeRepo.GetByID(ctx, id)
	if err != nil {
		return
	}

	if !exists || badge.Status == entity.BadgeStatusInactive {
		err = errors.BadRequest(reason.BadgeObjectNotFound)
		return
	}

	if len(userID) > 0 {
		earnedTotal = b.badgeAwardRepo.CountByUserIdAndBadgeId(ctx, userID, badge.ID)
	}

	general, siteErr := b.siteInfoCommonService.GetSiteGeneral(ctx)
	var baseURL = ""
	if siteErr != nil {
		baseURL = ""
	} else {
		baseURL = general.SiteUrl
	}

	info = &schema.GetBadgeInfoResp{
		ID:          uid.EnShortID(badge.ID),
		Name:        translator.Tr(handler.GetLangByCtx(ctx), badge.Name),
		Description: translator.TrWithData(handler.GetLangByCtx(ctx), badge.Description, &schema.BadgeTplData{ProfileURL: baseURL + "/users/settings/profile"}),
		Icon:        badge.Icon,
		AwardCount:  badge.AwardCount,
		EarnedCount: earnedTotal,
		IsSingle:    badge.Single == entity.BadgeSingleAward,
		Level:       badge.Level,
	}
	return
}

// UpdateStatus update badge status
func (b *BadgeService) UpdateStatus(ctx *gin.Context, req *schema.UpdateBadgeStatusReq) (err error) {
	req.ID = uid.DeShortID(req.ID)

	badge, exists, err := b.badgeRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.BadRequest(reason.BadgeObjectNotFound)
	}

	// check duplicate action
	status, ok := schema.BadgeStatusEMap[req.Status]
	if !ok {
		err = errors.BadRequest(reason.StatusInvalid)
		return
	}
	if badge.Status == status {
		return
	}

	err = b.badgeRepo.UpdateStatus(ctx, req.ID, status)
	if err != nil {
		return err
	}

	if status == entity.BadgeStatusActive {
		count, err := b.badgeAwardRepo.CountByBadgeID(ctx, badge.ID)
		if err != nil {
			log.Errorf("count badge award failed: %v", err)
			return nil
		}
		err = b.badgeRepo.UpdateAwardCount(ctx, badge.ID, int(count))
		if err != nil {
			log.Errorf("update badge award count failed: %v", err)
			return nil
		}
	}
	return nil
}
