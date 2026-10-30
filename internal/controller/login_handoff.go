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

package controller

import (
	"fmt"
	"net/http"

	"github.com/apache/answer/internal/service/auth"
	"github.com/gin-gonic/gin"
)

func redirectToAuthLanding(ctx *gin.Context, siteURL string, loginHandoffService *auth.LoginHandoffService, accessToken string) error {
	code, err := loginHandoffService.Create(ctx, accessToken)
	if err != nil {
		return err
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/users/auth-landing#code=%s", siteURL, code))
	return nil
}
