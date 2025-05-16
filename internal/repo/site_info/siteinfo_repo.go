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

package site_info

import (
	"context"
	"encoding/json"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
)

type siteInfoRepo struct {
	data *data.Data
}

func NewSiteInfo(data *data.Data) siteinfo_common.SiteInfoRepo {
	return &siteInfoRepo{
		data: data,
	}
}

// SaveByType save site setting by type
func (sr *siteInfoRepo) SaveByType(ctx context.Context, siteType string, data *entity.SiteInfo) (err error) {
	old := &entity.SiteInfo{}
	exist, err := sr.data.DB.Context(ctx).Where(builder.Eq{"type": siteType}).Get(old)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		_, err = sr.data.DB.Context(ctx).ID(old.ID).Update(data)
	} else {
		_, err = sr.data.DB.Context(ctx).Insert(data)
	}
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	sr.setCache(ctx, siteType, data)
	return
}

// GetByType get site info by type
func (sr *siteInfoRepo) GetByType(ctx context.Context, siteType string) (siteInfo *entity.SiteInfo, exist bool, err error) {
	siteInfo = sr.getCache(ctx, siteType)
	if siteInfo != nil {
		return siteInfo, true, nil
	}
	siteInfo = &entity.SiteInfo{}
	exist, err = sr.data.DB.Context(ctx).Where(builder.Eq{"type": siteType}).Get(siteInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return nil, false, err
	}
	if exist {
		sr.setCache(ctx, siteType, siteInfo)
	}
	return
}

func (sr *siteInfoRepo) getCache(ctx context.Context, siteType string) (siteInfo *entity.SiteInfo) {
	siteInfoCache, exist, err := sr.data.Cache.GetString(ctx, constant.SiteInfoCacheKey+siteType)
	if err != nil {
		return nil
	}
	if !exist {
		return nil
	}
	siteInfo = &entity.SiteInfo{}
	_ = json.Unmarshal([]byte(siteInfoCache), siteInfo)
	return siteInfo
}

func (sr *siteInfoRepo) setCache(ctx context.Context, siteType string, siteInfo *entity.SiteInfo) {
	siteInfoCache, _ := json.Marshal(siteInfo)
	err := sr.data.Cache.SetString(ctx,
		constant.SiteInfoCacheKey+siteType, string(siteInfoCache), constant.SiteInfoCacheTime)
	if err != nil {
		log.Error(err)
	}
}

func (sr *siteInfoRepo) IsBrandingFileUsed(ctx context.Context, filePath string) (bool, error) {
	siteInfo := &entity.SiteInfo{}
	count, err := sr.data.DB.Context(ctx).
		Table("site_info").
		Where(builder.Eq{"type": "branding"}).
		And(builder.Like{"content", "%" + filePath + "%"}).
		Count(&siteInfo)

	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	return count > 0, nil
}
