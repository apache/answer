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

package plugin_config

import (
	"context"
	"github.com/apache/answer/internal/base/pager"
	"xorm.io/xorm"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/plugin_common"
	"github.com/segmentfault/pacman/errors"
)

type pluginUserConfigRepo struct {
	data *data.Data
}

// NewPluginUserConfigRepo new repository
func NewPluginUserConfigRepo(data *data.Data) plugin_common.PluginUserConfigRepo {
	return &pluginUserConfigRepo{
		data: data,
	}
}

func (ur *pluginUserConfigRepo) SaveUserPluginConfig(ctx context.Context, userID string,
	pluginSlugName, configValue string) (err error) {
	_, err = ur.data.DB.Transaction(func(session *xorm.Session) (interface{}, error) {
		session = session.Context(ctx)
		old := &entity.PluginUserConfig{
			UserID:         userID,
			PluginSlugName: pluginSlugName,
		}
		exist, err := session.Get(old)
		if err != nil {
			return nil, err
		}
		if exist {
			old.Value = configValue
			_, err = session.ID(old.ID).Update(old)
		} else {
			_, err = session.Insert(&entity.PluginUserConfig{
				UserID:         userID,
				PluginSlugName: pluginSlugName,
				Value:          configValue,
			})
		}
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *pluginUserConfigRepo) GetPluginUserConfig(ctx context.Context, userID, pluginSlugName string) (
	pluginUserConfig *entity.PluginUserConfig, exist bool, err error) {
	pluginUserConfig = &entity.PluginUserConfig{
		UserID:         userID,
		PluginSlugName: pluginSlugName,
	}
	exist, err = ur.data.DB.Context(ctx).Get(pluginUserConfig)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return pluginUserConfig, exist, err
}

func (ur *pluginUserConfigRepo) GetPluginUserConfigPage(ctx context.Context, page, pageSize int) (
	pluginUserConfigs []*entity.PluginUserConfig, total int64, err error) {
	pluginUserConfigs = make([]*entity.PluginUserConfig, 0)
	total, err = pager.Help(page, pageSize, &pluginUserConfigs, &entity.PluginUserConfig{}, ur.data.DB.Context(ctx))
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ur *pluginUserConfigRepo) DeleteUserPluginConfig(ctx context.Context, userID string) (err error) {
	_, err = ur.data.DB.Context(ctx).Where("user_id = ?", userID).Delete(&entity.PluginUserConfig{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
