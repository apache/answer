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

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/plugin"
	"xorm.io/xorm"
)

// avatarData is the JSON structure stored in the user.avatar column.
type avatarData struct {
	Type     string `json:"type"`
	Gravatar string `json:"gravatar,omitempty"`
	Custom   string `json:"custom,omitempty"`
}

func fixAvatar(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState) (fixResult, error) {
	result := fixResult{}

	err := runBatchedFix(state,
		func(offset int) ([]*entity.User, error) { return queryCustomAvatarUsers(x, offset) },
		func(user *entity.User) error { return processAvatarUser(x, &result, opts, state, user) })
	if err != nil {
		return fixResult{}, err
	}

	frResult, err := fixFileRecordBySources(x, opts, state,
		[]string{string(plugin.UserAvatar)}, fixTypeAvatar)
	if err != nil {
		return fixResult{}, err
	}
	result.add(frResult)

	return result, nil
}

func queryCustomAvatarUsers(x *xorm.Engine, offset int) ([]*entity.User, error) {
	users := make([]*entity.User, 0)
	err := x.Select("id, avatar").
		Where("avatar LIKE ?", `%"custom":"%`).
		OrderBy("id").
		Limit(fixBatchSize, offset).
		Find(&users)
	if err != nil {
		return nil, fmt.Errorf("query user failed: %w", err)
	}
	return users, nil
}

func processAvatarUser(x *xorm.Engine, result *fixResult, opts *FixURLPrefixOptions, state *fixRunState, user *entity.User) error {
	avatar := &avatarData{}
	if err := json.Unmarshal([]byte(user.Avatar), avatar); err != nil {
		return nil
	}
	if avatar.Type != constant.AvatarTypeCustom || len(avatar.Custom) == 0 {
		return nil
	}
	return applyValueFix(result, opts, state, fixTypeAvatar, fmt.Sprintf("user %s", user.ID), avatar.Custom,
		func(newCustom string) error {
			avatar.Custom = newCustom
			newJSON, err := json.Marshal(avatar)
			if err != nil {
				return fmt.Errorf("marshal avatar JSON (user=%s) failed: %w", user.ID, err)
			}
			if _, err := x.ID(user.ID).Cols("avatar").NoAutoTime().Update(&entity.User{Avatar: string(newJSON)}); err != nil {
				return fmt.Errorf("update user avatar (id=%s) failed: %w", user.ID, err)
			}
			return nil
		})
}
