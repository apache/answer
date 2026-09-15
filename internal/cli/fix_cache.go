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

package cli

import (
	"context"
	"fmt"
	"slices"

	"github.com/apache/answer/internal/base/constant"
	"github.com/segmentfault/pacman/contrib/cache/memory"
)

func invalidateSiteInfoCache(cacheFilePath string, fixTypes []string) {
	if !slices.Contains(fixTypes, fixTypeBranding) {
		return
	}

	// The branding site-info cache only lives in the file-backed memory cache.
	// When a cache plugin (e.g. Redis) is used, no file path is configured, so
	// warn the admin to flush it manually instead of silently doing nothing.
	if len(cacheFilePath) == 0 {
		fmt.Println("[cache] no file cache configured; flush your cache backend or restart the application to refresh branding")
		return
	}

	memCache := memory.NewCache()
	if err := memory.Load(memCache, cacheFilePath); err != nil {
		fmt.Printf("[cache] cannot load cache file %s: %v (skipping)\n", cacheFilePath, err)
		return
	}

	key := constant.SiteInfoCacheKey + constant.SiteTypeBranding
	if err := memCache.Del(context.Background(), key); err != nil {
		fmt.Printf("[cache] failed to delete key %s: %v\n", key, err)
	} else {
		fmt.Printf("[cache] invalidated %s\n", key)
	}

	if err := memory.Save(memCache, cacheFilePath); err != nil {
		fmt.Printf("[cache] failed to save cache file: %v\n", err)
		return
	}

	fmt.Println("[cache] cache file updated, restart the application to take full effect")
}
