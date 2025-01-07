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
	"github.com/apache/answer/internal/base/constant"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// VisitAuth when user visit the site image, check visit token. This only for private mode.
func (am *AuthUserMiddleware) VisitAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// If visit brand image, no need to check visit token. Because the brand image is public.
		if strings.HasPrefix(ctx.Request.URL.Path, "/uploads/branding/") {
			ctx.Next()
			return
		}

		siteLogin, err := am.siteInfoCommonService.GetSiteLogin(ctx)
		if err != nil {
			return
		}
		if !siteLogin.LoginRequired {
			ctx.Next()
			return
		}

		visitToken, err := ctx.Cookie(constant.UserVisitCookiesCacheKey)
		if err != nil || len(visitToken) == 0 {
			ctx.Abort()
			ctx.Redirect(http.StatusFound, "/403")
			return
		}

		if !am.authService.CheckUserVisitToken(ctx, visitToken) {
			ctx.Abort()
			ctx.Redirect(http.StatusFound, "/403")
			return
		}
	}
}
