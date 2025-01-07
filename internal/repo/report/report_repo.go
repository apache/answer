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

package report

import (
	"context"

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/report_common"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
)

// reportRepo report repository
type reportRepo struct {
	data         *data.Data
	uniqueIDRepo unique.UniqueIDRepo
}

// NewReportRepo new repository
func NewReportRepo(data *data.Data, uniqueIDRepo unique.UniqueIDRepo) report_common.ReportRepo {
	return &reportRepo{
		data:         data,
		uniqueIDRepo: uniqueIDRepo,
	}
}

// AddReport add report
func (rr *reportRepo) AddReport(ctx context.Context, report *entity.Report) (err error) {
	report.ID, err = rr.uniqueIDRepo.GenUniqueIDStr(ctx, report.TableName())
	if err != nil {
		return err
	}
	_, err = rr.data.DB.Context(ctx).Insert(report)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetReportListPage get report list page
func (rr *reportRepo) GetReportListPage(ctx context.Context, dto *schema.GetReportListPageDTO) (
	reports []*entity.Report, total int64, err error) {
	cond := &entity.Report{}
	cond.Status = dto.Status
	session := rr.data.DB.Context(ctx).Desc("updated_at")
	total, err = pager.Help(dto.Page, dto.PageSize, &reports, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByID get report by ID
func (rr *reportRepo) GetByID(ctx context.Context, id string) (report *entity.Report, exist bool, err error) {
	report = &entity.Report{}
	exist, err = rr.data.DB.Context(ctx).ID(id).Get(report)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateStatus update report status by ID
func (rr *reportRepo) UpdateStatus(ctx context.Context, id string, status int) (err error) {
	_, err = rr.data.DB.Context(ctx).ID(id).Update(&entity.Report{Status: status})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (rr *reportRepo) GetReportCount(ctx context.Context) (count int64, err error) {
	list := make([]*entity.Report, 0)
	count, err = rr.data.DB.Context(ctx).Where("status =?", entity.ReportStatusPending).FindAndCount(&list)
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
