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

package feature_toggle

import (
	"context"
	"encoding/json"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/segmentfault/pacman/errors"
)

// Feature keys
const (
	FeatureBadge        = "badge"
	FeatureCustomDomain = "custom_domain"
	FeatureMCP          = "mcp"
	FeaturePrivateAPI   = "private_api"
	FeatureAIChatbot    = "ai_chatbot"
	FeatureArticle      = "article"
	FeatureCategory     = "category"
)

type toggleConfig struct {
	Toggles map[string]bool `json:"toggles"`
}

// FeatureToggleService persist and query feature switches.
type FeatureToggleService struct {
	siteInfoRepo siteinfo_common.SiteInfoRepo
}

// NewFeatureToggleService creates a new feature toggle service instance.
func NewFeatureToggleService(siteInfoRepo siteinfo_common.SiteInfoRepo) *FeatureToggleService {
	return &FeatureToggleService{
		siteInfoRepo: siteInfoRepo,
	}
}

// UpdateAll overwrites the feature toggle configuration.
func (s *FeatureToggleService) UpdateAll(ctx context.Context, toggles map[string]bool) error {
	cfg := &toggleConfig{
		Toggles: sanitizeToggleMap(toggles),
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	info := &entity.SiteInfo{
		Type:    constant.SiteTypeFeatureToggle,
		Content: string(data),
		Status:  1,
	}

	return s.siteInfoRepo.SaveByType(ctx, constant.SiteTypeFeatureToggle, info)
}

// GetAll returns all feature toggles.
func (s *FeatureToggleService) GetAll(ctx context.Context) (map[string]bool, error) {
	siteInfo, exist, err := s.siteInfoRepo.GetByType(ctx, constant.SiteTypeFeatureToggle, true)
	if err != nil {
		return nil, err
	}
	if !exist || siteInfo == nil || siteInfo.Content == "" {
		return map[string]bool{}, nil
	}

	cfg := &toggleConfig{}
	if err := json.Unmarshal([]byte(siteInfo.Content), cfg); err != nil {
		return map[string]bool{}, err
	}
	return sanitizeToggleMap(cfg.Toggles), nil
}

// IsEnabled returns whether a feature is enabled. Missing config defaults to true.
func (s *FeatureToggleService) IsEnabled(ctx context.Context, feature string) (bool, error) {
	toggles, err := s.GetAll(ctx)
	if err != nil {
		return false, err
	}
	if len(toggles) == 0 {
		return true, nil
	}
	value, ok := toggles[feature]
	if !ok {
		return true, nil
	}
	return value, nil
}

// EnsureEnabled returns error if feature disabled.
func (s *FeatureToggleService) EnsureEnabled(ctx context.Context, feature string) error {
	enabled, err := s.IsEnabled(ctx, feature)
	if err != nil {
		return err
	}
	if !enabled {
		return errors.BadRequest(reason.ErrFeatureDisabled)
	}
	return nil
}

func sanitizeToggleMap(in map[string]bool) map[string]bool {
	if in == nil {
		return map[string]bool{}
	}
	return in
}
