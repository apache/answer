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

package limit

import (
	"context"
	"fmt"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/segmentfault/pacman/errors"
	"time"
)

// LimitRepo auth repository
type LimitRepo struct {
	data *data.Data
}

// NewRateLimitRepo new repository
func NewRateLimitRepo(data *data.Data) *LimitRepo {
	return &LimitRepo{
		data: data,
	}
}

// CheckAndRecord check
func (lr *LimitRepo) CheckAndRecord(ctx context.Context, key string) (limit bool, err error) {
	_, exist, err := lr.data.Cache.GetString(ctx, constant.RateLimitCacheKeyPrefix+key)
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if exist {
		return true, nil
	}
	err = lr.data.Cache.SetString(ctx, constant.RateLimitCacheKeyPrefix+key,
		fmt.Sprintf("%d", time.Now().Unix()), constant.RateLimitCacheTime)
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return false, nil
}

// ClearRecord clear
func (lr *LimitRepo) ClearRecord(ctx context.Context, key string) error {
	return lr.data.Cache.Del(ctx, constant.RateLimitCacheKeyPrefix+key)
}
