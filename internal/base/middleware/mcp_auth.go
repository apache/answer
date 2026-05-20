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
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// AuthMcpEnable check mcp is enabled
func (am *AuthUserMiddleware) AuthMcpEnable() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		mcpConfig, err := am.siteInfoCommonService.GetSiteMCP(ctx)
		if err != nil {
			handler.HandleResponse(ctx, errors.InternalServer(reason.UnknownError), nil)
			ctx.Abort()
			return
		}
		if mcpConfig != nil && mcpConfig.Enabled {
			ctx.Next()
			return
		}
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		ctx.Abort()
		log.Error("abort mcp auth middleware, get mcp config error: ", err)
	}
}
