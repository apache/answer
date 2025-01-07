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

package migrations

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/permission"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func addRoleFeatures(ctx context.Context, x *xorm.Engine) error {
	err := x.Context(ctx).Sync(new(entity.Role), new(entity.RolePowerRel), new(entity.Power), new(entity.UserRoleRel))
	if err != nil {
		return err
	}

	roles := []*entity.Role{
		{ID: 1, Name: "User", Description: "Default with no special access."},
		{ID: 2, Name: "Admin", Description: "Have the full power to access the site."},
		{ID: 3, Name: "Moderator", Description: "Has access to all posts except admin settings."},
	}

	// insert default roles
	for _, role := range roles {
		exist, err := x.Context(ctx).Get(&entity.Role{ID: role.ID, Name: role.Name})
		if err != nil {
			return err
		}
		if exist {
			continue
		}
		_, err = x.Context(ctx).Insert(role)
		if err != nil {
			return err
		}
	}

	powers := []*entity.Power{
		{ID: 1, Name: "admin access", PowerType: permission.AdminAccess, Description: "admin access"},
		{ID: 2, Name: "question add", PowerType: permission.QuestionAdd, Description: "question add"},
		{ID: 3, Name: "question edit", PowerType: permission.QuestionEdit, Description: "question edit"},
		{ID: 4, Name: "question edit without review", PowerType: permission.QuestionEditWithoutReview, Description: "question edit without review"},
		{ID: 5, Name: "question delete", PowerType: permission.QuestionDelete, Description: "question delete"},
		{ID: 6, Name: "question close", PowerType: permission.QuestionClose, Description: "question close"},
		{ID: 7, Name: "question reopen", PowerType: permission.QuestionReopen, Description: "question reopen"},
		{ID: 8, Name: "question vote up", PowerType: permission.QuestionVoteUp, Description: "question vote up"},
		{ID: 9, Name: "question vote down", PowerType: permission.QuestionVoteDown, Description: "question vote down"},
		{ID: 10, Name: "answer add", PowerType: permission.AnswerAdd, Description: "answer add"},
		{ID: 11, Name: "answer edit", PowerType: permission.AnswerEdit, Description: "answer edit"},
		{ID: 12, Name: "answer edit without review", PowerType: permission.AnswerEditWithoutReview, Description: "answer edit without review"},
		{ID: 13, Name: "answer delete", PowerType: permission.AnswerDelete, Description: "answer delete"},
		{ID: 14, Name: "answer accept", PowerType: permission.AnswerAccept, Description: "answer accept"},
		{ID: 15, Name: "answer vote up", PowerType: permission.AnswerVoteUp, Description: "answer vote up"},
		{ID: 16, Name: "answer vote down", PowerType: permission.AnswerVoteDown, Description: "answer vote down"},
		{ID: 17, Name: "comment add", PowerType: permission.CommentAdd, Description: "comment add"},
		{ID: 18, Name: "comment edit", PowerType: permission.CommentEdit, Description: "comment edit"},
		{ID: 19, Name: "comment delete", PowerType: permission.CommentDelete, Description: "comment delete"},
		{ID: 20, Name: "comment vote up", PowerType: permission.CommentVoteUp, Description: "comment vote up"},
		{ID: 21, Name: "comment vote down", PowerType: permission.CommentVoteDown, Description: "comment vote down"},
		{ID: 22, Name: "report add", PowerType: permission.ReportAdd, Description: "report add"},
		{ID: 23, Name: "tag add", PowerType: permission.TagAdd, Description: "tag add"},
		{ID: 24, Name: "tag edit", PowerType: permission.TagEdit, Description: "tag edit"},
		{ID: 25, Name: "tag edit without review", PowerType: permission.TagEditWithoutReview, Description: "tag edit without review"},
		{ID: 26, Name: "tag edit slug name", PowerType: permission.TagEditSlugName, Description: "tag edit slug name"},
		{ID: 27, Name: "tag delete", PowerType: permission.TagDelete, Description: "tag delete"},
		{ID: 28, Name: "tag synonym", PowerType: permission.TagSynonym, Description: "tag synonym"},
		{ID: 29, Name: "link url limit", PowerType: permission.LinkUrlLimit, Description: "link url limit"},
		{ID: 30, Name: "vote detail", PowerType: permission.VoteDetail, Description: "vote detail"},
		{ID: 31, Name: "answer audit", PowerType: permission.AnswerAudit, Description: "answer audit"},
		{ID: 32, Name: "question audit", PowerType: permission.QuestionAudit, Description: "question audit"},
		{ID: 33, Name: "tag audit", PowerType: permission.TagAudit, Description: "tag audit"},
	}
	// insert default powers
	for _, power := range powers {
		exist, err := x.Context(ctx).Get(&entity.Power{ID: power.ID})
		if err != nil {
			return err
		}
		if exist {
			_, err = x.Context(ctx).ID(power.ID).Update(power)
		} else {
			_, err = x.Context(ctx).Insert(power)
		}
		if err != nil {
			return err
		}
	}

	rolePowerRels := []*entity.RolePowerRel{
		{RoleID: 2, PowerType: permission.AdminAccess},
		{RoleID: 2, PowerType: permission.QuestionAdd},
		{RoleID: 2, PowerType: permission.QuestionEdit},
		{RoleID: 2, PowerType: permission.QuestionEditWithoutReview},
		{RoleID: 2, PowerType: permission.QuestionDelete},
		{RoleID: 2, PowerType: permission.QuestionClose},
		{RoleID: 2, PowerType: permission.QuestionReopen},
		{RoleID: 2, PowerType: permission.QuestionVoteUp},
		{RoleID: 2, PowerType: permission.QuestionVoteDown},
		{RoleID: 2, PowerType: permission.AnswerAdd},
		{RoleID: 2, PowerType: permission.AnswerEdit},
		{RoleID: 2, PowerType: permission.AnswerEditWithoutReview},
		{RoleID: 2, PowerType: permission.AnswerDelete},
		{RoleID: 2, PowerType: permission.AnswerAccept},
		{RoleID: 2, PowerType: permission.AnswerVoteUp},
		{RoleID: 2, PowerType: permission.AnswerVoteDown},
		{RoleID: 2, PowerType: permission.CommentAdd},
		{RoleID: 2, PowerType: permission.CommentEdit},
		{RoleID: 2, PowerType: permission.CommentDelete},
		{RoleID: 2, PowerType: permission.CommentVoteUp},
		{RoleID: 2, PowerType: permission.CommentVoteDown},
		{RoleID: 2, PowerType: permission.ReportAdd},
		{RoleID: 2, PowerType: permission.TagAdd},
		{RoleID: 2, PowerType: permission.TagEdit},
		{RoleID: 2, PowerType: permission.TagEditSlugName},
		{RoleID: 2, PowerType: permission.TagEditWithoutReview},
		{RoleID: 2, PowerType: permission.TagDelete},
		{RoleID: 2, PowerType: permission.TagSynonym},
		{RoleID: 2, PowerType: permission.LinkUrlLimit},
		{RoleID: 2, PowerType: permission.VoteDetail},
		{RoleID: 2, PowerType: permission.AnswerAudit},
		{RoleID: 2, PowerType: permission.QuestionAudit},
		{RoleID: 2, PowerType: permission.TagAudit},
		{RoleID: 2, PowerType: permission.TagUseReservedTag},

		{RoleID: 3, PowerType: permission.QuestionAdd},
		{RoleID: 3, PowerType: permission.QuestionEdit},
		{RoleID: 3, PowerType: permission.QuestionEditWithoutReview},
		{RoleID: 3, PowerType: permission.QuestionDelete},
		{RoleID: 3, PowerType: permission.QuestionClose},
		{RoleID: 3, PowerType: permission.QuestionReopen},
		{RoleID: 3, PowerType: permission.QuestionVoteUp},
		{RoleID: 3, PowerType: permission.QuestionVoteDown},
		{RoleID: 3, PowerType: permission.AnswerAdd},
		{RoleID: 3, PowerType: permission.AnswerEdit},
		{RoleID: 3, PowerType: permission.AnswerEditWithoutReview},
		{RoleID: 3, PowerType: permission.AnswerDelete},
		{RoleID: 3, PowerType: permission.AnswerAccept},
		{RoleID: 3, PowerType: permission.AnswerVoteUp},
		{RoleID: 3, PowerType: permission.AnswerVoteDown},
		{RoleID: 3, PowerType: permission.CommentAdd},
		{RoleID: 3, PowerType: permission.CommentEdit},
		{RoleID: 3, PowerType: permission.CommentDelete},
		{RoleID: 3, PowerType: permission.CommentVoteUp},
		{RoleID: 3, PowerType: permission.CommentVoteDown},
		{RoleID: 3, PowerType: permission.ReportAdd},
		{RoleID: 3, PowerType: permission.TagAdd},
		{RoleID: 3, PowerType: permission.TagEdit},
		{RoleID: 3, PowerType: permission.TagEditSlugName},
		{RoleID: 3, PowerType: permission.TagEditWithoutReview},
		{RoleID: 3, PowerType: permission.TagDelete},
		{RoleID: 3, PowerType: permission.TagSynonym},
		{RoleID: 3, PowerType: permission.LinkUrlLimit},
		{RoleID: 3, PowerType: permission.VoteDetail},
		{RoleID: 3, PowerType: permission.AnswerAudit},
		{RoleID: 3, PowerType: permission.QuestionAudit},
		{RoleID: 3, PowerType: permission.TagAudit},
		{RoleID: 3, PowerType: permission.TagUseReservedTag},
	}

	// insert default powers
	for _, rel := range rolePowerRels {
		exist, err := x.Context(ctx).Get(&entity.RolePowerRel{RoleID: rel.RoleID, PowerType: rel.PowerType})
		if err != nil {
			return err
		}
		if exist {
			continue
		}
		_, err = x.Context(ctx).Insert(rel)
		if err != nil {
			return err
		}
	}

	adminUserRoleRel := &entity.UserRoleRel{
		UserID: "1",
		RoleID: 2,
	}

	exist, err := x.Context(ctx).Get(adminUserRoleRel)
	if err != nil {
		return err
	}
	if !exist {
		_, err = x.Context(ctx).Insert(adminUserRoleRel)
		if err != nil {
			return err
		}
	}

	defaultConfigTable := []*entity.Config{
		{ID: 115, Key: "rank.question.close", Value: `-1`},
		{ID: 116, Key: "rank.question.reopen", Value: `-1`},
		{ID: 117, Key: "rank.tag.use_reserved_tag", Value: `-1`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Context(ctx).Get(&entity.Config{ID: c.ID, Key: c.Key})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			if _, err = x.Context(ctx).Update(c, &entity.Config{ID: c.ID, Key: c.Key}); err != nil {
				log.Errorf("update %+v config failed: %s", c, err)
				return fmt.Errorf("update config failed: %w", err)
			}
			continue
		}
		if _, err = x.Context(ctx).Insert(&entity.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			log.Errorf("insert %+v config failed: %s", c, err)
			return fmt.Errorf("add config failed: %w", err)
		}
	}
	return nil
}
