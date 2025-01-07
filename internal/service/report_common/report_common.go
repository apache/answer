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

package report_common

import (
	"context"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

// ReportRepo report repository
type ReportRepo interface {
	AddReport(ctx context.Context, report *entity.Report) (err error)
	GetReportListPage(ctx context.Context, query *schema.GetReportListPageDTO) (
		reports []*entity.Report, total int64, err error)
	GetByID(ctx context.Context, id string) (report *entity.Report, exist bool, err error)
	UpdateStatus(ctx context.Context, id string, status int) (err error)
	GetReportCount(ctx context.Context) (count int64, err error)
}
