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
	"xorm.io/xorm"
)

func addCampusForumSections(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(new(entity.ForumSection), new(entity.Question)); err != nil {
		return fmt.Errorf("sync campus forum sections failed: %w", err)
	}

	sections := []*entity.ForumSection{
		{ID: 1, Slug: "career-future", Name: "前程似锦", Sort: 10, Status: 1},
		{ID: 2, Slug: "hainanu-campus", Name: "海大校园", Sort: 20, Status: 1},
		{ID: 3, Slug: "technology-life", Name: "科技生活", Sort: 30, Status: 1},
		{ID: 4, Slug: "life-information", Name: "生活信息", Sort: 40, Status: 1},
		{ID: 5, Slug: "site-management", Name: "站务管理", Sort: 50, Status: 1},
		{ID: 101, ParentID: 1, Slug: "employment-startup", Name: "就业创业", Sort: 101, Status: 1},
		{ID: 102, ParentID: 1, Slug: "study-abroad", Name: "出国留学", Sort: 102, Status: 1},
		{ID: 103, ParentID: 1, Slug: "civil-service", Name: "公考选调", Sort: 103, Status: 1},
		{ID: 104, ParentID: 1, Slug: "postgraduate", Name: "保研考研", Sort: 104, Status: 1},
		{ID: 201, ParentID: 2, Slug: "freshmen", Name: "新生专区", Sort: 201, Status: 1},
		{ID: 202, ParentID: 2, Slug: "graduation", Name: "毕业感言", Sort: 202, Status: 1},
		{ID: 203, ParentID: 2, Slug: "food-entertainment", Name: "吃喝玩乐", Sort: 203, Status: 1},
		{ID: 204, ParentID: 2, Slug: "campus-matchmaking", Name: "海大鹊桥", Sort: 204, Status: 1},
		{ID: 205, ParentID: 2, Slug: "campus-hotspot", Name: "校园热点", Sort: 205, Status: 1},
		{ID: 301, ParentID: 3, Slug: "exams", Name: "考试专区", Sort: 301, Status: 1},
		{ID: 302, ParentID: 3, Slug: "academic-exchange", Name: "学术交流", Sort: 302, Status: 1},
		{ID: 303, ParentID: 3, Slug: "programming", Name: "程序之家", Sort: 303, Status: 1},
		{ID: 304, ParentID: 3, Slug: "competitions", Name: "竞赛专区", Sort: 304, Status: 1},
		{ID: 401, ParentID: 4, Slug: "second-hand", Name: "二手专区", Sort: 401, Status: 1},
		{ID: 402, ParentID: 4, Slug: "lost-found", Name: "失物招领", Sort: 402, Status: 1},
		{ID: 403, ParentID: 4, Slug: "carpool", Name: "拼车同行", Sort: 403, Status: 1},
		{ID: 404, ParentID: 4, Slug: "part-time-jobs", Name: "兼职信息中心", Sort: 404, Status: 1},
		{ID: 501, ParentID: 5, Slug: "site-announcements", Name: "站务公告", Sort: 501, AdminOnly: true, Status: 1},
	}
	for _, section := range sections {
		exists, err := x.Context(ctx).ID(section.ID).Exist(new(entity.ForumSection))
		if err != nil {
			return err
		}
		if !exists {
			if _, err = x.Context(ctx).Insert(section); err != nil {
				return fmt.Errorf("insert forum section %s: %w", section.Slug, err)
			}
		}
	}
	return nil
}
