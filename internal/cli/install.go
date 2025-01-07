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
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/apache/answer/configs"
	"github.com/apache/answer/i18n"
	"github.com/apache/answer/pkg/dir"
	"github.com/apache/answer/pkg/writer"
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
	formatAllPathONCE sync.Once
)

// GetConfigFilePath get config file path
func GetConfigFilePath() string {
	return filepath.Join(ConfigFileDir, DefaultConfigFileName)
}

func FormatAllPath(dataDirPath string) {
	formatAllPathONCE.Do(func() {
		ConfigFileDir = filepath.Join(dataDirPath, ConfigFileDir)
		UploadFilePath = filepath.Join(dataDirPath, UploadFilePath)
		I18nPath = filepath.Join(dataDirPath, I18nPath)
		CacheDir = filepath.Join(dataDirPath, CacheDir)
	})
}

// InstallAllInitialEnvironment install all initial environment
func InstallAllInitialEnvironment(dataDirPath string) {
	FormatAllPath(dataDirPath)
	installUploadDir()
	InstallI18nBundle(false)
	fmt.Println("install all initial environment done")
}

func InstallConfigFile(configFilePath string) error {
	if len(configFilePath) == 0 {
		configFilePath = filepath.Join(ConfigFileDir, DefaultConfigFileName)
	}
	fmt.Println("[config-file] try to create at ", configFilePath)

	// if config file already exists do nothing.
	if CheckConfigFile(configFilePath) {
		fmt.Printf("[config-file] %s already exists\n", configFilePath)
		return nil
	}

	if err := dir.CreateDirIfNotExist(ConfigFileDir); err != nil {
		fmt.Printf("[config-file] create directory fail %s\n", err.Error())
		return fmt.Errorf("create directory fail %s", err.Error())
	}
	fmt.Printf("[config-file] create directory success, config file is %s\n", configFilePath)

	if err := writer.WriteFile(configFilePath, string(configs.Config)); err != nil {
		fmt.Printf("[config-file] install fail %s\n", err.Error())
		return fmt.Errorf("write file failed %s", err)
	}
	fmt.Printf("[config-file] install success\n")
	return nil
}

func installUploadDir() {
	fmt.Println("[upload-dir] try to install...")
	if err := dir.CreateDirIfNotExist(UploadFilePath); err != nil {
		fmt.Printf("[upload-dir] install fail %s\n", err.Error())
	} else {
		fmt.Printf("[upload-dir] install success, upload directory is %s\n", UploadFilePath)
	}
}

func InstallI18nBundle(replace bool) {
	fmt.Println("[i18n] try to install i18n bundle...")
	// if SKIP_REPLACE_I18N is set, skip replace i18n bundles
	if len(os.Getenv("SKIP_REPLACE_I18N")) > 0 {
		replace = false
	}
	if err := dir.CreateDirIfNotExist(I18nPath); err != nil {
		fmt.Println(err.Error())
		return
	}

	i18nList, err := i18n.I18n.ReadDir(".")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("[i18n] find i18n bundle %d\n", len(i18nList))
	for _, item := range i18nList {
		path := filepath.Join(I18nPath, item.Name())
		content, err := i18n.I18n.ReadFile(item.Name())
		if err != nil {
			continue
		}
		exist := dir.CheckFileExist(path)
		if exist && !replace {
			continue
		}
		if exist {
			fmt.Printf("[i18n] install %s file exist, try to replace it\n", item.Name())
			if err = os.Remove(path); err != nil {
				fmt.Println(err)
			}
		}
		fmt.Printf("[i18n] install %s bundle...\n", item.Name())
		err = writer.WriteFile(path, string(content))
		if err != nil {
			fmt.Printf("[i18n] install %s bundle fail: %s\n", item.Name(), err.Error())
		} else {
			fmt.Printf("[i18n] install %s bundle success\n", item.Name())
		}
	}
}
