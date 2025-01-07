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
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

type ConfigField struct {
	AllowPasswordLogin bool `json:"allow_password_login"`
	// The slug name of plugin that you want to deactivate
	DeactivatePluginSlugName string `json:"deactivate_plugin_slug_name"`
}

// SetDefaultConfig set default config
func SetDefaultConfig(dbConf *data.Database, cacheConf *data.CacheConf, field *ConfigField) error {
	db, err := data.NewDB(false, dbConf)
	if err != nil {
		return err
	}
	defer db.Close()

	cache, cacheCleanup, err := data.NewCache(cacheConf)
	if err != nil {
		fmt.Println("new cache failed")
	}
	defer func() {
		if cache != nil {
			cache.Flush(context.Background())
			cacheCleanup()
		}
	}()

	if field.AllowPasswordLogin {
		return defaultLoginConfig(db)
	}
	if len(field.DeactivatePluginSlugName) > 0 {
		return deactivatePlugin(db, field.DeactivatePluginSlugName)
	}

	return nil
}

func defaultLoginConfig(x *xorm.Engine) (err error) {
	fmt.Println("set default login config")

	loginSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeLogin,
	}
	exist, err := x.Get(loginSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		var content map[string]any
		_ = json.Unmarshal([]byte(loginSiteInfo.Content), &content)
		content["allow_password_login"] = true
		dataByte, _ := json.Marshal(content)
		loginSiteInfo.Content = string(dataByte)
		_, err = x.ID(loginSiteInfo.ID).Cols("content").Update(loginSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	}
	return nil
}

func deactivatePlugin(x *xorm.Engine, pluginSlugName string) (err error) {
	fmt.Printf("try to deactivate plugin: %s\n", pluginSlugName)

	item := &entity.Config{Key: constant.PluginStatus}
	exist, err := x.Get(item)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if !exist {
		return nil
	}

	pluginStatusMapping := make(map[string]bool)
	_ = json.Unmarshal([]byte(item.Value), &pluginStatusMapping)
	status, ok := pluginStatusMapping[pluginSlugName]
	if !ok {
		fmt.Printf("plugin %s not exist\n", pluginSlugName)
		return nil
	}
	if !status {
		fmt.Printf("plugin %s already deactivated\n", pluginSlugName)
		return nil
	}

	pluginStatusMapping[pluginSlugName] = false
	dataByte, _ := json.Marshal(pluginStatusMapping)
	item.Value = string(dataByte)
	_, err = x.ID(item.ID).Cols("value").Update(item)
	if err != nil {
		return fmt.Errorf("update plugin status failed: %w", err)
	}
	return nil
}
