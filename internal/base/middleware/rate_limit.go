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

package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/repo/limit"
	"github.com/apache/answer/pkg/encryption"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

type RateLimitMiddleware struct {
	limitRepo *limit.LimitRepo
}

// NewRateLimitMiddleware new rate limit middleware
func NewRateLimitMiddleware(limitRepo *limit.LimitRepo) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limitRepo: limitRepo,
	}
}

// DuplicateRequestRejection detects and rejects duplicate requests
// It only works for the requests that post content. Such as add question, add answer, comment etc.
func (rm *RateLimitMiddleware) DuplicateRequestRejection(ctx *gin.Context, req any) (reject bool, key string) {
	userID := GetLoginUserIDFromContext(ctx)
	fullPath := ctx.FullPath()
	reqJson, _ := json.Marshal(req)
	key = encryption.MD5(fmt.Sprintf("%s:%s:%s", userID, fullPath, string(reqJson)))
	var err error
	reject, err = rm.limitRepo.CheckAndRecord(ctx, key)
	if err != nil {
		log.Errorf("check and record rate limit error: %s", err.Error())
		return false, key
	}
	if !reject {
		return false, key
	}
	log.Debugf("duplicate request: [%s] %s", fullPath, string(reqJson))
	handler.HandleResponse(ctx, errors.BadRequest(reason.DuplicateRequestError), nil)
	return true, key
}

// DuplicateRequestClear clear duplicate request record
func (rm *RateLimitMiddleware) DuplicateRequestClear(ctx *gin.Context, key string) {
	err := rm.limitRepo.ClearRecord(ctx, key)
	if err != nil {
		log.Errorf("clear rate limit error: %s", err.Error())
	}
}
