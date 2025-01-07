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

package entity

import (
	"encoding/json"
	"github.com/segmentfault/pacman/log"

	"github.com/apache/answer/pkg/converter"
)

// Config config
type Config struct {
	ID    int    `xorm:"not null pk autoincr INT(11) id"`
	Key   string `xorm:"unique VARCHAR(128) key"`
	Value string `xorm:"TEXT value"`
}

// TableName config table name
func (c *Config) TableName() string {
	return "config"
}

func (c *Config) BuildByJSON(data []byte) {
	cf := &Config{}
	_ = json.Unmarshal(data, cf)
	c.ID = cf.ID
	c.Key = cf.Key
	c.Value = cf.Value
}

func (c *Config) JsonString() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// GetIntValue get int value
func (c *Config) GetIntValue() int {
	if len(c.Value) == 0 {
		log.Warnf("config value is empty, key: %s, value: %s", c.Key, c.Value)
	}
	return converter.StringToInt(c.Value)
}

// GetArrayStringValue get array string value
func (c *Config) GetArrayStringValue() []string {
	var arr []string
	_ = json.Unmarshal([]byte(c.Value), &arr)
	return arr
}

// GetByteValue get byte value
func (c *Config) GetByteValue() []byte {
	return []byte(c.Value)
}
