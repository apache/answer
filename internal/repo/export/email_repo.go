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

package export

import (
	"context"
	"github.com/apache/answer/internal/base/constant"
	"github.com/tidwall/gjson"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/service/export"
	"github.com/segmentfault/pacman/errors"
)

// emailRepo email repository
type emailRepo struct {
	data *data.Data
}

// NewEmailRepo new repository
func NewEmailRepo(data *data.Data) export.EmailRepo {
	return &emailRepo{
		data: data,
	}
}

// SetCode The email code is used to verify that the link in the message is out of date
func (e *emailRepo) SetCode(ctx context.Context, userID, code, content string, duration time.Duration) error {
	// Setting the latest code is to help ensure that only one link is active at a time.
	// Set userID -> latest code
	if err := e.data.Cache.SetString(ctx, constant.UserLatestEmailCodeCacheKey+userID, code, duration); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	// Set latest code -> content
	if err := e.data.Cache.SetString(ctx, constant.UserEmailCodeCacheKey+code, content, duration); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// VerifyCode verify the code if out of date
func (e *emailRepo) VerifyCode(ctx context.Context, code string) (content string, err error) {
	// Get latest code -> content
	codeCacheKey := constant.UserEmailCodeCacheKey + code
	content, exist, err := e.data.Cache.GetString(ctx, codeCacheKey)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", nil
	}

	// Delete the code after verification
	_ = e.data.Cache.Del(ctx, codeCacheKey)

	// If some email content does not need to verify the latest code is the same as the code, skip it.
	// For example, some unsubscribe email content does not need to verify the latest code.
	// This link always works before the code is out of date.
	if skipValidationLatestCode := gjson.Get(content, "skip_validation_latest_code").Bool(); skipValidationLatestCode {
		return content, nil
	}
	userID := gjson.Get(content, "user_id").String()

	// Get userID -> latest code
	latestCode, exist, err := e.data.Cache.GetString(ctx, constant.UserLatestEmailCodeCacheKey+userID)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", nil
	}

	// Check if the latest code is the same as the code, if not, means the code is out of date
	if latestCode != code {
		return "", nil
	}
	return content, nil
}
