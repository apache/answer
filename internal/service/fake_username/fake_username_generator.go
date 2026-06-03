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

package fake_username

import (
	"fmt"
	"math/rand"
)

type FakeUsernameGenerator struct{}

func NewFakeUsernameGenerator() *FakeUsernameGenerator {
	return &FakeUsernameGenerator{}
}

func (fg *FakeUsernameGenerator) GenerateFakeName() string {
	firstParts := []string{
		"Cool",
		"Dark",
		"Fast",
		"Silent",
		"Crazy",
	}

	secondParts := []string{
		"Tiger",
		"Wolf",
		"Ninja",
		"Dragon",
		"Hawk",
	}

	number := rand.Intn(10000)

	first := firstParts[rand.Intn(len(firstParts))]
	second := secondParts[rand.Intn(len(secondParts))]

	return fmt.Sprintf("%s%s%d", first, second, number)
}
