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
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

func Recovery(apiPrefixes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic recovered: %v\n%s", err, debug.Stack())

				// Headers/body already flushed (SSE or any streamed response).
				// We can no longer rewrite the response cleanly; just stop the chain.
				if ctx.Writer.Written() {
					ctx.Abort()
					return
				}

				path := ctx.Request.URL.Path
				for _, p := range apiPrefixes {
					if strings.HasPrefix(path, p) {
						ctx.AbortWithStatusJSON(http.StatusInternalServerError,
							handler.NewRespBody(http.StatusInternalServerError, reason.UnknownError).
								TrMsg(handler.GetLangByCtx(ctx)),
						)
						return
					}
				}

				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}
