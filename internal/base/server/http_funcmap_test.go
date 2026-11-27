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

package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// timeFormatISO feeds every server-rendered <time datetime="..."> attribute:
// it must render the instant in the requested zone with a real ISO-8601 zone
// marker ("Z" only for UTC), not the server's local wall clock labeled as Z.
func TestTimeFormatISO(t *testing.T) {
	// 2023-11-14 22:13:20 UTC
	const ts = int64(1700000000)
	formatISO := funcMap["timeFormatISO"].(func(string, int64) string)

	tests := []struct {
		name     string
		tz       string
		expected string
	}{
		{"utc_renders_z", "UTC", "2023-11-14T22:13:20.000Z"},
		{"tokyo_offset", "Asia/Tokyo", "2023-11-15T07:13:20.000+09:00"},
		{"new_york_offset", "America/New_York", "2023-11-14T17:13:20.000-05:00"},
		{"kolkata_half_hour", "Asia/Kolkata", "2023-11-15T03:43:20.000+05:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatISO(tt.tz, ts))
		})
	}
}

// An empty or unknown zone must fall back to the server's local zone rendered
// with a real marker, matching the semantics day.Format established.
func TestTimeFormatISOFallsBackToLocalForUnknownZone(t *testing.T) {
	const ts = int64(1700000000)
	formatISO := funcMap["timeFormatISO"].(func(string, int64) string)
	expected := time.Unix(ts, 0).In(time.Local).Format("2006-01-02T15:04:05.000Z07:00")

	for _, tz := range []string{"", "Not/AZone", "UTC+99"} {
		t.Run("tz="+tz, func(t *testing.T) {
			assert.Equal(t, expected, formatISO(tz, ts))
		})
	}
}

// The long-date/short-date choice depends on the calendar year of the
// rendered zone: two instants that share a year in Tokyo can sit on opposite
// sides of New Year in Los Angeles.
func TestSameYearInZone(t *testing.T) {
	// 2024-01-01 08:00 JST == 2023-12-31 15:00 PST
	const jstNewYear = int64(1704067200)
	// 2024-06-01 08:00 JST == 2024-05-31 23:00 UTC
	const midYear = int64(1717200000)

	assert.True(t, sameYearInZone(midYear, jstNewYear, "Asia/Tokyo"))
	assert.False(t, sameYearInZone(midYear, jstNewYear, "America/Los_Angeles"))
}
