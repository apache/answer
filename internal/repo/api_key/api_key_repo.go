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

package api_key

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/apikey"
	"github.com/segmentfault/pacman/errors"
)

type apiKeyRepo struct {
	data *data.Data
}

// NewAPIKeyRepo creates a new apiKey repository
func NewAPIKeyRepo(data *data.Data) apikey.APIKeyRepo {
	return &apiKeyRepo{
		data: data,
	}
}

func (ar *apiKeyRepo) GetAPIKeyList(ctx context.Context) (keys []*entity.APIKey, err error) {
	keys = make([]*entity.APIKey, 0)
	err = ar.data.DB.Context(ctx).Where("hidden = ?", 0).Find(&keys)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ar *apiKeyRepo) GetAPIKey(ctx context.Context, apiKey string) (key *entity.APIKey, exist bool, err error) {
	key = &entity.APIKey{}
	exist, err = ar.data.DB.Context(ctx).Where("access_key = ?", apiKey).Get(key)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ar *apiKeyRepo) UpdateAPIKey(ctx context.Context, apiKey entity.APIKey) (err error) {
	_, err = ar.data.DB.Context(ctx).ID(apiKey.ID).Update(&apiKey)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ar *apiKeyRepo) AddAPIKey(ctx context.Context, apiKey entity.APIKey) (err error) {
	_, err = ar.data.DB.Context(ctx).Insert(&apiKey)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ar *apiKeyRepo) DeleteAPIKey(ctx context.Context, id int) (err error) {
	_, err = ar.data.DB.Context(ctx).ID(id).Delete(&entity.APIKey{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
