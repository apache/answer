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

package translator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/wire"
	myTran "github.com/segmentfault/pacman/contrib/i18n"
	"github.com/segmentfault/pacman/i18n"
	"github.com/segmentfault/pacman/log"
	"gopkg.in/yaml.v3"
)

// ProviderSet is providers.
var ProviderSet = wire.NewSet(NewTranslator)
var GlobalTrans i18n.Translator

// LangOption language option
type LangOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
	// Translation completion percentage
	Progress int `json:"progress"`
}

// DefaultLangOption default language option. If user config the language is default, the language option is admin choose.
const DefaultLangOption = "Default"

var (
	// LanguageOptions language
	LanguageOptions []*LangOption
)

// NewTranslator new a translator
func NewTranslator(c *I18n) (tr i18n.Translator, err error) {
	entries, err := os.ReadDir(c.BundleDir)
	if err != nil {
		return nil, err
	}

	// read the Bundle resources file from entries
	for _, file := range entries {
		// ignore directory
		if file.IsDir() {
			continue
		}
		// ignore non-YAML file
		if filepath.Ext(file.Name()) != ".yaml" && file.Name() != "i18n.yaml" {
			continue
		}
		log.Debugf("try to read file: %s", file.Name())
		buf, err := os.ReadFile(filepath.Join(c.BundleDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("read file failed: %s %s", file.Name(), err)
		}

		// parse the backend translation
		originalTr := struct {
			Backend map[string]map[string]any `yaml:"backend"`
			UI      map[string]any            `yaml:"ui"`
			Plugin  map[string]any            `yaml:"plugin"`
		}{}
		if err = yaml.Unmarshal(buf, &originalTr); err != nil {
			return nil, err
		}
		translation := make(map[string]any, 0)
		for k, v := range originalTr.Backend {
			translation[k] = v
		}
		translation["backend"] = originalTr.Backend
		translation["ui"] = originalTr.UI
		translation["plugin"] = originalTr.Plugin

		content, err := yaml.Marshal(translation)
		if err != nil {
			log.Debugf("marshal translation content failed: %s %s", file.Name(), err)
			continue
		}

		// add translator use backend translation
		if err = myTran.AddTranslator(content, file.Name()); err != nil {
			log.Debugf("add translator failed: %s %s", file.Name(), err)
			reportTranslatorFormatError(file.Name(), buf)
			continue
		}
	}
	GlobalTrans = myTran.GlobalTrans

	i18nFile, err := os.ReadFile(filepath.Join(c.BundleDir, "i18n.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read i18n file failed: %s", err)
	}

	s := struct {
		LangOption []*LangOption `yaml:"language_options"`
	}{}
	err = yaml.Unmarshal(i18nFile, &s)
	if err != nil {
		return nil, fmt.Errorf("i18n file parsing failed: %s", err)
	}
	LanguageOptions = s.LangOption
	for _, option := range LanguageOptions {
		option.Label = fmt.Sprintf("%s (%d%%)", option.Label, option.Progress)
	}
	return GlobalTrans, err
}

// CheckLanguageIsValid check user input language is valid
func CheckLanguageIsValid(lang string) bool {
	if lang == DefaultLangOption {
		return true
	}
	for _, option := range LanguageOptions {
		if option.Value == lang {
			return true
		}
	}
	return false
}

// Tr use language to translate data. If this language translation is not available, return default english translation.
func Tr(lang i18n.Language, data string) string {
	if GlobalTrans == nil {
		return data
	}
	translation := GlobalTrans.Tr(lang, data)
	if translation == data {
		return GlobalTrans.Tr(i18n.DefaultLanguage, data)
	}
	return translation
}

// TrWithData translate key with template data, it will replace the template data {{ .PlaceHolder }} in the translation.
func TrWithData(lang i18n.Language, key string, templateData any) string {
	if GlobalTrans == nil {
		return key
	}
	translation := GlobalTrans.TrWithData(lang, key, templateData)
	if translation == key {
		return GlobalTrans.TrWithData(i18n.DefaultLanguage, key, templateData)
	}
	return translation
}

// reportTranslatorFormatError re-parses the YAML file to locate the invalid entry
// when go-i18n fails to add the translator.
func reportTranslatorFormatError(fileName string, content []byte) {
	var raw any
	if err := yaml.Unmarshal(content, &raw); err != nil {
		log.Errorf("parse translator file %s failed when diagnosing format error: %s", fileName, err)
		return
	}
	if err := inspectTranslatorNode(raw, nil, true); err != nil {
		log.Errorf("translator file %s invalid: %s", fileName, err)
	}
}

func inspectTranslatorNode(node any, path []string, isRoot bool) error {
	switch data := node.(type) {
	case nil:
		if isRoot {
			return fmt.Errorf("root value is empty")
		}
		return fmt.Errorf("%s contains an empty value", formatTranslationPath(path))
	case string:
		if isRoot {
			return fmt.Errorf("root value must be an object but found string")
		}
		return nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		if isRoot {
			return fmt.Errorf("root value must be an object but found %T", data)
		}
		return fmt.Errorf("%s expects a string translation but found %T", formatTranslationPath(path), data)
	case map[string]any:
		if isMessageMap(data) {
			return nil
		}
		keys := make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := inspectTranslatorNode(data[key], append(path, key), false); err != nil {
				return err
			}
		}
		return nil
	case map[string]string:
		mapped := make(map[string]any, len(data))
		for k, v := range data {
			mapped[k] = v
		}
		return inspectTranslatorNode(mapped, path, isRoot)
	case map[any]any:
		if isMessageMap(data) {
			return nil
		}
		type kv struct {
			key string
			val any
		}
		items := make([]kv, 0, len(data))
		for key, val := range data {
			strKey, ok := key.(string)
			if !ok {
				return fmt.Errorf("%s uses a non-string key %#v", formatTranslationPath(path), key)
			}
			items = append(items, kv{key: strKey, val: val})
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].key < items[j].key
		})
		for _, item := range items {
			if err := inspectTranslatorNode(item.val, append(path, item.key), false); err != nil {
				return err
			}
		}
		return nil
	case []any:
		for idx, child := range data {
			nextPath := append(path, fmt.Sprintf("[%d]", idx)) //nolint: gocritic
			if err := inspectTranslatorNode(child, nextPath, false); err != nil {
				return err
			}
		}
		return nil
	case []map[string]any:
		for idx, child := range data {
			nextPath := append(path, fmt.Sprintf("[%d]", idx)) //nolint: gocritic
			if err := inspectTranslatorNode(child, nextPath, false); err != nil {
				return err
			}
		}
		return nil
	default:
		if isRoot {
			return fmt.Errorf("root value must be an object but found %T", data)
		}
		return fmt.Errorf("%s contains unsupported value type %T", formatTranslationPath(path), data)
	}
}

var translatorReservedKeys = []string{
	"id", "description", "hash", "leftdelim", "rightdelim",
	"zero", "one", "two", "few", "many", "other",
}

func isMessageMap(data any) bool {
	switch v := data.(type) {
	case map[string]any:
		for _, key := range translatorReservedKeys {
			val, ok := v[key]
			if !ok {
				continue
			}
			if _, ok := val.(string); ok {
				return true
			}
		}
	case map[string]string:
		for _, key := range translatorReservedKeys {
			val, ok := v[key]
			if !ok {
				continue
			}
			if val != "" {
				return true
			}
		}
	case map[any]any:
		for _, key := range translatorReservedKeys {
			val, ok := v[key]
			if !ok {
				continue
			}
			if _, ok := val.(string); ok {
				return true
			}
		}
	}
	return false
}

func formatTranslationPath(path []string) string {
	if len(path) == 0 {
		return "root"
	}
	var b strings.Builder
	for _, part := range path {
		if part == "" {
			continue
		}
		if part[0] == '[' {
			b.WriteString(part)
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(part)
	}
	return b.String()
}
