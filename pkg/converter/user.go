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
	"regexp"

	"github.com/segmentfault/pacman/utils"
)

func DeleteUserDisplay(userID string) string {
	return utils.EnShortID(StringToInt64(userID), 100)
}

func GetMentionUsernameList(text string) []string {
	re := regexp.MustCompile(`\[@([^\]]+)\]\(/users/[^\)]+\)`)
	matches := re.FindAllStringSubmatch(text, -1)

	var usernames []string
	for _, match := range matches {
		if len(match) > 1 {
			usernames = append(usernames, match[1])
		}
	}
	return usernames
}
