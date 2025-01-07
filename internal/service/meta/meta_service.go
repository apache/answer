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

package meta

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/apache/answer/internal/service/event_queue"
	"strconv"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	answercommon "github.com/apache/answer/internal/service/answer_common"
	metacommon "github.com/apache/answer/internal/service/meta_common"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/obj"
	myErrors "github.com/segmentfault/pacman/errors"
)

// MetaService user service
type MetaService struct {
	metaCommonService *metacommon.MetaCommonService
	userCommon        *usercommon.UserCommon
	questionRepo      questioncommon.QuestionRepo
	answerRepo        answercommon.AnswerRepo
	eventQueueService event_queue.EventQueueService
}

func NewMetaService(
	metaCommonService *metacommon.MetaCommonService,
	userCommon *usercommon.UserCommon,
	answerRepo answercommon.AnswerRepo,
	questionRepo questioncommon.QuestionRepo,
	eventQueueService event_queue.EventQueueService,
) *MetaService {
	return &MetaService{
		metaCommonService: metaCommonService,
		questionRepo:      questionRepo,
		userCommon:        userCommon,
		answerRepo:        answerRepo,
		eventQueueService: eventQueueService,
	}
}

// GetReactionByObjectId get reaction
func (ms *MetaService) GetReactionByObjectId(ctx context.Context, req *schema.GetReactionReq) (resp *schema.GetReactionByObjectIdResp, err error) {
	reactionMeta, err := ms.metaCommonService.GetMetaByObjectIdAndKey(ctx, req.ObjectID, entity.ObjectReactSummaryKey)

	// if not exist, return nil
	if err != nil {
		var pacmanErr *myErrors.Error
		if errors.As(err, &pacmanErr) && pacmanErr.Reason == reason.MetaObjectNotFound {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var reaction schema.ReactionsSummaryMeta
	err = json.Unmarshal([]byte(reactionMeta.Value), &reaction)
	if err != nil {
		return nil, err
	}
	return ms.convertToReactionResp(ctx, req.UserID, &reaction)
}

// AddOrUpdateReaction add or update reaction
func (ms *MetaService) AddOrUpdateReaction(ctx context.Context, req *schema.UpdateReactionReq) (resp *schema.GetReactionByObjectIdResp, err error) {
	// check if object exist and it's answer or question
	objectType, err := obj.GetObjectTypeStrByObjectID(req.ObjectID)
	if err != nil {
		return nil, err
	}
	var event *schema.EventMsg
	if objectType == constant.AnswerObjectType {
		answerInfo, exist, err := ms.answerRepo.GetAnswer(ctx, req.ObjectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, myErrors.BadRequest(reason.AnswerNotFound)
		}
		event = schema.NewEvent(constant.EventAnswerReact, req.UserID).TID(answerInfo.ID).
			AID(answerInfo.ID, answerInfo.UserID)
	} else if objectType == constant.QuestionObjectType {
		questionInfo, exist, err := ms.questionRepo.GetQuestion(ctx, req.ObjectID)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, myErrors.BadRequest(reason.QuestionNotFound)
		}
		event = schema.NewEvent(constant.EventQuestionReact, req.UserID).TID(questionInfo.ID).
			QID(questionInfo.ID, questionInfo.UserID)
	} else {
		return nil, myErrors.BadRequest(reason.ObjectNotFound)
	}

	// add or update
	reactions := &schema.ReactionsSummaryMeta{}
	err = ms.metaCommonService.AddOrUpdateMetaByObjectIdAndKey(ctx, req.ObjectID, entity.ObjectReactSummaryKey, func(meta *entity.Meta, exist bool) (*entity.Meta, error) {
		// if not exist, create new one
		if exist {
			_ = json.Unmarshal([]byte(meta.Value), reactions)
		}

		// update reaction
		ms.updateReaction(req, reactions)

		// write back to meta repo
		reactSumBytes, err := json.Marshal(reactions)
		if err != nil {
			return nil, err
		}

		return &entity.Meta{
			ObjectID: req.ObjectID,
			Key:      entity.ObjectReactSummaryKey,
			Value:    string(reactSumBytes),
		}, nil
	})

	if err != nil {
		return nil, myErrors.InternalServer(reason.DatabaseError).WithError(err)
	}

	resp, err = ms.convertToReactionResp(ctx, req.UserID, reactions)
	if err != nil {
		return nil, err
	}
	ms.eventQueueService.Send(ctx, event)
	return resp, nil
}

// updateReaction update reaction
func (ms *MetaService) updateReaction(req *schema.UpdateReactionReq, reactions *schema.ReactionsSummaryMeta) {
	if req.Reaction == "activate" {
		reactions.AddReactionSummary(req.Emoji, req.UserID)
	} else if req.Reaction == "deactivate" {
		reactions.RemoveReactionSummary(req.Emoji, req.UserID)
	}
}

func (ms *MetaService) convertToReactionResp(ctx context.Context, userId string, r *schema.ReactionsSummaryMeta) (
	resp *schema.GetReactionByObjectIdResp, err error) {
	lang := handler.GetLangByCtx(ctx)
	resp = &schema.GetReactionByObjectIdResp{
		ReactionSummary: make([]*schema.ReactionRespItem, 0),
	}
	// traverse map and convert to username
	for _, reaction := range r.Reactions {
		item := &schema.ReactionRespItem{
			Emoji:    reaction.Emoji,
			IsActive: r.CheckUserInReactionSummary(reaction.Emoji, userId),
		}

		usernames := make([]string, 0)
		userBasicInfos, err := ms.userCommon.BatchUserBasicInfoByID(ctx, reaction.UserIDs)
		item.Count = len(userBasicInfos)
		if err != nil {
			return resp, err
		}
		// get username
		for _, userBasicInfo := range userBasicInfos {
			usernames = append(usernames, userBasicInfo.Username)
			if len(usernames) == 5 && len(userBasicInfos) > 5 {
				item.Tooltip = translator.TrWithData(lang, constant.ReactionTooltipLabel, map[string]string{
					"Count": strconv.Itoa(len(userBasicInfos) - 5),
					"Names": strings.Join(usernames, ", "),
				})
				break
			}
		}
		if len(userBasicInfos) <= 5 {
			item.Tooltip = strings.Join(usernames, ", ")
		}
		resp.ReactionSummary = append(resp.ReactionSummary, item)
	}

	return resp, nil
}
