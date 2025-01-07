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
	"fmt"

	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func addReview(ctx context.Context, x *xorm.Engine) error {
	c := &entity.Config{Key: "reason.not_clarity", Value: `{"name":"needs details or clarity","description":"This question currently includes multiple questions in one. It should focus on one problem only."}`}
	if _, err := x.Context(ctx).Update(c, &entity.Config{Key: "reason.not_clarity"}); err != nil {
		log.Errorf("update %+v config failed: %s", c, err)
		return fmt.Errorf("update config failed: %w", err)
	}
	return x.Context(ctx).Sync(new(entity.Review))
}
