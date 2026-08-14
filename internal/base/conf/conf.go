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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/path"
	"github.com/apache/answer/internal/base/server"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/router"
	"github.com/apache/answer/internal/service/service_config"
	"github.com/apache/answer/pkg/writer"
	"github.com/segmentfault/pacman/contrib/conf/viper"
	"gopkg.in/yaml.v3"
)

// AllConfig all config
type AllConfig struct {
	Debug         bool                          `json:"debug" mapstructure:"debug" yaml:"debug"`
	Server        *Server                       `json:"server" mapstructure:"server" yaml:"server"`
	Data          *Data                         `json:"data" mapstructure:"data" yaml:"data"`
	I18n          *translator.I18n              `json:"i18n" mapstructure:"i18n" yaml:"i18n"`
	ServiceConfig *service_config.ServiceConfig `json:"service_config" mapstructure:"service_config" yaml:"service_config"`
	Swaggerui     *router.SwaggerConfig         `json:"swaggerui" mapstructure:"swaggerui" yaml:"swaggerui"`
	UI            *server.UI                    `json:"ui" mapstructure:"ui" yaml:"ui"`
}

type envConfigOverrides struct {
	SwaggerHost        string
	SwaggerAddressPort string
	SiteAddr           string
	TrustedProxies     string
	CacheType          string
	RedisHost          string
	RedisPort          string
	RedisUsername      string
	RedisPassword      string
	RedisDB            string
	RedisKeyPrefix     string
	RedisPoolSize      string
}

func loadEnvs() (envOverrides *envConfigOverrides) {
	return &envConfigOverrides{
		SwaggerHost:        os.Getenv("SWAGGER_HOST"),
		SwaggerAddressPort: os.Getenv("SWAGGER_ADDRESS_PORT"),
		SiteAddr:           os.Getenv("SITE_ADDR"),
		TrustedProxies:     os.Getenv("TRUSTED_PROXIES"),
		CacheType:          os.Getenv("CACHE_TYPE"),
		RedisHost:          os.Getenv("REDIS_HOST"),
		RedisPort:          os.Getenv("REDIS_PORT"),
		RedisUsername:      os.Getenv("REDIS_USERNAME"),
		RedisPassword:      os.Getenv("REDIS_PASSWORD"),
		RedisDB:            os.Getenv("REDIS_DB"),
		RedisKeyPrefix:     os.Getenv("REDIS_KEY_PREFIX"),
		RedisPoolSize:      os.Getenv("REDIS_POOL_SIZE"),
	}
}

type PathIgnore struct {
	Users []string `yaml:"users"`
}

// Server server config
type Server struct {
	HTTP *server.HTTP `json:"http" mapstructure:"http" yaml:"http"`
}

// Data data config
type Data struct {
	Database *data.Database  `json:"database" mapstructure:"database" yaml:"database"`
	Cache    *data.CacheConf `json:"cache" mapstructure:"cache" yaml:"cache"`
}

// SetDefault set default config
func (c *AllConfig) SetDefault() {
	if c.Server == nil {
		c.Server = &Server{}
	}
	if c.Server.HTTP == nil {
		c.Server.HTTP = &server.HTTP{}
	}
	if c.Server.HTTP.TrustedProxies == nil {
		c.Server.HTTP.TrustedProxies = []string{"127.0.0.1", "::1"}
	}
	if c.UI == nil {
		c.UI = &server.UI{}
	}
	if c.Data == nil {
		c.Data = &Data{}
	}
	if c.Data.Cache == nil {
		c.Data.Cache = &data.CacheConf{}
	}

	// Application configuration defaults to Redis.
	// Direct NewCache(&CacheConf{}) calls still use Memory for tests.
	if c.Data.Cache.Type == "" {
		c.Data.Cache.Type = data.CacheTypeRedis
	}
	if c.Data.Cache.Redis.Host == "" {
		c.Data.Cache.Redis.Host = data.DefaultRedisHost
	}
	if c.Data.Cache.Redis.Port == 0 {
		c.Data.Cache.Redis.Port = data.DefaultRedisPort
	}
	if c.Data.Cache.Redis.KeyPrefix == "" {
		c.Data.Cache.Redis.KeyPrefix = data.DefaultRedisKeyPrefix
	}
	if c.Data.Cache.Redis.PoolSize == 0 {
		c.Data.Cache.Redis.PoolSize = data.DefaultRedisPoolSize
	}
}

func (c *AllConfig) SetEnvironmentOverrides() error {
	envs := loadEnvs()

	if envs.SiteAddr != "" {
		c.Server.HTTP.Addr = envs.SiteAddr
	}
	if envs.TrustedProxies != "" {
		c.Server.HTTP.TrustedProxies = c.Server.HTTP.TrustedProxies[:0]
		if !strings.EqualFold(strings.TrimSpace(envs.TrustedProxies), "none") {
			for _, proxy := range strings.Split(envs.TrustedProxies, ",") {
				if proxy = strings.TrimSpace(proxy); proxy != "" {
					c.Server.HTTP.TrustedProxies = append(c.Server.HTTP.TrustedProxies, proxy)
				}
			}
		}
	}
	if envs.SwaggerHost != "" {
		c.Swaggerui.Host = envs.SwaggerHost
	}
	if envs.SwaggerAddressPort != "" {
		c.Swaggerui.Address = envs.SwaggerAddressPort
	}
	if envs.CacheType != "" {
		c.Data.Cache.Type = envs.CacheType
	}
	if envs.RedisHost != "" {
		c.Data.Cache.Redis.Host = envs.RedisHost
	}
	if envs.RedisPort != "" {
		port, err := strconv.Atoi(envs.RedisPort)
		if err != nil || port < 1 || port > 65535 {
			return fmt.Errorf("REDIS_PORT must be an integer between 1 and 65535, got %q", envs.RedisPort)
		}
		c.Data.Cache.Redis.Port = port
	}
	if envs.RedisUsername != "" {
		c.Data.Cache.Redis.Username = envs.RedisUsername
	}
	if envs.RedisPassword != "" {
		c.Data.Cache.Redis.Password = envs.RedisPassword
	}
	if envs.RedisDB != "" {
		db, err := strconv.Atoi(envs.RedisDB)
		if err != nil || db < 0 {
			return fmt.Errorf("REDIS_DB must be a non-negative integer, got %q", envs.RedisDB)
		}
		c.Data.Cache.Redis.DB = db
	}
	if envs.RedisKeyPrefix != "" {
		c.Data.Cache.Redis.KeyPrefix = envs.RedisKeyPrefix
	}
	if envs.RedisPoolSize != "" {
		poolSize, err := strconv.Atoi(envs.RedisPoolSize)
		if err != nil || poolSize <= 0 {
			return fmt.Errorf(
				"REDIS_POOL_SIZE must be a positive integer, got %q",
				envs.RedisPoolSize,
			)
		}
		c.Data.Cache.Redis.PoolSize = poolSize
	}

	return nil
}

// ReadConfig read config
func ReadConfig(configFilePath string) (c *AllConfig, err error) {
	if len(configFilePath) == 0 {
		configFilePath = filepath.Join(path.ConfigFileDir, path.DefaultConfigFileName)
	}
	c = &AllConfig{}
	config, err := viper.NewWithPath(configFilePath)
	if err != nil {
		return nil, err
	}
	if err = config.Parse(&c); err != nil {
		return nil, err
	}
	c.SetDefault()
	if err = c.SetEnvironmentOverrides(); err != nil {
		return nil, err
	}
	return c, nil
}

// RewriteConfig rewrite config file path
func RewriteConfig(configFilePath string, allConfig *AllConfig) error {
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	defer func() {
		_ = enc.Close()
	}()
	enc.SetIndent(2)
	if err := enc.Encode(allConfig); err != nil {
		return err
	}
	return writer.ReplaceFile(configFilePath, buf.String())
}
