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

package config

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/config"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// configRepo config repository
type configRepo struct {
	data *data.Data
}

// NewConfigRepo new repository
func NewConfigRepo(data *data.Data) config.ConfigRepo {
	repo := &configRepo{
		data: data,
	}
	return repo
}

func (cr configRepo) GetConfigByID(ctx context.Context, id int) (c *entity.Config, err error) {
	cacheKey := fmt.Sprintf("%s%d", constant.ConfigID2KEYCacheKeyPrefix, id)
	cacheData, exist, err := cr.data.Cache.GetString(ctx, cacheKey)
	if err == nil && exist && len(cacheData) > 0 {
		c = &entity.Config{}
		c.BuildByJSON([]byte(cacheData))
		if c.ID > 0 {
			return c, nil
		}
	}

	c = &entity.Config{}
	exist, err = cr.data.DB.Context(ctx).ID(id).Get(c)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil, fmt.Errorf("config not found by id: %d", id)
	}

	// update cache
	if err := cr.data.Cache.SetString(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime); err != nil {
		log.Error(err)
	}
	return c, nil
}

func (cr configRepo) GetConfigByKey(ctx context.Context, key string) (c *entity.Config, err error) {
	cacheKey := constant.ConfigKEY2ContentCacheKeyPrefix + key
	cacheData, exist, err := cr.data.Cache.GetString(ctx, cacheKey)
	if err == nil && exist && len(cacheData) > 0 {
		c = &entity.Config{}
		c.BuildByJSON([]byte(cacheData))
		if c.ID > 0 {
			return c, nil
		}
	}

	c = &entity.Config{Key: key}
	exist, err = cr.data.DB.Context(ctx).Get(c)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil, fmt.Errorf("config not found by key: %s", key)
	}

	// update cache
	if err := cr.data.Cache.SetString(ctx, cacheKey, c.JsonString(), constant.ConfigCacheTime); err != nil {
		log.Error(err)
	}
	return c, nil
}

func (cr configRepo) UpdateConfig(ctx context.Context, key string, value string) (err error) {
	// check if key exists
	oldConfig := &entity.Config{Key: key}
	exist, err := cr.data.DB.Context(ctx).Get(oldConfig)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return errors.BadRequest(reason.ObjectNotFound)
	}

	// update database
	_, err = cr.data.DB.Context(ctx).ID(oldConfig.ID).Update(&entity.Config{Value: value})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	oldConfig.Value = value
	cacheVal := oldConfig.JsonString()
	// update cache
	if err := cr.data.Cache.SetString(ctx,
		constant.ConfigKEY2ContentCacheKeyPrefix+key, cacheVal, constant.ConfigCacheTime); err != nil {
		log.Error(err)
	}
	if err := cr.data.Cache.SetString(ctx,
		fmt.Sprintf("%s%d", constant.ConfigID2KEYCacheKeyPrefix, oldConfig.ID), cacheVal, constant.ConfigCacheTime); err != nil {
		log.Error(err)
	}
	return
}
