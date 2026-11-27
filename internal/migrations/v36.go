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

// legacyGeminiAPIHost is the Gemini default shipped by the v1.8.1 AI
// migration. It normalizes to the invalid "/v1" path; Gemini's
// OpenAI-compatible endpoint lives under "/v1beta/openai". Only the
// unchanged legacy default is upgraded, so administrator-customized hosts
// are never overwritten.
const (
	legacyGeminiAPIHost = "https://generativelanguage.googleapis.com"
	geminiAPIHost       = "https://generativelanguage.googleapis.com/v1beta/openai"
)

func fixGeminiDefaultAPIHost(ctx context.Context, x *xorm.Engine) error {
	if err := fixGeminiProviderConfig(ctx, x); err != nil {
		return fmt.Errorf("fix gemini provider config failed: %w", err)
	}
	if err := fixGeminiSiteAIConfig(ctx, x); err != nil {
		return fmt.Errorf("fix gemini site ai config failed: %w", err)
	}
	log.Info("gemini default api host migration completed successfully")
	return nil
}

// fixGeminiProviderConfig upgrades the unchanged legacy Gemini default in the
// "ai_config.provider" config row (the provider list behind the admin UI).
func fixGeminiProviderConfig(ctx context.Context, x *xorm.Engine) error {
	cfg := &entity.Config{Key: "ai_config.provider"}
	exist, err := x.Context(ctx).Get(cfg)
	if err != nil {
		return fmt.Errorf("get ai_config.provider config failed: %w", err)
	}
	if !exist {
		return nil
	}
	providers := []map[string]any{}
	if err = json.Unmarshal([]byte(cfg.Value), &providers); err != nil {
		// Unrecognized content is left untouched.
		log.Warnf("ai_config.provider config is not a provider list, skip: %s", err)
		return nil
	}
	if !fixGeminiProviderList(providers) {
		return nil
	}
	data, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	cfg.Value = string(data)
	_, err = x.Context(ctx).ID(cfg.ID).Cols("value").Update(cfg)
	return err
}

// fixGeminiProviderList upgrades the unchanged legacy Gemini default in a
// provider list, in place. Reports whether anything changed.
func fixGeminiProviderList(providers []map[string]any) bool {
	changed := false
	for _, p := range providers {
		if name, _ := p["name"].(string); name != "gemini" {
			continue
		}
		if host, _ := p["default_api_host"].(string); host == legacyGeminiAPIHost {
			p["default_api_host"] = geminiAPIHost
			changed = true
		}
	}
	return changed
}

// fixGeminiSiteAIConfig upgrades the unchanged legacy Gemini host in the saved
// site AI configuration (SaveSiteAI backfills every provider entry with the
// provider-list default at save time, so existing installs carry the stale
// host in their ai_providers entries).
func fixGeminiSiteAIConfig(ctx context.Context, x *xorm.Engine) error {
	siteInfo := &entity.SiteInfo{Type: constant.SiteTypeAI}
	exist, err := x.Context(ctx).Where("type = ?", constant.SiteTypeAI).Get(siteInfo)
	if err != nil {
		return fmt.Errorf("get site ai info failed: %w", err)
	}
	if !exist {
		return nil
	}
	content, changed, err := fixGeminiSiteAIContent(siteInfo.Content)
	if err != nil {
		// Unrecognized content is left untouched.
		log.Warnf("site ai info content is not valid, skip: %s", err)
		return nil
	}
	if !changed {
		return nil
	}
	siteInfo.Content = content
	_, err = x.Context(ctx).ID(siteInfo.ID).Cols("content").Update(siteInfo)
	return err
}

// fixGeminiSiteAIContent upgrades unchanged legacy Gemini hosts in the saved
// site AI configuration JSON. Returns the new content and whether it changed.
func fixGeminiSiteAIContent(content string) (string, bool, error) {
	aiConfig := &schema.SiteAIReq{}
	if err := json.Unmarshal([]byte(content), aiConfig); err != nil {
		return content, false, err
	}
	changed := false
	for _, provider := range aiConfig.SiteAIProviders {
		if provider.Provider == "gemini" && provider.APIHost == legacyGeminiAPIHost {
			provider.APIHost = geminiAPIHost
			changed = true
		}
	}
	if !changed {
		return content, false, nil
	}
	data, err := json.Marshal(aiConfig)
	if err != nil {
		return content, false, err
	}
	return string(data), true, nil
}
