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

package rank

import (
	"context"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity_type"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/object_info"
	"github.com/apache/answer/internal/service/permission"
	"github.com/apache/answer/internal/service/role"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/uid"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

const (
	PermissionPrefix = "rank."
)

type UserRankRepo interface {
	GetMaxDailyRank(ctx context.Context) (maxDailyRank int, err error)
	CheckReachLimit(ctx context.Context, session *xorm.Session, userID string, maxDailyRank int) (reach bool, err error)
	ChangeUserRank(ctx context.Context, session *xorm.Session,
		userID string, userCurrentScore, deltaRank int) (err error)
	TriggerUserRank(ctx context.Context, session *xorm.Session, userId string, rank int, activityType int) (isReachStandard bool, err error)
	UserRankPage(ctx context.Context, userId string, page, pageSize int) (rankPage []*entity.Activity, total int64, err error)
}

// RankService rank service
type RankService struct {
	userCommon        *usercommon.UserCommon
	configService     *config.ConfigService
	userRankRepo      UserRankRepo
	objectInfoService *object_info.ObjService
	roleService       *role.UserRoleRelService
	rolePowerService  *role.RolePowerRelService
}

// NewRankService new rank service
func NewRankService(
	userCommon *usercommon.UserCommon,
	userRankRepo UserRankRepo,
	objectInfoService *object_info.ObjService,
	roleService *role.UserRoleRelService,
	rolePowerService *role.RolePowerRelService,
	configService *config.ConfigService) *RankService {
	return &RankService{
		userCommon:        userCommon,
		configService:     configService,
		userRankRepo:      userRankRepo,
		objectInfoService: objectInfoService,
		roleService:       roleService,
		rolePowerService:  rolePowerService,
	}
}

// CheckOperationPermission verify that the user has permission
func (rs *RankService) CheckOperationPermission(ctx context.Context, userID string, action string, objectID string) (
	can bool, err error) {
	if len(userID) == 0 {
		return false, nil
	}

	// get the rank of the current user
	userInfo, exist, err := rs.userCommon.GetUserBasicInfoByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	powerMapping := rs.getUserPowerMapping(ctx, userID)
	if powerMapping[action] {
		return true, nil
	}

	if len(objectID) > 0 {
		objectInfo, err := rs.objectInfoService.GetInfo(ctx, objectID)
		if err != nil {
			return can, err
		}
		// if the user is this object creator, the user can operate this object.
		if objectInfo != nil &&
			objectInfo.ObjectCreatorUserID == userID {
			return true, nil
		}
	}

	can, _ = rs.checkUserRank(ctx, userInfo.ID, userInfo.Rank, PermissionPrefix+action)
	return can, nil
}

// CheckOperationPermissionsForRanks verify that the user has permission
func (rs *RankService) CheckOperationPermissionsForRanks(ctx context.Context, userID string, actions []string) (
	can []bool, requireRanks []int, err error) {
	can = make([]bool, len(actions))
	requireRanks = make([]int, len(actions))
	if len(userID) == 0 {
		return can, requireRanks, nil
	}

	// get the rank of the current user
	userInfo, exist, err := rs.userCommon.GetUserBasicInfoByID(ctx, userID)
	if err != nil {
		return can, requireRanks, err
	}
	if !exist {
		return can, requireRanks, nil
	}

	powerMapping := rs.getUserPowerMapping(ctx, userID)
	for idx, action := range actions {
		if powerMapping[action] {
			can[idx] = true
			continue
		}
		meetRank, requireRank := rs.checkUserRank(ctx, userInfo.ID, userInfo.Rank, PermissionPrefix+action)
		can[idx] = meetRank
		requireRanks[idx] = requireRank
	}
	return can, requireRanks, nil
}

// CheckOperationPermissions verify that the user has permission
func (rs *RankService) CheckOperationPermissions(ctx context.Context, userID string, actions []string) (
	can []bool, err error) {
	can, _, err = rs.CheckOperationPermissionsForRanks(ctx, userID, actions)
	return can, err
}

// CheckOperationObjectOwner check operation object owner
func (rs *RankService) CheckOperationObjectOwner(ctx context.Context, userID, objectID string) bool {
	objectID = uid.DeShortID(objectID)
	objectInfo, err := rs.objectInfoService.GetInfo(ctx, objectID)
	if err != nil {
		log.Error(err)
		return false
	}
	// if the user is this object creator, the user can operate this object.
	if objectInfo != nil &&
		objectInfo.ObjectCreatorUserID == userID {
		return true
	}
	return false
}

// CheckVotePermission verify that the user has vote permission
func (rs *RankService) CheckVotePermission(ctx context.Context, userID, objectID string, voteUp bool) (
	can bool, needRank int, err error) {
	if len(userID) == 0 || len(objectID) == 0 {
		return false, 0, nil
	}

	// get the rank of the current user
	userInfo, exist, err := rs.userCommon.GetUserBasicInfoByID(ctx, userID)
	if err != nil {
		return can, 0, err
	}
	if !exist {
		return can, 0, nil
	}
	objectInfo, err := rs.objectInfoService.GetInfo(ctx, objectID)
	if err != nil {
		return can, 0, err
	}
	action := ""
	switch objectInfo.ObjectType {
	case constant.QuestionObjectType:
		if voteUp {
			action = permission.QuestionVoteUp
		} else {
			action = permission.QuestionVoteDown
		}
	case constant.AnswerObjectType:
		if voteUp {
			action = permission.AnswerVoteUp
		} else {
			action = permission.AnswerVoteDown
		}
	case constant.CommentObjectType:
		if voteUp {
			action = permission.CommentVoteUp
		} else {
			action = permission.CommentVoteDown
		}
	}
	powerMapping := rs.getUserPowerMapping(ctx, userID)
	if powerMapping[action] {
		return true, 0, nil
	}
	can, needRank = rs.checkUserRank(ctx, userInfo.ID, userInfo.Rank, PermissionPrefix+action)
	return can, needRank, nil
}

// getUserPowerMapping get user power mapping
func (rs *RankService) getUserPowerMapping(ctx context.Context, userID string) (powerMapping map[string]bool) {
	powerMapping = make(map[string]bool, 0)
	userRole, err := rs.roleService.GetUserRole(ctx, userID)
	if err != nil {
		log.Error(err)
		return powerMapping
	}
	powers, err := rs.rolePowerService.GetRolePowerList(ctx, userRole)
	if err != nil {
		log.Error(err)
		return powerMapping
	}

	for _, power := range powers {
		powerMapping[power] = true
	}
	return powerMapping
}

// checkUserRank verify that the user meets the prestige criteria
func (rs *RankService) checkUserRank(ctx context.Context, userID string, userRank int, action string) (
	can bool, rank int) {
	// get the amount of rank required for the current operation
	requireRank, err := rs.configService.GetIntValue(ctx, action)
	if err != nil {
		log.Error(err)
		return false, requireRank
	}
	if userRank < requireRank || requireRank < 0 {
		log.Debugf("user %s want to do action %s, but rank %d < %d",
			userID, action, userRank, requireRank)
		return false, requireRank
	}
	return true, requireRank
}

// GetRankPersonalPage get personal comment list page
func (rs *RankService) GetRankPersonalPage(ctx context.Context, req *schema.GetRankPersonalWithPageReq) (
	pageModel *pager.PageModel, err error) {
	if plugin.RankAgentEnabled() {
		return pager.NewPageModel(0, []string{}), nil
	}
	if len(req.Username) > 0 {
		userInfo, exist, err := rs.userCommon.GetUserBasicInfoByUserName(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, errors.BadRequest(reason.UserNotFound)
		}
		req.UserID = userInfo.ID
	}
	if len(req.UserID) == 0 {
		return nil, errors.BadRequest(reason.UserNotFound)
	}

	userRankPage, total, err := rs.userRankRepo.UserRankPage(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	resp := rs.decorateRankPersonalPageResp(ctx, userRankPage)
	return pager.NewPageModel(total, resp), nil
}

func (rs *RankService) decorateRankPersonalPageResp(
	ctx context.Context, userRankPage []*entity.Activity) []*schema.GetRankPersonalPageResp {
	resp := make([]*schema.GetRankPersonalPageResp, 0)
	lang := handler.GetLangByCtx(ctx)

	for _, userRankInfo := range userRankPage {
		if len(userRankInfo.ObjectID) == 0 || userRankInfo.ObjectID == "0" {
			continue
		}
		objInfo, err := rs.objectInfoService.GetInfo(ctx, userRankInfo.ObjectID)
		if err != nil {
			log.Error(err)
			continue
		}

		commentResp := &schema.GetRankPersonalPageResp{
			CreatedAt:  userRankInfo.CreatedAt.Unix(),
			ObjectID:   userRankInfo.ObjectID,
			Reputation: userRankInfo.Rank,
		}
		cfg, err := rs.configService.GetConfigByID(ctx, userRankInfo.ActivityType)
		if err != nil {
			log.Error(err)
			continue
		}
		commentResp.RankType = translator.Tr(lang, activity_type.ActivityTypeFlagMapping[cfg.Key])
		commentResp.ObjectType = objInfo.ObjectType
		commentResp.Title = objInfo.Title
		commentResp.UrlTitle = htmltext.UrlTitle(objInfo.Title)
		commentResp.Content = objInfo.Content
		if objInfo.QuestionStatus == entity.QuestionStatusDeleted {
			commentResp.Title = translator.Tr(lang, constant.DeletedQuestionTitleTrKey)
		}
		commentResp.QuestionID = objInfo.QuestionID
		commentResp.AnswerID = objInfo.AnswerID
		resp = append(resp, commentResp)
	}
	return resp
}
