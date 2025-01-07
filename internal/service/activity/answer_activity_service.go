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
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activity_type"
	"github.com/apache/answer/internal/service/config"
	"github.com/segmentfault/pacman/log"
)

// AnswerActivityRepo answer activity
type AnswerActivityRepo interface {
	SaveAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (err error)
	SaveCancelAcceptAnswerActivity(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (err error)
}

// AnswerActivityService answer activity service
type AnswerActivityService struct {
	answerActivityRepo AnswerActivityRepo
	configService      *config.ConfigService
}

// NewAnswerActivityService new comment service
func NewAnswerActivityService(
	answerActivityRepo AnswerActivityRepo,
	configService *config.ConfigService,
) *AnswerActivityService {
	return &AnswerActivityService{
		answerActivityRepo: answerActivityRepo,
		configService:      configService,
	}
}

// AcceptAnswer accept answer change activity
func (as *AnswerActivityService) AcceptAnswer(ctx context.Context,
	loginUserID, answerObjID, questionObjID, questionUserID, answerUserID string, isSelf bool) (err error) {
	log.Debugf("user %s want to accept answer %s[%s] for question %s[%s]", loginUserID,
		answerObjID, answerUserID,
		questionObjID, questionUserID)
	operationInfo := as.createAcceptAnswerOperationInfo(ctx, loginUserID,
		answerObjID, questionObjID, questionUserID, answerUserID, isSelf)
	return as.answerActivityRepo.SaveAcceptAnswerActivity(ctx, operationInfo)
}

// CancelAcceptAnswer cancel accept answer change activity
func (as *AnswerActivityService) CancelAcceptAnswer(ctx context.Context,
	loginUserID, answerObjID, questionObjID, questionUserID, answerUserID string) (err error) {
	operationInfo := as.createAcceptAnswerOperationInfo(ctx, loginUserID,
		answerObjID, questionObjID, questionUserID, answerUserID, false)
	return as.answerActivityRepo.SaveCancelAcceptAnswerActivity(ctx, operationInfo)
}

func (as *AnswerActivityService) createAcceptAnswerOperationInfo(ctx context.Context, loginUserID,
	answerObjID, questionObjID, questionUserID, answerUserID string, isSelf bool) *schema.AcceptAnswerOperationInfo {
	operationInfo := &schema.AcceptAnswerOperationInfo{
		TriggerUserID:    loginUserID,
		QuestionObjectID: questionObjID,
		QuestionUserID:   questionUserID,
		AnswerObjectID:   answerObjID,
		AnswerUserID:     answerUserID,
	}
	operationInfo.Activities = as.getActivities(ctx, operationInfo)
	if isSelf {
		for _, activity := range operationInfo.Activities {
			activity.Rank = 0
		}
	}
	return operationInfo
}

func (as *AnswerActivityService) getActivities(ctx context.Context, op *schema.AcceptAnswerOperationInfo) (
	activities []*schema.AcceptAnswerActivity) {
	activities = make([]*schema.AcceptAnswerActivity, 0)

	for _, action := range []string{activity_type.AnswerAccept, activity_type.AnswerAccepted} {
		t := &schema.AcceptAnswerActivity{}
		cfg, err := as.configService.GetConfigByKey(ctx, action)
		if err != nil {
			log.Warnf("get config by key error: %v", err)
			continue
		}
		t.ActivityType, t.Rank = cfg.ID, cfg.GetIntValue()

		if action == activity_type.AnswerAccept {
			t.ActivityUserID = op.QuestionUserID
			t.TriggerUserID = op.TriggerUserID
			t.OriginalObjectID = op.QuestionObjectID // if activity is 'accept' means this question is accept the answer.
		} else {
			t.ActivityUserID = op.AnswerUserID
			t.TriggerUserID = op.TriggerUserID
			t.OriginalObjectID = op.AnswerObjectID // if activity is 'accepted' means this answer was accepted.
		}
		activities = append(activities, t)
	}
	return activities
}
