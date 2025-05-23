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

package plugin_common

import (
	"context"
	"encoding/json"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/repo/search_sync"

	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/importer"
	"github.com/apache/answer/plugin"
)

type PluginConfigRepo interface {
	SavePluginConfig(ctx context.Context, pluginSlugName, configValue string) (err error)
	GetPluginConfigAll(ctx context.Context) (pluginConfigs []*entity.PluginConfig, err error)
}

type PluginUserConfigRepo interface {
	SaveUserPluginConfig(ctx context.Context, userID string, pluginSlugName, configValue string) (err error)
	GetPluginUserConfig(ctx context.Context, userID, pluginSlugName string) (
		pluginUserConfig *entity.PluginUserConfig, exist bool, err error)
	GetPluginUserConfigPage(ctx context.Context, page, pageSize int) (
		pluginUserConfigs []*entity.PluginUserConfig, total int64, err error)
	DeleteUserPluginConfig(ctx context.Context, userID string) (err error)
}

// PluginCommonService user service
type PluginCommonService struct {
	configService        *config.ConfigService
	pluginConfigRepo     PluginConfigRepo
	pluginUserConfigRepo PluginUserConfigRepo
	data                 *data.Data
	importerService      *importer.ImporterService
}

// NewPluginCommonService new report service
func NewPluginCommonService(
	pluginConfigRepo PluginConfigRepo,
	pluginUserConfigRepo PluginUserConfigRepo,
	configService *config.ConfigService,
	data *data.Data,
	importerService *importer.ImporterService,
) *PluginCommonService {

	p := &PluginCommonService{
		configService:        configService,
		pluginConfigRepo:     pluginConfigRepo,
		pluginUserConfigRepo: pluginUserConfigRepo,
		data:                 data,
		importerService:      importerService,
	}
	p.initPluginData()
	return p
}

// UpdatePluginStatus update plugin status
func (ps *PluginCommonService) UpdatePluginStatus(ctx context.Context) (err error) {
	content, err := plugin.StatusManager.MarshalJSON()
	if err != nil {
		return errors.InternalServer(reason.UnknownError).WithError(err)
	}
	return ps.configService.UpdateConfig(ctx, constant.PluginStatus, string(content))
}

// UpdatePluginConfig update plugin config
func (ps *PluginCommonService) UpdatePluginConfig(ctx context.Context, req *schema.UpdatePluginConfigReq) (err error) {
	configValue, _ := json.Marshal(req.ConfigFields)
	err = ps.pluginConfigRepo.SavePluginConfig(ctx, req.PluginSlugName, string(configValue))
	if err != nil {
		return err
	}

	_ = plugin.CallSearch(func(search plugin.Search) error {
		if search.Info().SlugName == req.PluginSlugName {
			search.RegisterSyncer(ctx, search_sync.NewPluginSyncer(ps.data))
		}
		return nil
	})
	_ = plugin.CallImporter(func(importer plugin.Importer) error {
		importer.RegisterImporterFunc(ctx, ps.importerService.NewImporterFunc())
		return nil
	})
	return nil
}

// UpdatePluginUserConfig update plugin config
func (ps *PluginCommonService) UpdatePluginUserConfig(ctx context.Context, req *schema.UpdateUserPluginConfigReq) (err error) {
	configValue, _ := json.Marshal(req.ConfigFields)
	err = ps.pluginUserConfigRepo.SaveUserPluginConfig(ctx, req.UserID, req.PluginSlugName, string(configValue))
	if err != nil {
		return err
	}
	return nil
}

// GetUserPluginConfig get user plugin config
func (ps *PluginCommonService) GetUserPluginConfig(ctx context.Context, req *schema.GetUserPluginConfigReq) (
	configValue string, err error) {
	pluginUserConfig, exist, err := ps.pluginUserConfigRepo.GetPluginUserConfig(ctx, req.UserID, req.PluginSlugName)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", nil
	}
	return pluginUserConfig.Value, nil
}

func (ps *PluginCommonService) initPluginData() {
	_ = plugin.CallKVStorage(func(k plugin.KVStorage) error {
		k.SetOperator(plugin.NewKVOperator(
			ps.data.DB,
			ps.data.Cache,
			k.Info().SlugName,
		))
		return nil
	})

	// init plugin status
	pluginStatus, err := ps.configService.GetStringValue(context.TODO(), constant.PluginStatus)
	if err != nil {
		log.Error(err)
	} else {
		if err := plugin.StatusManager.UnmarshalJSON([]byte(pluginStatus)); err != nil {
			log.Error(err)
		}
	}

	// init plugin config
	pluginConfigs, err := ps.pluginConfigRepo.GetPluginConfigAll(context.Background())
	if err != nil {
		log.Error(err)
	} else {
		for _, pluginConfig := range pluginConfigs {
			err := plugin.CallConfig(func(fn plugin.Config) error {
				if fn.Info().SlugName == pluginConfig.PluginSlugName {
					return fn.ConfigReceiver([]byte(pluginConfig.Value))
				}
				return nil
			})
			if err != nil {
				log.Errorf("parse plugin config failed: %s %v", pluginConfig.PluginSlugName, err)
			}
		}

		_ = plugin.CallCache(func(cache plugin.Cache) error {
			ps.data.Cache = cache
			return nil
		})
	}

	// init plugin user config
	plugin.RegisterGetPluginUserConfigFunc(func(userID, pluginSlugName string) []byte {
		pluginUserConfig, exist, err := ps.pluginUserConfigRepo.GetPluginUserConfig(context.Background(), userID, pluginSlugName)
		if err != nil {
			log.Error(err)
			return nil
		}
		if !exist {
			return nil
		}
		return []byte(pluginUserConfig.Value)
	})

	// init plugin user config data
	go func() {
		page, pageSize := 1, 1000
		for {
			userConfigs, _, err := ps.pluginUserConfigRepo.GetPluginUserConfigPage(context.Background(), page, pageSize)
			if err != nil {
				log.Error(err)
				return
			}
			if len(userConfigs) == 0 {
				return
			}
			for _, userConfig := range userConfigs {
				err := plugin.CallUserConfig(func(fn plugin.UserConfig) error {
					if fn.Info().SlugName == userConfig.PluginSlugName {
						return fn.UserConfigReceiver(userConfig.UserID, []byte(userConfig.Value))
					}
					return nil
				})
				if err != nil {
					log.Errorf("parse plugin user config failed: %s %v", userConfig.PluginSlugName, err)
				}
			}
			page++
		}
	}()
}
