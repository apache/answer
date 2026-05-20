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

package constant

const (
	DefaultPageSize = 20 // Default number of pages
	DefaultBulkUser = 5000
)

var (
	Version   = ""
	Revision  = ""
	GoVersion = ""
)

var Timezones = []string{
	// Americas
	"America/New_York",
	"America/Chicago",
	"America/Los_Angeles",
	"America/Toronto",
	"America/Vancouver",
	"America/Mexico_City",
	"America/Sao_Paulo",
	"America/Buenos_Aires",

	// Europe
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Madrid",
	"Europe/Rome",
	"Europe/Moscow",

	// Asia
	"Asia/Shanghai",
	"Asia/Tokyo",
	"Asia/Singapore",
	"Asia/Dubai",
	"Asia/Hong_Kong",
	"Asia/Seoul",
	"Asia/Bangkok",
	"Asia/Kolkata",

	// Pacific
	"Australia/Sydney",
	"Australia/Melbourne",
	"Pacific/Auckland",

	// Africa
	"Africa/Cairo",
	"Africa/Johannesburg",
	"Africa/Lagos",
}
