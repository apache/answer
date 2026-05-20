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

func aiFeat(ctx context.Context, x *xorm.Engine) error {
	if err := addAIConversationTables(ctx, x); err != nil {
		return fmt.Errorf("add ai conversation tables failed: %w", err)
	}
	if err := addAPIKey(ctx, x); err != nil {
		return fmt.Errorf("add api key failed: %w", err)
	}
	log.Info("AI feature migration completed successfully")
	return nil
}

func addAIConversationTables(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(new(entity.AIConversation)); err != nil {
		return fmt.Errorf("sync ai_conversation table failed: %w", err)
	}

	if err := x.Context(ctx).Sync(new(entity.AIConversationRecord)); err != nil {
		return fmt.Errorf("sync ai_conversation_record table failed: %w", err)
	}

	return nil
}

func addAPIKey(ctx context.Context, x *xorm.Engine) error {
	err := x.Context(ctx).Sync(new(entity.APIKey))
	if err != nil {
		return err
	}

	defaultConfigTable := []*entity.Config{
		{ID: 131, Key: "ai_config.provider", Value: `[{"default_api_host":"https://api.openai.com","display_name":"OpenAI","name":"openai"},{"default_api_host":"https://generativelanguage.googleapis.com","display_name":"Gemini","name":"gemini"},{"default_api_host":"https://api.anthropic.com","display_name":"Anthropic","name":"anthropic"}]`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Context(ctx).Get(&entity.Config{Key: c.Key})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			continue
		}
		if _, err = x.Context(ctx).Insert(&entity.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			log.Errorf("insert %+v config failed: %s", c, err)
			return fmt.Errorf("add config failed: %w", err)
		}
	}

	aiSiteInfo := &entity.SiteInfo{
		Type: constant.SiteTypeAI,
	}
	exist, err := x.Context(ctx).Get(aiSiteInfo)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}
	if exist {
		content := &schema.SiteAIReq{}
		_ = json.Unmarshal([]byte(aiSiteInfo.Content), content)
		content.PromptConfig = &schema.AIPromptConfig{
			ZhCN: constant.DefaultAIPromptConfigZhCN,
			EnUS: constant.DefaultAIPromptConfigEnUS,
		}
		data, _ := json.Marshal(content)
		aiSiteInfo.Content = string(data)
		_, err = x.Context(ctx).ID(aiSiteInfo.ID).Cols("content").Update(aiSiteInfo)
		if err != nil {
			return fmt.Errorf("update site info failed: %w", err)
		}
	} else {
		content := &schema.SiteAIReq{
			PromptConfig: &schema.AIPromptConfig{
				ZhCN: constant.DefaultAIPromptConfigZhCN,
				EnUS: constant.DefaultAIPromptConfigEnUS,
			},
		}
		data, _ := json.Marshal(content)
		aiSiteInfo.Content = string(data)
		aiSiteInfo.Type = constant.SiteTypeAI
		if _, err = x.Context(ctx).Insert(aiSiteInfo); err != nil {
			return fmt.Errorf("insert site info failed: %w", err)
		}
		log.Infof("insert site info %+v", aiSiteInfo)
	}
	return nil
}
