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

package day

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestFormat(t *testing.T) {
	sec := int64(1700000000)
	tz := "UTC"

	actual := Format(sec, "MMM D, YYYY [a las] HH:mm", tz)
	expected := time.Unix(sec, 0).Format("Jan 2, 2006") + " a las " + time.Unix(sec, 0).Format("15:04")
	assert.Equal(t, expected, actual)
}

func TestFormat_AllLanguagesNoHang(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("unable to determine test file path")
	}
	i18nDir := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "i18n"))
	entries, err := os.ReadDir(i18nDir)
	if err != nil {
		t.Fatalf("read i18n dir: %v", err)
	}

	type datesConfig struct {
		LongDate         string `yaml:"long_date"`
		LongDateWithYear string `yaml:"long_date_with_year"`
		LongDateWithTime string `yaml:"long_date_with_time"`
	}
	type uiConfig struct {
		Dates datesConfig `yaml:"dates"`
	}
	type fileConfig struct {
		UI uiConfig `yaml:"ui"`
	}

	sec := int64(1700000000)
	tz := "UTC"

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || filepath.Ext(name) != ".yaml" || name == "i18n.yaml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(i18nDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		var cfg fileConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}

		formats := []string{
			cfg.UI.Dates.LongDate,
			cfg.UI.Dates.LongDateWithYear,
			cfg.UI.Dates.LongDateWithTime,
		}
		for _, format := range formats {
			if format == "" {
				continue
			}
			done := make(chan struct{})
			go func(f string) {
				_ = Format(sec, f, tz)
				close(done)
			}(format)

			select {
			case <-done:
			case <-time.After(200 * time.Millisecond):
				t.Fatalf("format hang in %s: %q", name, format)
			}
		}
	}
}
