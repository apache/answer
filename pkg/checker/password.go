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
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	levelD = iota
	LevelC
	LevelB
	LevelA
	LevelS
)

const (
	PasswordCannotContainSpaces = "error.password.space_invalid"
)

// CheckPassword checks the password strength
func CheckPassword(password string) error {
	if strings.Contains(password, " ") {
		return errors.New(PasswordCannotContainSpaces)
	}

	// TODO Currently there is no requirement for password strength
	minLevel := 0

	// The password strength level is initialized to D.
	// The regular is used to verify the password strength.
	// If the matching is successful, the password strength increases by 1
	level := levelD
	patternList := []string{`[0-9]+`, `[a-z]+`, `[A-Z]+`, `[~!@#$%^&*?_-]+`}
	for _, pattern := range patternList {
		match, _ := regexp.MatchString(pattern, password)
		if match {
			level++
		}
	}

	// If the final password strength falls below the required minimum strength, return with an error
	if level < minLevel {
		return fmt.Errorf("the password does not satisfy the current policy requirements")
	}
	return nil
}
