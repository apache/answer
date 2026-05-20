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

	"github.com/apache/answer/internal/service/content"
	"github.com/apache/answer/internal/service/file_record"
	"github.com/apache/answer/internal/service/service_config"
	"github.com/apache/answer/internal/service/siteinfo_common"
	"github.com/apache/answer/internal/service/user_admin"
	"github.com/robfig/cron/v3"
	"github.com/segmentfault/pacman/log"
)

// ScheduledTaskManager scheduled task manager
type ScheduledTaskManager struct {
	siteInfoService   siteinfo_common.SiteInfoCommonService
	questionService   *content.QuestionService
	fileRecordService *file_record.FileRecordService
	userAdminService  *user_admin.UserAdminService
	serviceConfig     *service_config.ServiceConfig
}

// NewScheduledTaskManager new scheduled task manager
func NewScheduledTaskManager(
	siteInfoService siteinfo_common.SiteInfoCommonService,
	questionService *content.QuestionService,
	fileRecordService *file_record.FileRecordService,
	userAdminService *user_admin.UserAdminService,
	serviceConfig *service_config.ServiceConfig,
) *ScheduledTaskManager {
	manager := &ScheduledTaskManager{
		siteInfoService:   siteInfoService,
		questionService:   questionService,
		fileRecordService: fileRecordService,
		userAdminService:  userAdminService,
		serviceConfig:     serviceConfig,
	}
	return manager
}

func (s *ScheduledTaskManager) Run() {
	log.Infof("cron job manager start")

	s.questionService.SitemapCron(context.Background())
	c := cron.New()
	_, err := c.AddFunc("0 */1 * * *", func() {
		ctx := context.Background()
		log.Infof("sitemap cron execution")
		s.questionService.SitemapCron(ctx)
	})
	if err != nil {
		log.Error(err)
	}

	_, err = c.AddFunc("0 */1 * * *", func() {
		ctx := context.Background()
		log.Infof("refresh hottest cron execution")
		s.questionService.RefreshHottestCron(ctx)
	})
	if err != nil {
		log.Error(err)
	}

	// Check for expired user suspensions every 10 minutes
	_, err = c.AddFunc("*/10 * * * *", func() {
		ctx := context.Background()
		log.Infof("checking expired user suspensions")
		err := s.userAdminService.CheckAndUnsuspendExpiredUsers(ctx)
		if err != nil {
			log.Errorf("failed to check expired user suspensions: %v", err)
		}
	})
	if err != nil {
		log.Error(err)
	}

	if s.serviceConfig.CleanUpUploads {
		log.Infof("clean up uploads cron enabled")

		conf := s.serviceConfig
		_, err = c.AddFunc(fmt.Sprintf("0 */%d * * *", conf.CleanOrphanUploadsPeriodHours), func() {
			log.Infof("clean orphan upload files cron execution")
			s.fileRecordService.CleanOrphanUploadFiles(context.Background())
		})
		if err != nil {
			log.Error(err)
		}

		_, err = c.AddFunc(fmt.Sprintf("0 0 */%d * *", conf.PurgeDeletedFilesPeriodDays), func() {
			log.Infof("purge deleted files cron execution")
			s.fileRecordService.PurgeDeletedFiles(context.Background())
		})
		if err != nil {
			log.Error(err)
		}
	}
	c.Start()
}
