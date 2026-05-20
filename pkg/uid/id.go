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

package uid

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/snowflake"
)

// SnowFlakeID snowflake id
type SnowFlakeID struct {
	*snowflake.Node
}

var snowFlakeIDGenerator *SnowFlakeID

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	node, err := snowflake.NewNode(int64(r.Intn(1000)) + 1)
	if err != nil {
		panic(err.Error())
	}
	snowFlakeIDGenerator = &SnowFlakeID{node}
}

func ID() snowflake.ID {
	id := snowFlakeIDGenerator.Generate()
	return id
}

func IDStr12() string {
	id := snowFlakeIDGenerator.Generate()
	return id.Base58()
}

func IDStr() string {
	id := snowFlakeIDGenerator.Generate()
	return id.Base32()
}
