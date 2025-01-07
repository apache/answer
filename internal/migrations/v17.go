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

func addRecoverPermission(ctx context.Context, x *xorm.Engine) error {
	powers := []*entity.Power{
		{ID: 39, Name: "recover answer", PowerType: permission.AnswerUnDelete, Description: "recover deleted answer"},
		{ID: 40, Name: "recover question", PowerType: permission.QuestionUnDelete, Description: "recover deleted question"},
		{ID: 41, Name: "recover tag", PowerType: permission.TagUnDelete, Description: "recover deleted tag"},
	}
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
		{RoleID: 2, PowerType: permission.AnswerUnDelete},
		{RoleID: 2, PowerType: permission.QuestionUnDelete},
		{RoleID: 2, PowerType: permission.TagUnDelete},

		{RoleID: 3, PowerType: permission.AnswerUnDelete},
		{RoleID: 3, PowerType: permission.QuestionUnDelete},
		{RoleID: 3, PowerType: permission.TagUnDelete},
	}
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

	defaultConfigTable := []*entity.Config{
		{ID: 128, Key: "rank.answer.undeleted", Value: `-1`},
		{ID: 129, Key: "rank.question.undeleted", Value: `-1`},
		{ID: 130, Key: "rank.tag.undeleted", Value: `-1`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Context(ctx).Get(&entity.Config{ID: c.ID})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			if _, err = x.Context(ctx).Update(c, &entity.Config{ID: c.ID}); err != nil {
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
