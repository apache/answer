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

func addQuestionLinkedCount(ctx context.Context, x *xorm.Engine) error {
	writeSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeWrite,
	}
	exist, err := x.Context(ctx).Get(writeSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		type OldSiteWriteReq struct {
			RestrictAnswer                 bool                   `json:"restrict_answer"`
			RequiredTag                    bool                   `json:"required_tag"`
			RecommendTags                  []*schema.SiteWriteTag `json:"recommend_tags"`
			ReservedTags                   []*schema.SiteWriteTag `json:"reserved_tags"`
			MaxImageSize                   int                    `json:"max_image_size"`
			MaxAttachmentSize              int                    `json:"max_attachment_size"`
			MaxImageMegapixel              int                    `json:"max_image_megapixel"`
			AuthorizedImageExtensions      []string               `json:"authorized_image_extensions"`
			AuthorizedAttachmentExtensions []string               `json:"authorized_attachment_extensions"`
		}
		content := &OldSiteWriteReq{}
		_ = json.Unmarshal([]byte(writeSiteInfo.Content), content)
		content.MaxImageSize = 4
		content.MaxAttachmentSize = 8
		content.MaxImageMegapixel = 40
		content.AuthorizedImageExtensions = []string{"jpg", "jpeg", "png", "gif", "webp"}
		content.AuthorizedAttachmentExtensions = []string{}
		data, _ := json.Marshal(content)
		writeSiteInfo.Content = string(data)
		_, err = x.Context(ctx).ID(writeSiteInfo.ID).Cols("content").Update(writeSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}

	type Question struct {
		LinkedCount int `xorm:"not null default 0 INT(11) linked_count"`
	}
	return x.Context(ctx).Sync(new(entity.Question))
}
