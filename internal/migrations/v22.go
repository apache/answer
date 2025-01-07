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
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/unique"
	"xorm.io/xorm"
)

func addBadges(ctx context.Context, x *xorm.Engine) (err error) {
	uniqueIDRepo := unique.NewUniqueIDRepo(&data.Data{DB: x})

	err = x.Context(ctx).Sync(new(entity.Badge), new(entity.BadgeGroup), new(entity.BadgeAward))
	if err != nil {
		return fmt.Errorf("sync table failed: %w", err)
	}

	for _, badgeGroup := range defaultBadgeGroupTable {
		exist, err := x.Context(ctx).Get(&entity.BadgeGroup{ID: badgeGroup.ID})
		if err != nil {
			return err
		}
		if exist {
			_, err = x.Context(ctx).ID(badgeGroup.ID).Update(badgeGroup)
		} else {
			_, err = x.Context(ctx).Insert(badgeGroup)
		}
		if err != nil {
			return fmt.Errorf("insert badge group failed: %w", err)
		}
	}

	for _, badge := range defaultBadgeTable {
		beans := &entity.Badge{Name: badge.Name}
		exist, err := x.Context(ctx).Get(beans)
		if err != nil {
			return err
		}
		if exist {
			badge.ID = beans.ID
			_, err = x.Context(ctx).ID(beans.ID).Update(badge)
			continue
		}
		badge.ID, err = uniqueIDRepo.GenUniqueIDStr(ctx, new(entity.Badge).TableName())
		if err != nil {
			return err
		}

		if _, err := x.Context(ctx).Insert(badge); err != nil {
			return err
		}
	}
	return
}
