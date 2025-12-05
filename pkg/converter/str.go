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

package converter

import (
	"fmt"
	"strconv"

	"github.com/segmentfault/pacman/log"
)

func StringToInt64(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func StringToInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func IntToString(data int64) string {
	return fmt.Sprintf("%d", data)
}

// InterfaceToString converts data to string
// It will be used in template render
func InterfaceToString(data any) string {
	switch t := data.(type) {
	case int:
		i := data.(int)
		return strconv.Itoa(i)
	case int8:
		i := data.(int8)
		return strconv.Itoa(int(i))
	case int16:
		i := data.(int16)
		return strconv.Itoa(int(i))
	case int32:
		i := data.(int32)
		return string(i)
	case int64:
		i := data.(int64)
		return strconv.FormatInt(i, 10)
	case string:
		return data.(string)
	default:
		log.Warn("can't convert type:", t)
	}
	return ""
}
