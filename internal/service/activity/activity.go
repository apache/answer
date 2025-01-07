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

package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apache/answer/internal/service/activity_common"
	"github.com/apache/answer/internal/service/meta_common"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/comment_common"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/object_info"
	"github.com/apache/answer/internal/service/revision_common"
	"github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/pkg/obj"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/log"
)

// ActivityRepo activity repository
type ActivityRepo interface {
	GetObjectAllActivity(ctx context.Context, objectID string, showVote bool) (activityList []*entity.Activity, err error)
}

// ActivityService activity service
type ActivityService struct {
	activityRepo          ActivityRepo
	userCommon            *usercommon.UserCommon
	activityCommonService *activity_common.ActivityCommon
	tagCommonService      *tag_common.TagCommonService
	objectInfoService     *object_info.ObjService
	commentCommonService  *comment_common.CommentCommonService
	revisionService       *revision_common.RevisionService
	metaService           *metacommon.MetaCommonService
	configService         *config.ConfigService
}

// NewActivityService new activity service
func NewActivityService(
	activityRepo ActivityRepo,
	userCommon *usercommon.UserCommon,
	activityCommonService *activity_common.ActivityCommon,
	tagCommonService *tag_common.TagCommonService,
	objectInfoService *object_info.ObjService,
	commentCommonService *comment_common.CommentCommonService,
	revisionService *revision_common.RevisionService,
	metaService *metacommon.MetaCommonService,
	configService *config.ConfigService,
) *ActivityService {
	return &ActivityService{
		objectInfoService:     objectInfoService,
		activityRepo:          activityRepo,
		userCommon:            userCommon,
		activityCommonService: activityCommonService,
		tagCommonService:      tagCommonService,
		commentCommonService:  commentCommonService,
		revisionService:       revisionService,
		metaService:           metaService,
		configService:         configService,
	}
}

// GetObjectTimeline get object timeline
func (as *ActivityService) GetObjectTimeline(ctx context.Context, req *schema.GetObjectTimelineReq) (
	resp *schema.GetObjectTimelineResp, err error) {
	resp = &schema.GetObjectTimelineResp{
		ObjectInfo: &schema.ActObjectInfo{},
		Timeline:   make([]*schema.ActObjectTimeline, 0),
	}

	resp.ObjectInfo, err = as.getTimelineMainObjInfo(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}

	activityList, err := as.activityRepo.GetObjectAllActivity(ctx, req.ObjectID, req.ShowVote)
	if err != nil {
		return nil, err
	}
	for _, act := range activityList {
		item := &schema.ActObjectTimeline{
			ActivityID: act.ID,
			RevisionID: converter.IntToString(act.RevisionID),
			CreatedAt:  act.CreatedAt.Unix(),
			Cancelled:  act.Cancelled == entity.ActivityCancelled,
			ObjectID:   act.ObjectID,
			UserInfo:   &schema.UserBasicInfo{},
		}
		item.ObjectType, _ = obj.GetObjectTypeStrByObjectID(act.ObjectID)
		if item.Cancelled {
			item.CancelledAt = act.CancelledAt.Unix()
		}

		if item.ObjectType == constant.QuestionObjectType || item.ObjectType == constant.AnswerObjectType {
			if handler.GetEnableShortID(ctx) {
				item.ObjectID = uid.EnShortID(act.ObjectID)
			}
		}

		cfg, err := as.configService.GetConfigByID(ctx, act.ActivityType)
		if err != nil {
			log.Errorf("fail to get config by id: %d, err: %v, act id is: %s", act.ActivityType, err, act.ID)
		} else {
			// database save activity type is number, change to activity type string is like "question.asked".
			// so we need to cut the front part of '.', only need string like 'asked'
			_, item.ActivityType, _ = strings.Cut(cfg.Key, ".")
			// format activity type string to show
			if isHidden, formattedActivityType := formatActivity(item.ActivityType); isHidden {
				continue
			} else {
				item.ActivityType = formattedActivityType
			}
		}

		// if activity is down vote, only admin can see who does it.
		if item.ActivityType == constant.ActDownVote && !req.IsAdmin {
			item.UserInfo.Username = "N/A"
			item.UserInfo.DisplayName = "N/A"
		} else {
			if act.TriggerUserID > 0 {
				item.UserInfo.ID = fmt.Sprintf("%d", act.TriggerUserID)
			} else {
				item.UserInfo.ID = act.UserID
			}
		}

		item.Comment = as.getTimelineActivityComment(ctx, item.ObjectID, item.ObjectType, item.ActivityType, item.RevisionID)
		resp.Timeline = append(resp.Timeline, item)
	}
	as.formatTimelineUserInfo(ctx, resp.Timeline)
	return
}

func (as *ActivityService) getTimelineMainObjInfo(ctx context.Context, objectID string) (
	resp *schema.ActObjectInfo, err error) {
	resp = &schema.ActObjectInfo{}
	objInfo, err := as.objectInfoService.GetInfo(ctx, objectID)
	if err != nil {
		return nil, err
	}
	resp.Title = objInfo.Title
	if objInfo.ObjectType == constant.TagObjectType {
		tag, exist, _ := as.tagCommonService.GetTagByID(ctx, objInfo.TagID)
		if exist {
			resp.Title = tag.SlugName
			resp.MainTagSlugName = tag.MainTagSlugName
		}
	}
	resp.ObjectType = objInfo.ObjectType
	resp.QuestionID = objInfo.QuestionID
	resp.AnswerID = objInfo.AnswerID
	if len(objInfo.ObjectCreatorUserID) > 0 {
		// get object creator user info
		userBasicInfo, exist, err := as.userCommon.GetUserBasicInfoByID(ctx, objInfo.ObjectCreatorUserID)
		if err != nil {
			return nil, err
		}
		if exist {
			resp.Username = userBasicInfo.Username
			resp.DisplayName = userBasicInfo.DisplayName
		}
	}
	return resp, nil
}

func (as *ActivityService) getTimelineActivityComment(ctx context.Context, objectID, objectType,
	activityType, revisionID string) (comment string) {
	if objectType == constant.CommentObjectType {
		commentInfo, err := as.commentCommonService.GetComment(ctx, objectID)
		if err != nil {
			log.Error(err)
		} else {
			return commentInfo.ParsedText
		}
		return
	}

	if activityType == constant.ActEdited {
		revision, err := as.revisionService.GetRevision(ctx, revisionID)
		if err != nil {
			log.Error(err)
		} else {
			return converter.Markdown2HTML(revision.Log)
		}
		return
	}
	if activityType == constant.ActClosed {
		// only question can be closed
		metaInfo, err := as.metaService.GetMetaByObjectIdAndKey(ctx, objectID, entity.QuestionCloseReasonKey)
		if err != nil {
			log.Error(err)
		} else {
			closeMsg := &schema.CloseQuestionMeta{}
			if err := json.Unmarshal([]byte(metaInfo.Value), closeMsg); err == nil {
				return converter.Markdown2HTML(closeMsg.CloseMsg)
			}
		}
	}
	return ""
}

func (as *ActivityService) formatTimelineUserInfo(ctx context.Context, timeline []*schema.ActObjectTimeline) {
	userExist := make(map[string]bool)
	userIDs := make([]string, 0)
	for _, info := range timeline {
		if len(info.UserInfo.ID) == 0 || userExist[info.UserInfo.ID] {
			continue
		}
		userIDs = append(userIDs, info.UserInfo.ID)
	}
	if len(userIDs) == 0 {
		return
	}
	userInfoMapping, err := as.userCommon.BatchUserBasicInfoByID(ctx, userIDs)
	if err != nil {
		log.Error(err)
		return
	}
	for _, info := range timeline {
		if len(info.UserInfo.ID) == 0 {
			continue
		}
		info.UserInfo = userInfoMapping[info.UserInfo.ID]
	}
}

// GetObjectTimelineDetail get object timeline
func (as *ActivityService) GetObjectTimelineDetail(ctx context.Context, req *schema.GetObjectTimelineDetailReq) (
	resp *schema.GetObjectTimelineDetailResp, err error) {
	resp = &schema.GetObjectTimelineDetailResp{}
	resp.OldRevision, _ = as.getOneObjectDetail(ctx, req.OldRevisionID)
	resp.NewRevision, _ = as.getOneObjectDetail(ctx, req.NewRevisionID)
	return resp, nil
}

// getOneObjectDetail get object detail
func (as *ActivityService) getOneObjectDetail(ctx context.Context, revisionID string) (
	resp *schema.ObjectTimelineDetail, err error) {
	resp = &schema.ObjectTimelineDetail{Tags: make([]*schema.ObjectTimelineTag, 0)}

	// if request revision is 0, return null object detail.
	if revisionID == "0" {
		return nil, nil
	}

	revision, err := as.revisionService.GetRevision(ctx, revisionID)
	if err != nil {
		log.Warn(err)
		return nil, nil
	}
	objInfo, err := as.objectInfoService.GetInfo(ctx, revision.ObjectID)
	if err != nil {
		return nil, err
	}

	switch objInfo.ObjectType {
	case constant.QuestionObjectType:
		data := &entity.QuestionWithTagsRevision{}
		if err = json.Unmarshal([]byte(revision.Content), data); err != nil {
			log.Errorf("revision parsing error %s", err)
			return resp, nil
		}
		for _, tag := range data.Tags {
			resp.Tags = append(resp.Tags, &schema.ObjectTimelineTag{
				SlugName:        tag.SlugName,
				DisplayName:     tag.DisplayName,
				MainTagSlugName: tag.MainTagSlugName,
				Recommend:       tag.Recommend,
				Reserved:        tag.Reserved,
			})
		}
		resp.Title = data.Title
		resp.OriginalText = data.OriginalText
	case constant.AnswerObjectType:
		data := &entity.Answer{}
		if err = json.Unmarshal([]byte(revision.Content), data); err != nil {
			log.Errorf("revision parsing error %s", err)
			return resp, nil
		}
		resp.Title = objInfo.Title // answer show question title
		resp.OriginalText = data.OriginalText
	case constant.TagObjectType:
		data := &entity.Tag{}
		if err = json.Unmarshal([]byte(revision.Content), data); err != nil {
			log.Errorf("revision parsing error %s", err)
			return resp, nil
		}
		resp.Title = data.DisplayName
		resp.OriginalText = data.OriginalText
		resp.SlugName = data.SlugName
		resp.MainTagSlugName = data.MainTagSlugName
	default:
		log.Errorf("unknown object type %s", objInfo.ObjectType)
	}
	return resp, nil
}

func formatActivity(activityType string) (isHidden bool, formattedActivityType string) {
	if activityType == constant.ActVotedUp ||
		activityType == constant.ActVotedDown ||
		activityType == constant.ActFollow {
		return true, ""
	}
	if activityType == constant.ActVoteUp {
		return false, constant.ActUpVote
	}
	if activityType == constant.ActVoteDown {
		return false, constant.ActDownVote
	}
	if activityType == constant.ActAccepted {
		return false, constant.ActAccept
	}
	return false, activityType
}
