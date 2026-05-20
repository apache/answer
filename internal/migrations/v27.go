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
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"xorm.io/xorm"
)

func addSuspendedUntilToUser(ctx context.Context, x *xorm.Engine) error {
	type User struct {
		SuspendedUntil *time.Time `xorm:"DATETIME suspended_until"`
	}
	return x.Context(ctx).Sync(new(User))
}

func moveUserConfigToInterface(ctx context.Context, x *xorm.Engine) error {
	if err := addSuspendedUntilToUser(ctx, x); err != nil {
		return fmt.Errorf("add suspended_until to user failed: %w", err)
	}

	// Get old interface config
	interfaceSiteInfo := &entity.SiteInfo{Type: constant.SiteTypeInterface}
	exist, err := x.Context(ctx).Get(interfaceSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if !exist {
		return fmt.Errorf("interface site info not found")
	}

	interfaceConfig := &schema.SiteInterfaceReq{}
	_ = json.Unmarshal([]byte(interfaceSiteInfo.Content), interfaceConfig)

	// Get old user config
	usersConfig := &entity.SiteInfo{Type: constant.SiteTypeUsers}
	exist, err = x.Context(ctx).Get(usersConfig)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if !exist {
		return fmt.Errorf("users site info not found")
	}

	siteUsers := &schema.SiteUsersReq{}
	_ = json.Unmarshal([]byte(usersConfig.Content), siteUsers)

	interfaceConfig.DefaultAvatar = siteUsers.DefaultAvatar
	interfaceConfig.GravatarBaseURL = siteUsers.GravatarBaseURL

	interfaceConfigByte, _ := json.Marshal(interfaceConfig)
	interfaceSiteInfo.Content = string(interfaceConfigByte)

	_, err = x.Context(ctx).ID(interfaceSiteInfo.ID).Update(interfaceSiteInfo)
	if err != nil {
		return fmt.Errorf("insert site info failed: %w", err)
	}
	return nil
}
