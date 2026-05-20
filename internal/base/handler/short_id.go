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

package handler

import (
	"context"

	"github.com/apache/answer/internal/base/constant"
	"github.com/gin-gonic/gin"
)

// GetEnableShortID get short id flag from context
func GetEnableShortID(ctx context.Context) bool {
	// Check gin context first (set by ShortIDMiddleware via ctx.Set)
	if ginCtx, ok := ctx.(*gin.Context); ok {
		flag, ok := ginCtx.Get(constant.ShortIDFlag)
		if ok {
			if flag, ok := flag.(bool); ok {
				return flag
			}
			return false
		}
	}
	// Fallback for non-gin contexts (e.g., SitemapCron uses context.WithValue)
	flag, ok := ctx.Value(constant.ShortIDContextKey).(bool)
	if ok {
		return flag
	}
	return false
}
