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

package checker

import (
	"sync"

	"github.com/apache/answer/configs"
	"github.com/segmentfault/pacman/log"
	"gopkg.in/yaml.v3"
)

type PathIgnore struct {
	Users     []string `yaml:"users"`
	Questions []string `yaml:"questions"`
}

var (
	ignorePathInit sync.Once
	pathIgnore     = &PathIgnore{}
)

func initPathIgnore() {
	if err := yaml.Unmarshal(configs.PathIgnore, pathIgnore); err != nil {
		log.Error(err)
	}
}

// IsUsersIgnorePath checks whether the username is in ignore path
func IsUsersIgnorePath(username string) bool {
	ignorePathInit.Do(initPathIgnore)
	for _, u := range pathIgnore.Users {
		if u == username {
			return true
		}
	}
	return false
}

// IsQuestionsIgnorePath checks whether the questionID is in ignore path
func IsQuestionsIgnorePath(questionID string) bool {
	ignorePathInit.Do(initPathIgnore)
	for _, u := range pathIgnore.Questions {
		if u == questionID {
			return true
		}
	}
	return false
}
