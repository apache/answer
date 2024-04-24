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

package cron

import (
	"context"
	"fmt"

	"github.com/apache/incubator-answer/internal/service/content"
	"github.com/apache/incubator-answer/internal/service/siteinfo_common"
	"github.com/robfig/cron/v3"
	"github.com/segmentfault/pacman/log"
)

// ScheduledTaskManager scheduled task manager
type ScheduledTaskManager struct {
	siteInfoService siteinfo_common.SiteInfoCommonService
	questionService *content.QuestionService
}

// NewScheduledTaskManager new scheduled task manager
func NewScheduledTaskManager(
	siteInfoService siteinfo_common.SiteInfoCommonService,
	questionService *content.QuestionService,
) *ScheduledTaskManager {
	manager := &ScheduledTaskManager{
		siteInfoService: siteInfoService,
		questionService: questionService,
	}
	return manager
}

func (s *ScheduledTaskManager) Run() {
	fmt.Println("start cron")
	s.questionService.SitemapCron(context.Background())
	c := cron.New()
	_, err := c.AddFunc("0 */1 * * *", func() {
		ctx := context.Background()
		fmt.Println("sitemap cron execution")
		s.questionService.SitemapCron(ctx)
	})
	if err != nil {
		log.Error(err)
	}
	c.Start()
}
