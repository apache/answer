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

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/plugin"
	"xorm.io/xorm"
)

func fixAvatar(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState) (fixResult, error) {
	result := fixResult{}
	records, err := newFileRecordSync(x, opts, fixTypeAvatar, string(plugin.UserAvatar))
	if err != nil {
		return fixResult{}, err
	}

	err = runBatchedFix(state,
		func(afterID int64) ([]*entity.User, int64, error) { return queryCustomAvatarUsers(x, afterID) },
		func(user *entity.User) error { return processAvatarUser(x, &result, opts, state, records, user) })
	if err != nil {
		return fixResult{}, err
	}
	if err := fixNotificationAvatars(x, &result, opts, state); err != nil {
		return fixResult{}, err
	}

	return result, nil
}

func queryCustomAvatarUsers(x *xorm.Engine, afterID int64) ([]*entity.User, int64, error) {
	users := make([]*entity.User, 0)
	err := x.Select("id, avatar").
		Where("id > ?", afterID).
		And("avatar LIKE ?", `%"custom":"%`).
		OrderBy("id").
		Limit(fixBatchSize).
		Find(&users)
	if err != nil {
		return nil, 0, fmt.Errorf("query user failed: %w", err)
	}
	lastID, err := lastStringID(users, func(u *entity.User) string { return u.ID })
	return users, lastID, err
}

func processAvatarUser(x *xorm.Engine, result *fixResult, opts *FixURLPrefixOptions, state *fixRunState, records *fileRecordSync, user *entity.User) error {
	avatar := &schema.AvatarInfo{}
	if err := json.Unmarshal([]byte(user.Avatar), avatar); err != nil {
		fmt.Printf("[%s] WARN user %s: invalid avatar JSON, skipping (manual fix needed): %v\n", fixTypeAvatar, user.ID, err)
		result.Skipped++
		return nil
	}
	// Fixed whatever the current type: the UI keeps custom after switching to
	// another type, and switching back would bring the old URL back.
	if len(avatar.Custom) == 0 {
		return nil
	}
	err := applyValueFix(result, opts, state, fixTypeAvatar, fmt.Sprintf("user %s", user.ID), avatar.Custom,
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
	if err != nil {
		return err
	}
	return records.sync(avatar.Custom)
}

// fixNotificationAvatars fixes the sender avatar URL that a notification
// copies into its content when it is created and shows as is afterwards.
// Notifications not on SRC_PREFIX are left out silently, as there can be many.
func fixNotificationAvatars(x *xorm.Engine, result *fixResult, opts *FixURLPrefixOptions, state *fixRunState) error {
	return runBatchedFix(state,
		func(afterID int64) ([]*entity.Notification, int64, error) {
			notifications := make([]*entity.Notification, 0)
			err := x.Select("id, content").
				Where("id > ?", afterID).
				And("content LIKE ?", `%"avatar":"`+opts.SrcPrefix+`%`).
				OrderBy("id").
				Limit(fixBatchSize).
				Find(&notifications)
			if err != nil {
				return nil, 0, fmt.Errorf("query notification failed: %w", err)
			}
			lastID, err := lastStringID(notifications, func(n *entity.Notification) string { return n.ID })
			return notifications, lastID, err
		},
		func(n *entity.Notification) error {
			target := fmt.Sprintf("notification %s", n.ID)
			content := &schema.NotificationContent{}
			if err := json.Unmarshal([]byte(n.Content), content); err != nil {
				fmt.Printf("[%s] WARN %s: invalid content JSON, skipping (manual fix needed): %v\n", fixTypeAvatar, target, err)
				result.Skipped++
				return nil
			}
			if content.UserInfo == nil || classifyValue(content.UserInfo.Avatar, opts.SrcPrefix, opts.DstPrefix) != matchFix {
				return nil
			}
			return applyValueFix(result, opts, state, fixTypeAvatar, target, content.UserInfo.Avatar,
				func(newURL string) error {
					content.UserInfo.Avatar = newURL
					newJSON, err := json.Marshal(content)
					if err != nil {
						return fmt.Errorf("marshal notification content (id=%s) failed: %w", n.ID, err)
					}
					_, err = x.ID(n.ID).Cols("content").NoAutoTime().Update(&entity.Notification{Content: string(newJSON)})
					if err != nil {
						return fmt.Errorf("update notification content (id=%s) failed: %w", n.ID, err)
					}
					return nil
				})
		})
}
