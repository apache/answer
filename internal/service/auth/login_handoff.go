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

package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/segmentfault/pacman/errors"
)

type atomicStringConsumer interface {
	ConsumeString(ctx context.Context, key string) (data string, exist bool, err error)
}

// LoginHandoffService exchanges a short-lived, one-time code for an access token.
// Cache backends that implement atomicStringConsumer provide cross-process
// single-use semantics. The mutex protects the built-in in-memory cache.
type LoginHandoffService struct {
	data *data.Data
	mu   sync.Mutex
}

func NewLoginHandoffService(data *data.Data) *LoginHandoffService {
	return &LoginHandoffService{data: data}
}

// Create stores an access token behind an opaque short-lived code.
func (s *LoginHandoffService) Create(ctx context.Context, accessToken string) (string, error) {
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return "", errors.InternalServer(reason.UnknownError).WithError(err).WithStack()
	}
	code := base64.RawURLEncoding.EncodeToString(codeBytes)
	if err := s.data.Cache.SetString(ctx, constant.LoginHandoffCacheKey+code, accessToken, constant.LoginHandoffCacheTime); err != nil {
		return "", errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return code, nil
}

// Consume returns the token for code once. Expired, unknown, and already-used
// codes all return exists=false without disclosing which condition occurred.
func (s *LoginHandoffService) Consume(ctx context.Context, code string) (accessToken string, exists bool, err error) {
	if len(code) == 0 {
		return "", false, nil
	}

	key := constant.LoginHandoffCacheKey + code
	if consumer, ok := s.data.Cache.(atomicStringConsumer); ok {
		accessToken, exists, err = consumer.ConsumeString(ctx, key)
	} else {
		s.mu.Lock()
		defer s.mu.Unlock()
		accessToken, exists, err = s.data.Cache.GetString(ctx, key)
		if err == nil && exists {
			err = s.data.Cache.Del(ctx, key)
		}
	}
	if err != nil {
		return "", false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exists {
		return "", false, nil
	}
	return accessToken, true, nil
}
