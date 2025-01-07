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

package content

import (
	"context"
	"fmt"
	"github.com/apache/answer/internal/service/event_queue"
	"strings"

	"github.com/apache/answer/internal/service/activity_common"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity_type"
	"github.com/apache/answer/internal/service/comment_common"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/object_info"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/segmentfault/pacman/log"

	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/schema"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	"github.com/segmentfault/pacman/errors"
)

// VoteRepo activity repository
type VoteRepo interface {
	Vote(ctx context.Context, op *schema.VoteOperationInfo) (err error)
	CancelVote(ctx context.Context, op *schema.VoteOperationInfo) (err error)
	GetAndSaveVoteResult(ctx context.Context, objectID, objectType string) (up, down int64, err error)
	ListUserVotes(ctx context.Context, userID string, page int, pageSize int, activityTypes []int) (
		voteList []*entity.Activity, total int64, err error)
}

// VoteService user service
type VoteService struct {
	voteRepo          VoteRepo
	configService     *config.ConfigService
	questionRepo      questioncommon.QuestionRepo
	answerRepo        answercommon.AnswerRepo
	commentCommonRepo comment_common.CommentCommonRepo
	objectService     *object_info.ObjService
	activityRepo      activity_common.ActivityRepo
	eventQueueService event_queue.EventQueueService
}

func NewVoteService(
	voteRepo VoteRepo,
	configService *config.ConfigService,
	questionRepo questioncommon.QuestionRepo,
	answerRepo answercommon.AnswerRepo,
	commentCommonRepo comment_common.CommentCommonRepo,
	objectService *object_info.ObjService,
	eventQueueService event_queue.EventQueueService,
) *VoteService {
	return &VoteService{
		voteRepo:          voteRepo,
		configService:     configService,
		questionRepo:      questionRepo,
		answerRepo:        answerRepo,
		commentCommonRepo: commentCommonRepo,
		objectService:     objectService,
		eventQueueService: eventQueueService,
	}
}

// VoteUp vote up
func (vs *VoteService) VoteUp(ctx context.Context, req *schema.VoteReq) (resp *schema.VoteResp, err error) {
	objectInfo, err := vs.objectService.GetInfo(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}
	if objectInfo.IsDeleted() {
		return nil, errors.BadRequest(reason.NewObjectAlreadyDeleted)
	}
	// make object id must be decoded
	objectInfo.ObjectID = req.ObjectID

	// check user is voting self or not
	if objectInfo.ObjectCreatorUserID == req.UserID {
		return nil, errors.BadRequest(reason.DisallowVoteYourSelf)
	}

	voteUpOperationInfo := vs.createVoteOperationInfo(ctx, req.UserID, true, objectInfo)

	// vote operation
	if req.IsCancel {
		err = vs.voteRepo.CancelVote(ctx, voteUpOperationInfo)
	} else {
		// cancel vote down if exist
		voteOperationInfo := vs.createVoteOperationInfo(ctx, req.UserID, false, objectInfo)
		err = vs.voteRepo.CancelVote(ctx, voteOperationInfo)
		if err != nil {
			return nil, err
		}
		err = vs.voteRepo.Vote(ctx, voteUpOperationInfo)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}

	resp = &schema.VoteResp{}
	resp.UpVotes, resp.DownVotes, err = vs.voteRepo.GetAndSaveVoteResult(ctx, req.ObjectID, objectInfo.ObjectType)
	if err != nil {
		log.Error(err)
	}
	resp.Votes = resp.UpVotes - resp.DownVotes
	if !req.IsCancel {
		resp.VoteStatus = constant.ActVoteUp
		vs.sendEvent(ctx, req, objectInfo, resp)
	}
	return resp, nil
}

// VoteDown vote down
func (vs *VoteService) VoteDown(ctx context.Context, req *schema.VoteReq) (resp *schema.VoteResp, err error) {
	objectInfo, err := vs.objectService.GetInfo(ctx, req.ObjectID)
	if err != nil {
		return nil, err
	}
	if objectInfo.IsDeleted() {
		return nil, errors.BadRequest(reason.NewObjectAlreadyDeleted)
	}
	// make object id must be decoded
	objectInfo.ObjectID = req.ObjectID

	// check user is voting self or not
	if objectInfo.ObjectCreatorUserID == req.UserID {
		return nil, errors.BadRequest(reason.DisallowVoteYourSelf)
	}

	// vote operation
	voteDownOperationInfo := vs.createVoteOperationInfo(ctx, req.UserID, false, objectInfo)
	if req.IsCancel {
		err = vs.voteRepo.CancelVote(ctx, voteDownOperationInfo)
		if err != nil {
			return nil, err
		}
	} else {
		// cancel vote up if exist
		err = vs.voteRepo.CancelVote(ctx, vs.createVoteOperationInfo(ctx, req.UserID, true, objectInfo))
		if err != nil {
			return nil, err
		}
		err = vs.voteRepo.Vote(ctx, voteDownOperationInfo)
		if err != nil {
			return nil, err
		}
	}

	resp = &schema.VoteResp{}
	resp.UpVotes, resp.DownVotes, err = vs.voteRepo.GetAndSaveVoteResult(ctx, req.ObjectID, objectInfo.ObjectType)
	if err != nil {
		log.Error(err)
	}
	resp.Votes = resp.UpVotes - resp.DownVotes
	if !req.IsCancel {
		resp.VoteStatus = constant.ActVoteDown
		vs.sendEvent(ctx, req, objectInfo, resp)
	}
	return resp, nil
}

// ListUserVotes list user's votes
func (vs *VoteService) ListUserVotes(ctx context.Context, req schema.GetVoteWithPageReq) (resp *pager.PageModel, err error) {
	typeKeys := []string{
		activity_type.QuestionVoteUp,
		activity_type.QuestionVoteDown,
		activity_type.AnswerVoteUp,
		activity_type.AnswerVoteDown,
	}
	activityTypes := make([]int, 0)
	activityTypeMapping := make(map[int]string, 0)

	for _, typeKey := range typeKeys {
		cfg, err := vs.configService.GetConfigByKey(ctx, typeKey)
		if err != nil {
			continue
		}
		activityTypes = append(activityTypes, cfg.ID)
		activityTypeMapping[cfg.ID] = typeKey
	}

	voteList, total, err := vs.voteRepo.ListUserVotes(ctx, req.UserID, req.Page, req.PageSize, activityTypes)
	if err != nil {
		return nil, err
	}

	lang := handler.GetLangByCtx(ctx)

	votes := make([]*schema.GetVoteWithPageResp, 0)
	for _, voteInfo := range voteList {
		objInfo, err := vs.objectService.GetInfo(ctx, voteInfo.ObjectID)
		if err != nil {
			log.Error(err)
			continue
		}

		item := &schema.GetVoteWithPageResp{
			CreatedAt:  voteInfo.CreatedAt.Unix(),
			ObjectID:   objInfo.ObjectID,
			QuestionID: objInfo.QuestionID,
			AnswerID:   objInfo.AnswerID,
			ObjectType: objInfo.ObjectType,
			Title:      objInfo.Title,
			UrlTitle:   htmltext.UrlTitle(objInfo.Title),
			Content:    objInfo.Content,
		}
		item.VoteType = translator.Tr(lang,
			activity_type.ActivityTypeFlagMapping[activityTypeMapping[voteInfo.ActivityType]])
		if objInfo.QuestionStatus == entity.QuestionStatusDeleted {
			item.Title = translator.Tr(lang, constant.DeletedQuestionTitleTrKey)
		}
		votes = append(votes, item)
	}
	return pager.NewPageModel(total, votes), err
}

func (vs *VoteService) createVoteOperationInfo(ctx context.Context,
	userID string, voteUp bool, objectInfo *schema.SimpleObjectInfo) *schema.VoteOperationInfo {
	// warp vote operation
	voteOperationInfo := &schema.VoteOperationInfo{
		ObjectID:            objectInfo.ObjectID,
		ObjectType:          objectInfo.ObjectType,
		ObjectCreatorUserID: objectInfo.ObjectCreatorUserID,
		OperatingUserID:     userID,
		VoteUp:              voteUp,
		VoteDown:            !voteUp,
	}
	voteOperationInfo.Activities = vs.getActivities(ctx, voteOperationInfo)
	return voteOperationInfo
}

func (vs *VoteService) getActivities(ctx context.Context, op *schema.VoteOperationInfo) (
	activities []*schema.VoteActivity) {
	activities = make([]*schema.VoteActivity, 0)

	var actions []string
	switch op.ObjectType {
	case constant.QuestionObjectType:
		if op.VoteUp {
			actions = []string{activity_type.QuestionVoteUp, activity_type.QuestionVotedUp}
		} else {
			actions = []string{activity_type.QuestionVoteDown, activity_type.QuestionVotedDown}
		}
	case constant.AnswerObjectType:
		if op.VoteUp {
			actions = []string{activity_type.AnswerVoteUp, activity_type.AnswerVotedUp}
		} else {
			actions = []string{activity_type.AnswerVoteDown, activity_type.AnswerVotedDown}
		}
	case constant.CommentObjectType:
		actions = []string{activity_type.CommentVoteUp}
	}

	for _, action := range actions {
		t := &schema.VoteActivity{}
		cfg, err := vs.configService.GetConfigByKey(ctx, action)
		if err != nil {
			log.Warnf("get config by key error: %v", err)
			continue
		}
		t.ActivityType, t.Rank = cfg.ID, cfg.GetIntValue()

		if strings.Contains(action, "voted") {
			t.ActivityUserID = op.ObjectCreatorUserID
			t.TriggerUserID = op.OperatingUserID
		} else {
			t.ActivityUserID = op.OperatingUserID
			t.TriggerUserID = "0"
		}
		activities = append(activities, t)
	}
	return activities
}

func (vs *VoteService) sendEvent(ctx context.Context,
	req *schema.VoteReq, objectInfo *schema.SimpleObjectInfo, resp *schema.VoteResp) {
	var event *schema.EventMsg
	switch objectInfo.ObjectType {
	case constant.QuestionObjectType:
		event = schema.NewEvent(constant.EventQuestionVote, req.UserID).TID(objectInfo.QuestionID).
			QID(objectInfo.QuestionID, objectInfo.ObjectCreatorUserID)
	case constant.AnswerObjectType:
		event = schema.NewEvent(constant.EventAnswerVote, req.UserID).TID(objectInfo.AnswerID).
			AID(objectInfo.AnswerID, objectInfo.ObjectCreatorUserID)
	case constant.CommentObjectType:
		event = schema.NewEvent(constant.EventCommentVote, req.UserID).TID(objectInfo.CommentID).
			CID(objectInfo.CommentID, objectInfo.ObjectCreatorUserID)
	default:
		return
	}
	event.AddExtra("vote_up_amount", fmt.Sprintf("%d", resp.UpVotes))
	event.AddExtra("vote_down_amount", fmt.Sprintf("%d", resp.DownVotes))
	vs.eventQueueService.Send(ctx, event)
}
