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

package data

const (
	CacheTypeMemory = "memory"
	CacheTypeRedis  = "redis"

	DefaultRedisHost      = "127.0.0.1"
	DefaultRedisPort      = 6379
	DefaultRedisKeyPrefix = "hnu-forum:"
	DefaultRedisPoolSize  = 20
)

// Database database config
type Database struct {
	Driver          string `json:"driver" mapstructure:"driver" yaml:"driver"`
	Connection      string `json:"connection" mapstructure:"connection" yaml:"connection"`
	ConnMaxLifeTime int    `json:"conn_max_life_time" mapstructure:"conn_max_life_time" yaml:"conn_max_life_time,omitempty"`
	MaxOpenConn     int    `json:"max_open_conn" mapstructure:"max_open_conn" yaml:"max_open_conn,omitempty"`
	MaxIdleConn     int    `json:"max_idle_conn" mapstructure:"max_idle_conn" yaml:"max_idle_conn,omitempty"`
}

// CacheConf cache
type CacheConf struct {
	Type     string         `json:"type" mapstructure:"type" yaml:"type"`
	FilePath string         `json:"file_path" mapstructure:"file_path" yaml:"file_path,omitempty"`
	Redis    RedisCacheConf `json:"redis" mapstructure:"redis" yaml:"redis"`
}

// RedisCacheConf configures Redis as the shared application cache.
type RedisCacheConf struct {
	Host      string `json:"host" mapstructure:"host" yaml:"host"`
	Port      int    `json:"port" mapstructure:"port" yaml:"port"`
	Username  string `json:"username" mapstructure:"username" yaml:"username,omitempty"`
	Password  string `json:"password" mapstructure:"password" yaml:"password,omitempty"`
	DB        int    `json:"db" mapstructure:"db" yaml:"db"`
	KeyPrefix string `json:"key_prefix" mapstructure:"key_prefix" yaml:"key_prefix"`
	PoolSize  int    `json:"pool_size" mapstructure:"pool_size" yaml:"pool_size,omitempty"`
}
