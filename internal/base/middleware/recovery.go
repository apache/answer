/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the License, Version 2.0 (the
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

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

// Recovery returns a middleware that recovers from panics in handlers,
// logs the panic and stack trace, and returns a structured 500 JSON response
// consistent with the rest of the codebase.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic recovered: %v\n%s", r, debug.Stack())
				lang := handler.GetLangByCtx(c)
				c.JSON(http.StatusInternalServerError,
					handler.NewRespBody(http.StatusInternalServerError, reason.UnknownError).TrMsg(lang))
				c.Abort()
			}
		}()
		c.Next()
	}
}