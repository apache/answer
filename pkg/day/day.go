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
	"strings"
	"time"
)

var placeholder = map[string]string{
	"YY":   "06",      // 06	year
	"YYYY": "2006",    // 2006	full year
	"M":    "1",       // 1-12	month
	"MM":   "01",      // 01-12	month
	"MMM":  "Jan",     // Jan-Dec month
	"MMMM": "January", // January-December month
	"D":    "2",       // 1-31	date
	"DD":   "02",      // 01-31	date preset 0
	"H":    "15",      // 00-23	hour 24
	"HH":   "15",      // 00-23	hour 24
	"h":    "3",       // 1-12	hour 12
	"hh":   "03",      // 01-12	hour 12
	"m":    "4",       // 0-59	minute
	"mm":   "04",      // 00-59	minute
	"s":    "5",       // 0-59	second
	"ss":   "05",      // 00-59	second
	"A":    "PM",      // AM / PM
	"a":    "pm",      // am / pm
	"[at]": "at",      // at string
}

func Format(unix int64, format, tz string) (formatted string) {
	/*l := len(placeholders) - 1
	for i := l; i >= 0; i-- {
		format = strings.ReplaceAll(format, placeholders[i].old, placeholders[i].new)
	}*/
	var toFormat strings.Builder
	from := []rune(format)
	for len(from) > 0 {
		to, suffix := nextStdChunk(from)
		toFormat.WriteString(string(to))
		from = suffix
	}

	_, _ = time.LoadLocation(tz)
	formatted = time.Unix(unix, 0).Format(toFormat.String())
	return
}

func nextStdChunk(from []rune) (to, suffix []rune) {
	if len(from) == 0 {
		to = []rune{}
		suffix = []rune{}
		return
	}

	s := string(from[0])
	old := ""

	switch s {
	case "Y":
		if len(from) >= 4 && string(from[:4]) == "YYYY" {
			old = "YYYY"
		} else if len(from) >= 2 && string(from[:2]) == "YY" {
			old = "YY"
		}
	case "M":
		for i := 4; i > 0; i-- {
			if len(from) >= i {
				switch string(from[:i]) {
				case "MMMM":
					old = "MMMM"
				case "MMM":
					old = "MMM"
				case "MM":
					old = "MM"
				case "M":
					old = "M"
				}
			}
			if old != "" {
				break
			}
		}
	case "D":
		for i := 2; i >= 0; i-- {
			if len(from) >= i {
				switch string(from[:i]) {
				case "DD":
					old = "DD"
				case "D":
					old = "D"
				}
			}
			if old != "" {
				break
			}
		}
	case "H":
		for i := 2; i >= 0; i-- {
			if len(from) >= i {
				switch string(from[:i]) {
				case "HH":
					old = "HH"
				case "H":
					old = "H"
				}
			}
			if old != "" {
				break
			}
		}
	case "h":
		for i := 2; i >= 0; i-- {
			if len(from) >= i {
				switch string(from[:i]) {
				case "hh":
					old = "hh"
				case "h":
					old = "h"
				}
			}
			if old != "" {
				break
			}
		}
	case "m":
		for i := 2; i >= 0; i-- {
			if len(from) >= i {
				switch string(from[:i]) {
				case "mm":
					old = "mm"
				case "m":
					old = "m"
				}
			}
			if old != "" {
				break
			}
		}
	case "s":
		for i := 2; i >= 0; i-- {
			if len(from) >= i {
				switch string(from[:i]) {
				case "ss":
					old = "ss"
				case "s":
					old = "s"
				}
			}
			if old != "" {
				break
			}
		}
	case "A":
		old = "A"
	case "a":
		old = "a"
	case "[":
		if len(from) >= 4 && string(from[:4]) == "[at]" {
			old = "[at]"
		}
	default:
		old = s
	}

	tos, ok := placeholder[old]
	if !ok {
		to = []rune(old)
	} else {
		to = []rune(tos)
	}

	suffix = from[len([]rune(old)):]
	return
}
