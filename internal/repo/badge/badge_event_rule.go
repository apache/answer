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
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/badge"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"strconv"
)

// eventRuleRepo event rule repo
type eventRuleRepo struct {
	data             *data.Data
	EventRuleMapping map[constant.EventType][]badge.EventRuleHandler
}

// NewEventRuleRepo creates a new badge repository
func NewEventRuleRepo(data *data.Data) badge.EventRuleRepo {
	b := &eventRuleRepo{
		data: data,
	}
	b.EventRuleMapping = map[constant.EventType][]badge.EventRuleHandler{
		constant.EventUserUpdate:     {b.FirstUpdateUserProfile},
		constant.EventUserShare:      {b.FirstSharedPost},
		constant.EventQuestionCreate: nil,
		constant.EventQuestionUpdate: {b.FirstPostEdit},
		constant.EventQuestionDelete: nil,
		constant.EventQuestionVote:   {b.FirstVotedPost, b.ReachQuestionVote},
		constant.EventQuestionAccept: {b.FirstAcceptAnswer, b.ReachAnswerAcceptedAmount},
		constant.EventQuestionFlag:   {b.FirstFlaggedPost},
		constant.EventQuestionReact:  {b.FirstReactedPost},
		constant.EventAnswerCreate:   nil,
		constant.EventAnswerUpdate:   {b.FirstPostEdit},
		constant.EventAnswerDelete:   nil,
		constant.EventAnswerVote:     {b.FirstVotedPost, b.ReachAnswerVote},
		constant.EventAnswerFlag:     {b.FirstFlaggedPost},
		constant.EventAnswerReact:    {b.FirstReactedPost},
		constant.EventCommentCreate:  nil,
		constant.EventCommentUpdate:  nil,
		constant.EventCommentDelete:  nil,
		constant.EventCommentVote:    {b.FirstVotedPost},
		constant.EventCommentFlag:    {b.FirstFlaggedPost},
	}
	return b
}

// HandleEventWithRule handle event with rule
func (br *eventRuleRepo) HandleEventWithRule(ctx context.Context, msg *schema.EventMsg) (
	awards []*entity.BadgeAward) {
	handlers := br.EventRuleMapping[msg.EventType]
	for _, handler := range handlers {
		t, err := handler(ctx, msg)
		if err != nil {
			log.Errorf("error handling badge event %+v: %v", msg, err)
		} else {
			awards = append(awards, t...)
		}
	}
	return awards
}

// FirstUpdateUserProfile first update user profile
func (br *eventRuleRepo) FirstUpdateUserProfile(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "FirstUpdateUserProfile")
	for _, b := range badges {
		bean := &entity.User{ID: event.UserID}
		exist, err := br.data.DB.Context(ctx).Get(bean)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		if !exist {
			continue
		}
		if len(bean.Bio) > 0 {
			awards = append(awards, br.createBadgeAward(event.UserID, entity.BadgeEmptyAwardKey, b))
		}
	}
	return awards, nil
}

// FirstPostEdit first post edit
func (br *eventRuleRepo) FirstPostEdit(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "FirstPostEdit")
	for _, b := range badges {
		awards = append(awards, br.createBadgeAward(event.UserID, event.GetObjectID(), b))
	}
	return awards, nil
}

// FirstFlaggedPost first flagged post.
func (br *eventRuleRepo) FirstFlaggedPost(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "FirstFlaggedPost")
	for _, b := range badges {
		awards = append(awards, br.createBadgeAward(event.UserID, event.GetObjectID(), b))
	}
	return awards, nil
}

// FirstVotedPost first voted post
func (br *eventRuleRepo) FirstVotedPost(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "FirstVotedPost")
	for _, b := range badges {
		awards = append(awards, br.createBadgeAward(event.UserID, event.GetObjectID(), b))
	}
	return awards, nil
}

// FirstReactedPost first reacted post
func (br *eventRuleRepo) FirstReactedPost(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "FirstReactedPost")
	for _, b := range badges {
		awards = append(awards, br.createBadgeAward(event.UserID, event.GetObjectID(), b))
	}
	return awards, nil
}

// FirstSharedPost first shared post
func (br *eventRuleRepo) FirstSharedPost(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "FirstSharedPost")
	for _, b := range badges {
		awards = append(awards, br.createBadgeAward(event.UserID, event.GetObjectID(), b))
	}
	return awards, nil
}

// FirstAcceptAnswer user first accept answer
func (br *eventRuleRepo) FirstAcceptAnswer(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "FirstAcceptAnswer")
	for _, b := range badges {
		awards = append(awards, br.createBadgeAward(event.UserID, event.GetObjectID(), b))
	}
	return awards, nil
}

// ReachAnswerAcceptedAmount reach answer accepted amount
func (br *eventRuleRepo) ReachAnswerAcceptedAmount(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "ReachAnswerAcceptedAmount")
	if len(event.AnswerUserID) == 0 {
		return nil, nil
	}

	// count user's accepted answer amount
	amount, err := br.data.DB.Context(ctx).Count(&entity.Answer{
		UserID:   event.AnswerUserID,
		Accepted: schema.AnswerAcceptedEnable,
		Status:   entity.AnswerStatusAvailable,
	})
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	for _, b := range badges {
		// get badge requirement
		requirement := b.GetIntParam("amount")
		if requirement == 0 || amount < requirement {
			continue
		}
		awards = append(awards, br.createBadgeAward(event.AnswerUserID, event.AnswerID, b))
	}
	return awards, nil
}

// ReachAnswerVote reach answer vote
func (br *eventRuleRepo) ReachAnswerVote(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "ReachAnswerVote")
	// get vote amount
	amount, _ := strconv.Atoi(event.GetExtra("vote_up_amount"))
	if amount == 0 {
		return nil, nil
	}

	for _, b := range badges {
		// get badge requirement
		requirement := b.GetIntParam("amount")
		if requirement == 0 || int64(amount) < requirement {
			continue
		}
		awards = append(awards, br.createBadgeAward(event.AnswerUserID, event.AnswerID, b))
	}
	return awards, nil
}

// ReachQuestionVote reach question vote
func (br *eventRuleRepo) ReachQuestionVote(ctx context.Context,
	event *schema.EventMsg) (awards []*entity.BadgeAward, err error) {
	badges := br.getBadgesByHandler(ctx, "ReachQuestionVote")
	// get vote amount
	amount, _ := strconv.Atoi(event.GetExtra("vote_up_amount"))
	if amount == 0 {
		return nil, nil
	}

	for _, b := range badges {
		// get badge requirement
		requirement := b.GetIntParam("amount")
		if requirement == 0 || int64(amount) < requirement {
			continue
		}
		awards = append(awards, br.createBadgeAward(event.QuestionUserID, event.QuestionID, b))
	}
	return awards, nil
}

func (br *eventRuleRepo) getBadgesByHandler(ctx context.Context, handler string) (badges []*entity.Badge) {
	badges = make([]*entity.Badge, 0)
	err := br.data.DB.Context(ctx).Where("handler = ?", handler).Find(&badges)
	if err != nil {
		log.Errorf("error getting badge by handler %s: %v", handler, err)
		return nil
	}
	return badges
}

func (br *eventRuleRepo) createBadgeAward(userID, awardKey string, badge *entity.Badge) (awards *entity.BadgeAward) {
	return &entity.BadgeAward{
		UserID:   userID,
		BadgeID:  badge.ID,
		AwardKey: awardKey,
	}
}
