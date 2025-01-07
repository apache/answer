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

func updateRolePinAndHideFeatures(ctx context.Context, x *xorm.Engine) error {
	defaultConfigTable := []*entity.Config{
		{ID: 119, Key: "question.pin", Value: `0`},
		{ID: 120, Key: "question.unpin", Value: `0`},
		{ID: 121, Key: "question.show", Value: `0`},
		{ID: 122, Key: "question.hide", Value: `0`},
		{ID: 123, Key: "rank.question.pin", Value: `-1`},
		{ID: 124, Key: "rank.question.unpin", Value: `-1`},
		{ID: 125, Key: "rank.question.show", Value: `-1`},
		{ID: 126, Key: "rank.question.hide", Value: `-1`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Context(ctx).Get(&entity.Config{ID: c.ID})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			if _, err = x.Context(ctx).Update(c, &entity.Config{ID: c.ID}); err != nil {
				log.Errorf("update %+v config failed: %s", c, err)
				return fmt.Errorf("update config failed: %w", err)
			}
			continue
		}
		if _, err = x.Context(ctx).Insert(&entity.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			log.Errorf("insert %+v config failed: %s", c, err)
			return fmt.Errorf("add config failed: %w", err)
		}
	}

	return nil
}
