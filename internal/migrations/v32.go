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
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func addDeepSeekProvider(ctx context.Context, x *xorm.Engine) error {
	config := &entity.Config{Key: constant.AIConfigProvider}
	exist, err := x.Context(ctx).Get(config)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}

	defaultProviders := []*schema.GetAIProviderResp{
		{Name: "openai", DisplayName: "OpenAI", DefaultAPIHost: "https://api.openai.com"},
		{Name: "gemini", DisplayName: "Gemini", DefaultAPIHost: "https://generativelanguage.googleapis.com"},
		{Name: "anthropic", DisplayName: "Anthropic", DefaultAPIHost: "https://api.anthropic.com"},
		{Name: "deepseek", DisplayName: "DeepSeek", DefaultAPIHost: "https://api.deepseek.com"},
	}

	if !exist || config.Value == "" {
		value, err := json.Marshal(defaultProviders)
		if err != nil {
			return fmt.Errorf("marshal providers failed: %w", err)
		}
		if !exist {
			if _, err = x.Context(ctx).Insert(&entity.Config{
				ID:    131,
				Key:   constant.AIConfigProvider,
				Value: string(value),
			}); err != nil {
				log.Errorf("insert config failed: %s", err)
				return fmt.Errorf("add config failed: %w", err)
			}
			return nil
		}
		_, err = x.Context(ctx).Where("key = ?", constant.AIConfigProvider).Cols("value").Update(&entity.Config{Value: string(value)})
		if err != nil {
			return fmt.Errorf("update config failed: %w", err)
		}
		return nil
	}

	providers := make([]*schema.GetAIProviderResp, 0)
	if err := json.Unmarshal([]byte(config.Value), &providers); err != nil {
		value, marshalErr := json.Marshal(defaultProviders)
		if marshalErr != nil {
			return fmt.Errorf("marshal providers failed: %w", marshalErr)
		}
		_, err = x.Context(ctx).Where("key = ?", constant.AIConfigProvider).Cols("value").Update(&entity.Config{Value: string(value)})
		if err != nil {
			return fmt.Errorf("update config failed: %w", err)
		}
		return nil
	}

	for _, provider := range providers {
		if provider.Name == "deepseek" {
			return nil
		}
	}
	providers = append(providers, &schema.GetAIProviderResp{
		Name:           "deepseek",
		DisplayName:    "DeepSeek",
		DefaultAPIHost: "https://api.deepseek.com",
	})

	value, err := json.Marshal(providers)
	if err != nil {
		return fmt.Errorf("marshal providers failed: %w", err)
	}
	_, err = x.Context(ctx).Where("key = ?", constant.AIConfigProvider).Cols("value").Update(&entity.Config{Value: string(value)})
	if err != nil {
		return fmt.Errorf("update config failed: %w", err)
	}
	return nil
}
