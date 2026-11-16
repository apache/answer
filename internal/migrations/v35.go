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

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func repairAdvancedSiteInfo(ctx context.Context, x *xorm.Engine) error {
	advanced := &entity.SiteInfo{}
	exists, err := x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeAdvanced}).Get(advanced)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	write := &entity.SiteInfo{}
	exists, err = x.Context(ctx).Where(builder.Eq{"type": constant.SiteTypeWrite}).Get(write)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	siteWrite := &schema.SiteWriteResp{}
	if err := json.Unmarshal([]byte(write.Content), siteWrite); err != nil {
		return err
	}
	content, err := json.Marshal(&schema.SiteAdvancedResp{
		MaxImageSize:                   siteWrite.MaxImageSize,
		MaxAttachmentSize:              siteWrite.MaxAttachmentSize,
		MaxImageMegapixel:              siteWrite.MaxImageMegapixel,
		AuthorizedImageExtensions:      siteWrite.AuthorizedImageExtensions,
		AuthorizedAttachmentExtensions: siteWrite.AuthorizedAttachmentExtensions,
	})
	if err != nil {
		return err
	}
	_, err = x.Context(ctx).Insert(&entity.SiteInfo{
		Type:    constant.SiteTypeAdvanced,
		Content: string(content),
		Status:  1,
	})
	return err
}
