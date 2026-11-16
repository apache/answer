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

package constant

import "time"

const (
	UserStatusChangedCacheKey                  = "answer:user:status:"
	UserStatusChangedCacheTime                 = 7 * 24 * time.Hour
	UserTokenCacheKey                          = "answer:user:token:"
	UserTokenCacheTime                         = 7 * 24 * time.Hour
	UserVisitTokenCacheKey                     = "answer:user:visit:"
	UserVisitCacheTime                         = 7 * 24 * 60 * 60
	UserVisitCookiesCacheKey                   = "visit"
	AdminTokenCacheKey                         = "answer:admin:token:"
	AdminTokenCacheTime                        = 7 * 24 * time.Hour
	UserTokenMappingCacheKey                   = "answer:user-token:mapping:"
	UserEmailCodeCacheKey                      = "answer:user:email-code:"
	UserEmailCodeCacheTime                     = 10 * time.Minute
	UserLatestEmailCodeCacheKey                = "answer:user-id:email-code:"
	SiteInfoCacheKey                           = "answer:site-info:"
	SiteInfoCacheTime                          = 1 * time.Hour
	ConfigID2KEYCacheKeyPrefix                 = "answer:config:id:"
	ConfigKEY2ContentCacheKeyPrefix            = "answer:config:key:"
	ConfigCacheTime                            = 1 * time.Hour
	ConnectorUserExternalInfoCacheKey          = "answer:connector:"
	ConnectorUserExternalInfoCacheTime         = 10 * time.Minute
	ConnectorOAuthStateCacheKey                = "answer:connector:oauth-state:"
	ConnectorOAuthStateCacheTime               = 10 * time.Minute
	ConnectorOAuthBindStateCacheTime           = 5 * time.Minute
	SiteMapQuestionCacheKeyPrefix              = "answer:sitemap:question:%d"
	SiteMapQuestionCacheTime                   = time.Hour
	SitemapMaxSize                             = 50000
	NewQuestionNotificationLimitCacheKeyPrefix = "answer:new-question-notification-limit:"
	NewQuestionNotificationLimitCacheTime      = 7 * 24 * time.Hour
	NewQuestionNotificationLimitMax            = 50
	RateLimitCacheKeyPrefix                    = "answer:rate-limit:"
	RateLimitCacheTime                         = 5 * time.Minute
	RedDotCacheKey                             = "answer:red-dot:%s:%s"
	RedDotCacheTime                            = 30 * 24 * time.Hour
)
