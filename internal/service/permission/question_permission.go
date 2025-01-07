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

package permission

import (
	"context"
	"github.com/apache/answer/internal/entity"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/schema"
)

// GetQuestionPermission get question permission
func GetQuestionPermission(ctx context.Context, userID string, creatorUserID string, status int,
	canEdit, canDelete, canClose, canReopen, canPin, canHide, canUnPin, canShow, canRecover bool) (
	actions []*schema.PermissionMemberAction) {
	lang := handler.GetLangByCtx(ctx)
	actions = make([]*schema.PermissionMemberAction, 0)
	if len(userID) > 0 {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "report",
			Name:   translator.Tr(lang, reportActionName),
			Type:   "reason",
		})
	}
	if (canEdit || userID == creatorUserID) && status != entity.QuestionStatusDeleted {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "edit",
			Name:   translator.Tr(lang, editActionName),
			Type:   "edit",
		})
	}
	if canClose && status == entity.QuestionStatusAvailable {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "close",
			Name:   translator.Tr(lang, closeActionName),
			Type:   "confirm",
		})
	}
	if canReopen {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "reopen",
			Name:   translator.Tr(lang, reopenActionName),
			Type:   "confirm",
		})
	}
	if canPin {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "pin",
			Name:   translator.Tr(lang, pinActionName),
			Type:   "confirm",
		})
	}
	if canHide {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "hide",
			Name:   translator.Tr(lang, hideActionName),
			Type:   "confirm",
		})
	}

	if canUnPin {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "unpin",
			Name:   translator.Tr(lang, unpinActionName),
			Type:   "confirm",
		})
	}

	if canShow {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "show",
			Name:   translator.Tr(lang, showActionName),
			Type:   "confirm",
		})
	}

	if (canDelete || userID == creatorUserID) && status != entity.QuestionStatusDeleted {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "delete",
			Name:   translator.Tr(lang, deleteActionName),
			Type:   "confirm",
		})
	}

	if canRecover && status == entity.QuestionStatusDeleted {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "undelete",
			Name:   translator.Tr(lang, undeleteActionName),
			Type:   "confirm",
		})
	}
	return actions
}

// GetQuestionExtendsPermission get question extends permission
func GetQuestionExtendsPermission(ctx context.Context, canInviteOtherToAnswer bool) (
	actions []*schema.PermissionMemberAction) {
	lang := handler.GetLangByCtx(ctx)
	actions = make([]*schema.PermissionMemberAction, 0)
	if canInviteOtherToAnswer {
		actions = append(actions, &schema.PermissionMemberAction{
			Action: "invite_other_to_answer",
			Name:   translator.Tr(lang, inviteSomeoneToAnswerActionName),
			Type:   "confirm",
		})
	}
	return actions
}
