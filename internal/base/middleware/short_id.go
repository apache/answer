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
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/log"
)

type ShortIDMiddleware struct {
	siteInfoService siteinfo_common.SiteInfoCommonService
}

func NewShortIDMiddleware(siteInfoService siteinfo_common.SiteInfoCommonService) *ShortIDMiddleware {
	return &ShortIDMiddleware{
		siteInfoService: siteInfoService,
	}
}

func (sm *ShortIDMiddleware) SetShortIDFlag() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		siteSeo, err := sm.siteInfoService.GetSiteSeo(ctx)
		if err != nil {
			log.Error(err)
			return
		}
		ctx.Set(constant.ShortIDFlag, siteSeo.IsShortLink())
	}
}
