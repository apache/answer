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

package path

import (
	"path/filepath"
	"sync"
)

const (
	DefaultConfigFileName                  = "config.yaml"
	DefaultCacheFileName                   = "cache.db"
	DefaultReservedUsernamesConfigFileName = "reserved-usernames.json"
)

var (
	ConfigFileDir     = "/conf/"
	UploadFilePath    = "/uploads/"
	I18nPath          = "/i18n/"
	CacheDir          = "/cache/"
	formatAllPathOnce sync.Once
)

func FormatAllPath(dataDirPath string) {
	formatAllPathOnce.Do(func() {
		ConfigFileDir = filepath.Join(dataDirPath, ConfigFileDir)
		UploadFilePath = filepath.Join(dataDirPath, UploadFilePath)
		I18nPath = filepath.Join(dataDirPath, I18nPath)
		CacheDir = filepath.Join(dataDirPath, CacheDir)
	})
}

// GetConfigFilePath get config file path
func GetConfigFilePath() string {
	return filepath.Join(ConfigFileDir, DefaultConfigFileName)
}
