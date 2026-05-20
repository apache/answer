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

func addTagRecommendedAndReserved(ctx context.Context, x *xorm.Engine) error {
	type Tag struct {
		ID        string `xorm:"not null pk comment('tag_id') BIGINT(20) id"`
		SlugName  string `xorm:"not null default '' unique VARCHAR(35) slug_name"`
		Recommend bool   `xorm:"not null default false BOOL recommend"`
		Reserved  bool   `xorm:"not null default false BOOL reserved"`
	}
	return x.Context(ctx).Sync(new(Tag))
}
