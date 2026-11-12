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
	expected := time.Unix(sec, 0).UTC().Format("Jan 2, 2006") + " a las " + time.Unix(sec, 0).UTC().Format("15:04")
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

// Format takes an IANA time zone name and must render the instant in that zone.
// Before the fix the loaded location was discarded, so every caller (the UI,
// the email templates and the "account suspended until" notice, which passes
// "UTC" explicitly) got the server's local zone instead.
func TestFormatTimezone(t *testing.T) {
	// 2023-11-14 22:13:20 UTC
	const sec = int64(1700000000)

	tests := []struct {
		name     string
		format   string
		tz       string
		expected string
	}{
		{"utc", "YYYY-MM-DD HH:mm", "UTC", "2023-11-14 22:13"},
		{"tokyo_crosses_midnight", "YYYY-MM-DD HH:mm", "Asia/Tokyo", "2023-11-15 07:13"},
		{"new_york", "YYYY-MM-DD HH:mm", "America/New_York", "2023-11-14 17:13"},
		{"kolkata_half_hour_offset", "YYYY-MM-DD HH:mm", "Asia/Kolkata", "2023-11-15 03:43"},
		{"los_angeles_long_date", "MMMM D, YYYY HH:mm", "America/Los_Angeles", "November 14, 2023 14:13"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Format(sec, tt.format, tt.tz))
		})
	}
}

// A zone that cannot be resolved must fall back to the server's local zone
// instead of failing or silently shifting the instant.
func TestFormatUnknownTimezoneFallsBackToLocal(t *testing.T) {
	const sec = int64(1700000000)
	expected := time.Unix(sec, 0).In(time.Local).Format("2006-01-02 15:04")

	for _, tz := range []string{"", "Not/AZone", "UTC+99"} {
		t.Run("tz="+tz, func(t *testing.T) {
			assert.Equal(t, expected, Format(sec, "YYYY-MM-DD HH:mm", tz))
		})
	}
}

// Location resolves an IANA name and must fall back to the server's local
// zone for empty or unknown names, so a bad configuration never shifts the
// rendered instant.
func TestLocation(t *testing.T) {
	assert.Equal(t, time.UTC, Location("UTC"))
	assert.Equal(t, time.Local, Location(""))
	assert.Equal(t, time.Local, Location("Not/AZone"))

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	assert.Nil(t, err)
	assert.Equal(t, tokyo, Location("Asia/Tokyo"))
}

// FormatISO feeds every server-rendered <time datetime="..."> attribute: it
// renders the instant in the requested zone with a real ISO-8601 marker -
// "Z" only for UTC, an explicit offset otherwise.
func TestFormatISO(t *testing.T) {
	// 2023-11-14 22:13:20 UTC
	const sec = int64(1700000000)

	tests := []struct {
		name     string
		tz       string
		expected string
	}{
		{"utc_renders_z", "UTC", "2023-11-14T22:13:20.000Z"},
		{"tokyo_offset", "Asia/Tokyo", "2023-11-15T07:13:20.000+09:00"},
		{"new_york_offset", "America/New_York", "2023-11-14T17:13:20.000-05:00"},
		{"kolkata_half_hour", "Asia/Kolkata", "2023-11-15T03:43:20.000+05:30"},
		{"london_winter_is_z", "Europe/London", "2023-11-14T22:13:20.000Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, FormatISO(sec, tt.tz))
		})
	}
}
