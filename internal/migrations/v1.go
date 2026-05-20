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

package migrations

import (
	"context"

	"xorm.io/xorm"
)

func addUserLanguage(ctx context.Context, x *xorm.Engine) error {
	type User struct {
		ID       string `xorm:"not null pk autoincr BIGINT(20) id"`
		Username string `xorm:"not null default '' VARCHAR(50) UNIQUE username"`
		Language string `xorm:"not null default '' VARCHAR(100) language"`
	}
	return x.Context(ctx).Sync(new(User))
}
