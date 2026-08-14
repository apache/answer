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

package conf

import (
	"testing"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/server"
	"github.com/apache/answer/internal/router"
	"github.com/stretchr/testify/require"
)

func newTestConfig() *AllConfig {
	return &AllConfig{
		Server:    &Server{HTTP: &server.HTTP{}},
		Data:      &Data{Cache: &data.CacheConf{}},
		Swaggerui: &router.SwaggerConfig{},
	}
}

func TestRedisConfigDefaults(t *testing.T) {
	config := newTestConfig()
	config.SetDefault()

	require.Equal(t, data.CacheTypeRedis, config.Data.Cache.Type)
	require.Equal(t, data.DefaultRedisHost, config.Data.Cache.Redis.Host)
	require.Equal(t, data.DefaultRedisPort, config.Data.Cache.Redis.Port)
	require.Equal(t, data.DefaultRedisKeyPrefix, config.Data.Cache.Redis.KeyPrefix)
	require.Equal(t, data.DefaultRedisPoolSize, config.Data.Cache.Redis.PoolSize)
	require.Equal(t, []string{"127.0.0.1", "::1"}, config.Server.HTTP.TrustedProxies)
}

func TestRedisEnvironmentOverrides(t *testing.T) {
	t.Setenv("CACHE_TYPE", "redis")
	t.Setenv("REDIS_HOST", "redis.internal")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_USERNAME", "answer")
	t.Setenv("REDIS_PASSWORD", "test-password")
	t.Setenv("REDIS_DB", "3")
	t.Setenv("REDIS_KEY_PREFIX", "hnu-forum:test:")
	t.Setenv("REDIS_POOL_SIZE", "40")

	config := newTestConfig()
	config.SetDefault()
	require.NoError(t, config.SetEnvironmentOverrides())

	require.Equal(t, "redis", config.Data.Cache.Type)
	require.Equal(t, "redis.internal", config.Data.Cache.Redis.Host)
	require.Equal(t, 6380, config.Data.Cache.Redis.Port)
	require.Equal(t, "answer", config.Data.Cache.Redis.Username)
	require.Equal(t, "test-password", config.Data.Cache.Redis.Password)
	require.Equal(t, 3, config.Data.Cache.Redis.DB)
	require.Equal(t, "hnu-forum:test:", config.Data.Cache.Redis.KeyPrefix)
	require.Equal(t, 40, config.Data.Cache.Redis.PoolSize)
}

func TestTrustedProxyEnvironmentOverride(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "10.0.0.2, 10.0.0.0/24")
	config := newTestConfig()
	config.SetDefault()
	require.NoError(t, config.SetEnvironmentOverrides())
	require.Equal(t, []string{"10.0.0.2", "10.0.0.0/24"}, config.Server.HTTP.TrustedProxies)
}

func TestTrustedProxyEnvironmentCanDisableProxyTrust(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "none")
	config := newTestConfig()
	config.SetDefault()
	require.NoError(t, config.SetEnvironmentOverrides())
	require.Empty(t, config.Server.HTTP.TrustedProxies)
}

func TestInvalidRedisEnvironment(t *testing.T) {
	t.Setenv("REDIS_PORT", "invalid")

	config := newTestConfig()
	config.SetDefault()
	require.Error(t, config.SetEnvironmentOverrides())
}
