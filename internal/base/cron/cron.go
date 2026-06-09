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
	"github.com/apache/answer/internal/telemetry"
	"github.com/robfig/cron/v3"
	"github.com/segmentfault/pacman/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
		ctx, span := startCronSpan("sitemap_generation", "0 */1 * * *")
		defer span.End()
		log.Infof("sitemap cron execution")
		s.questionService.SitemapCron(ctx)
	})
	if err != nil {
		log.Error(err)
	}

	_, err = c.AddFunc("0 */1 * * *", func() {
		ctx, span := startCronSpan("hottest_refresh", "0 */1 * * *")
		defer span.End()
		log.Infof("refresh hottest cron execution")
		s.questionService.RefreshHottestCron(ctx)
	})
	if err != nil {
		log.Error(err)
	}

	_, err = c.AddFunc("*/10 * * * *", func() {
		ctx, span := startCronSpan("user_deactivation_check", "*/10 * * * *")
		defer span.End()
		log.Infof("checking expired user suspensions")
		if err := s.userAdminService.CheckAndUnsuspendExpiredUsers(ctx); err != nil {
			span.SetStatus(codes.Error, "")
			span.SetAttributes(attribute.String("error.type", "db_error"))
			log.Errorf("failed to check expired user suspensions: %v", err)
		}
	})
	if err != nil {
		log.Error(err)
	}

	if s.serviceConfig.CleanUpUploads {
		log.Infof("clean up uploads cron enabled")

		conf := s.serviceConfig
		orphanSchedule := fmt.Sprintf("0 */%d * * *", conf.CleanOrphanUploadsPeriodHours)
		_, err = c.AddFunc(orphanSchedule, func() {
			ctx, span := startCronSpan("orphan_uploads_cleanup", orphanSchedule)
			defer span.End()
			log.Infof("clean orphan upload files cron execution")
			s.fileRecordService.CleanOrphanUploadFiles(ctx)
		})
		if err != nil {
			log.Error(err)
		}

		purgeSchedule := fmt.Sprintf("0 0 */%d * *", conf.PurgeDeletedFilesPeriodDays)
		_, err = c.AddFunc(purgeSchedule, func() {
			ctx, span := startCronSpan("purge_deleted_files", purgeSchedule)
			defer span.End()
			log.Infof("purge deleted files cron execution")
			s.fileRecordService.PurgeDeletedFiles(ctx)
		})
		if err != nil {
			log.Error(err)
		}
	}
	c.Start()
}

func startCronSpan(jobName, schedule string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(telemetry.Scope).Start(context.Background(), "cron "+jobName,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	span.SetAttributes(
		attribute.String("answer.job.name", jobName),
		attribute.String("answer.job.schedule", schedule),
	)
	return ctx, span
}
