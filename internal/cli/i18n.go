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
	"github.com/apache/answer/i18n"
	"github.com/apache/answer/pkg/dir"
	"github.com/apache/answer/pkg/writer"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type YamlPluginContent struct {
	Plugin map[string]any `yaml:"plugin"`
}

// ReplaceI18nFilesLocal replace i18n files
func ReplaceI18nFilesLocal(i18nDir string) error {
	i18nList, err := i18n.I18n.ReadDir(".")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Printf("[i18n] find i18n bundle %d\n", len(i18nList))
	for _, item := range i18nList {
		path := filepath.Join(i18nDir, item.Name())
		content, err := i18n.I18n.ReadFile(item.Name())
		if err != nil {
			continue
		}
		exist := dir.CheckFileExist(path)
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
	return nil
}

// MergeI18nFilesLocal merge i18n files
func MergeI18nFilesLocal(originalI18nDir, targetI18nDir string) (err error) {
	pluginAllTranslations := make(map[string]*YamlPluginContent)

	err = findI18nFileInDir(pluginAllTranslations, targetI18nDir)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(originalI18nDir)
	if err != nil {
		return err
	}

	for _, file := range entries {
		// ignore directory
		if file.IsDir() {
			continue
		}
		// ignore non-YAML file
		filename := file.Name()
		if filepath.Ext(filename) != ".yaml" && filename != "i18n.yaml" {
			continue
		}

		// if plugin don't have this translation file, ignore it
		if pluginAllTranslations[filename] == nil {
			continue
		}

		out, _ := yaml.Marshal(pluginAllTranslations[filename])

		buf, err := os.OpenFile(filepath.Join(originalI18nDir, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("[i18n] read translation file failed: %s %s\n", filename, err)
			continue
		}

		_, _ = buf.WriteString("\n")
		_, _ = buf.Write(out)
		_ = buf.Close()
		fmt.Printf("[i18n] merge i18n file: %s success\n", filename)
	}

	return nil
}

// find i18n file in dir
func findI18nFileInDir(pluginAllTranslations map[string]*YamlPluginContent, i18nDir string) error {
	// if i18n dir is not i18n, find deeper
	dirBase := filepath.Base(i18nDir)
	if dirBase != "i18n" {
		if strings.HasPrefix(dirBase, ".") {
			return nil
		}
		// find all i18n dir in target dir
		targetDirs, err := os.ReadDir(i18nDir)
		if err != nil {
			return err
		}

		for _, targetDir := range targetDirs {
			if targetDir.IsDir() {
				if err := findI18nFileInDir(pluginAllTranslations, filepath.Join(i18nDir, targetDir.Name())); err != nil {
					fmt.Printf("[i18n] find i18n file in dir failed: %s %s\n", targetDir.Name(), err)
				}
			}
		}
		return nil
	}

	fmt.Printf("[i18n] find i18n file in dir: %s\n", i18nDir)

	// if i18nDir is i18n, find all yaml files
	entries, err := os.ReadDir(i18nDir)
	if err != nil {
		return err
	}

	for _, file := range entries {
		// ignore directory
		if file.IsDir() {
			continue
		}
		// ignore non-YAML file
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}
		buf, err := os.ReadFile(filepath.Join(i18nDir, file.Name()))
		if err != nil {
			fmt.Printf("[i18n] read translation file failed: %s %s\n", file.Name(), err)
			continue
		}

		translation := &YamlPluginContent{}
		if err = yaml.Unmarshal(buf, translation); err != nil {
			fmt.Printf("[i18n] unmarshal translation file failed: %s %s\n", file.Name(), err)
			continue
		}

		if pluginAllTranslations[file.Name()] == nil {
			pluginAllTranslations[file.Name()] = &YamlPluginContent{Plugin: make(map[string]any)}
		}
		for k, v := range translation.Plugin {
			pluginAllTranslations[file.Name()].Plugin[k] = v
		}
	}
	return nil
}
