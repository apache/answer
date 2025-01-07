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
	"xorm.io/xorm"
)

func addPasswordLoginControl(ctx context.Context, x *xorm.Engine) error {
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
		content.AllowPasswordLogin = true
		data, _ := json.Marshal(content)
		loginSiteInfo.Content = string(data)
		_, err = x.Context(ctx).ID(loginSiteInfo.ID).Cols("content").Update(loginSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}

	writeSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeWrite,
	}
	exist, err = x.Context(ctx).Get(writeSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		type OldSiteWriteReq struct {
			RestrictAnswer bool     `validate:"omitempty" form:"restrict_answer" json:"restrict_answer"`
			RequiredTag    bool     `validate:"omitempty" form:"required_tag" json:"required_tag"`
			RecommendTags  []string `validate:"omitempty" form:"recommend_tags" json:"recommend_tags"`
			ReservedTags   []string `validate:"omitempty" form:"reserved_tags" json:"reserved_tags"`
			UserID         string   `json:"-"`
		}
		content := &OldSiteWriteReq{}
		_ = json.Unmarshal([]byte(writeSiteInfo.Content), content)
		content.RestrictAnswer = true
		data, _ := json.Marshal(content)
		writeSiteInfo.Content = string(data)
		_, err = x.Context(ctx).ID(writeSiteInfo.ID).Cols("content").Update(writeSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}

	type User struct {
		Avatar string `xorm:"not null default '' VARCHAR(1024) avatar"`
	}
	return x.Context(ctx).Sync(new(User))
}
