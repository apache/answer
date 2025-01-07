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

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/tidwall/gjson"
	"xorm.io/xorm"
)

func addLoginLimitations(ctx context.Context, x *xorm.Engine) error {
	loginSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeLogin,
	}
	exist, err := x.Context(ctx).Get(loginSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		content := &schema.SiteLoginReq{}
		_ = json.Unmarshal([]byte(loginSiteInfo.Content), content)
		content.AllowEmailRegistrations = true
		content.AllowEmailDomains = make([]string, 0)
		data, _ := json.Marshal(content)
		loginSiteInfo.Content = string(data)
		_, err = x.Context(ctx).ID(loginSiteInfo.ID).Cols("content").Update(loginSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}

	interfaceSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeInterface,
	}
	exist, err = x.Context(ctx).Get(interfaceSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	siteUsers := &schema.SiteUsersReq{
		AllowUpdateDisplayName: true,
		AllowUpdateUsername:    true,
		AllowUpdateAvatar:      true,
		AllowUpdateBio:         true,
		AllowUpdateWebsite:     true,
		AllowUpdateLocation:    true,
	}
	if exist {
		siteUsers.DefaultAvatar = gjson.Get(interfaceSiteInfo.Content, "default_avatar").String()
	}
	data, _ := json.Marshal(siteUsers)

	exist, err = x.Context(ctx).Get(&entity.SiteInfo{Type: constant.SiteTypeUsers})
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if !exist {
		usersSiteInfo := &entity.SiteInfo{
			Type:    constant.SiteTypeUsers,
			Content: string(data),
			Status:  1,
		}
		_, err = x.Context(ctx).Insert(usersSiteInfo)
		if err != nil {
			return fmt.Errorf("insert site info failed: %w", err)
		}
	}
	return nil
}
