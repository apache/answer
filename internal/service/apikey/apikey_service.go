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

package apikey

import (
	"context"
	"strings"
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/token"
)

type APIKeyRepo interface {
	GetAPIKeyList(ctx context.Context) (keys []*entity.APIKey, err error)
	GetAPIKey(ctx context.Context, apiKey string) (key *entity.APIKey, exist bool, err error)
	UpdateAPIKey(ctx context.Context, apiKey entity.APIKey) (err error)
	AddAPIKey(ctx context.Context, apiKey entity.APIKey) (err error)
	DeleteAPIKey(ctx context.Context, id int) (err error)
}

type APIKeyService struct {
	apiKeyRepo APIKeyRepo
}

func NewAPIKeyService(
	apiKeyRepo APIKeyRepo,
) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

func (s *APIKeyService) GetAPIKeyList(ctx context.Context, req *schema.GetAPIKeyReq) (resp []*schema.GetAPIKeyResp, err error) {
	keys, err := s.apiKeyRepo.GetAPIKeyList(ctx)
	if err != nil {
		return nil, err
	}
	resp = make([]*schema.GetAPIKeyResp, 0)
	for _, key := range keys {
		// hide access key middle part, replace with *
		if len(key.AccessKey) < 10 {
			// If the access key is too short, do not mask it
			key.AccessKey = strings.Repeat("*", len(key.AccessKey))
		} else {
			key.AccessKey = key.AccessKey[:7] + strings.Repeat("*", 8) + key.AccessKey[len(key.AccessKey)-4:]
		}

		resp = append(resp, &schema.GetAPIKeyResp{
			ID:          key.ID,
			AccessKey:   key.AccessKey,
			Description: key.Description,
			Scope:       key.Scope,
			CreatedAt:   key.CreatedAt.Unix(),
			LastUsedAt:  key.LastUsedAt.Unix(),
		})
	}
	return resp, nil
}

func (s *APIKeyService) UpdateAPIKey(ctx context.Context, req *schema.UpdateAPIKeyReq) (err error) {
	apiKey := entity.APIKey{
		ID:          req.ID,
		Description: req.Description,
	}
	err = s.apiKeyRepo.UpdateAPIKey(ctx, apiKey)
	if err != nil {
		return err
	}
	return nil
}

func (s *APIKeyService) AddAPIKey(ctx context.Context, req *schema.AddAPIKeyReq) (resp *schema.AddAPIKeyResp, err error) {
	ak := "sk_" + strings.ReplaceAll(token.GenerateToken(), "-", "")
	apiKey := entity.APIKey{
		Description: req.Description,
		AccessKey:   ak,
		Scope:       req.Scope,
		LastUsedAt:  time.Now(),
		UserID:      req.UserID,
	}
	err = s.apiKeyRepo.AddAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	resp = &schema.AddAPIKeyResp{
		AccessKey: apiKey.AccessKey,
	}
	return resp, nil
}

func (s *APIKeyService) DeleteAPIKey(ctx context.Context, req *schema.DeleteAPIKeyReq) (err error) {
	err = s.apiKeyRepo.DeleteAPIKey(ctx, req.ID)
	if err != nil {
		return err
	}
	return nil
}
