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

package activity_log

import (
	"context"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/activity_log"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
)

// activityLogRepo activity log repository
type activityLogRepo struct {
	data *data.Data
}

// NewActivityLogRepo new repository. The table is created on start-up (idempotent), so the
// fork does not have to hook into upstream's numbered migrations.
func NewActivityLogRepo(data *data.Data) activity_log.ActivityLogRepo {
	if err := data.DB.Sync2(new(entity.ActivityLog)); err != nil {
		log.Errorf("activity_log: sync table: %v", err)
	}
	return &activityLogRepo{data: data}
}

// Add inserts one row
func (r *activityLogRepo) Add(ctx context.Context, entry *entity.ActivityLog) (err error) {
	_, err = r.data.DB.Context(ctx).Insert(entry)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// ExistsSourceRef reports whether a row with the given source_ref exists (dedup for back-fill / page views)
func (r *activityLogRepo) ExistsSourceRef(ctx context.Context, sourceRef string) (bool, error) {
	exist, err := r.data.DB.Context(ctx).Where("source_ref = ?", sourceRef).Exist(&entity.ActivityLog{})
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return exist, nil
}

func (r *activityLogRepo) cond(q *activity_log.Query) builder.Cond {
	cond := builder.NewCond()
	if !q.From.IsZero() {
		cond = cond.And(builder.Gte{"created_at": q.From})
	}
	if !q.To.IsZero() {
		cond = cond.And(builder.Lt{"created_at": q.To})
	}
	if len(q.UserIDs) > 0 {
		cond = cond.And(builder.Or(builder.In("user_id", q.UserIDs), builder.In("target_user_id", q.UserIDs)))
	}
	if len(q.Actions) > 0 {
		cond = cond.And(builder.In("action", q.Actions))
	}
	if q.Text != "" {
		like := "%" + q.Text + "%"
		textCond := builder.Or(builder.Like{"detail", like}, builder.Like{"action", like}, builder.Like{"ip", like})
		if len(q.TextUserIDs) > 0 {
			textCond = textCond.Or(builder.In("user_id", q.TextUserIDs), builder.In("target_user_id", q.TextUserIDs))
		}
		if q.TextObjectID != "" {
			textCond = textCond.Or(builder.Eq{"object_id": q.TextObjectID}, builder.Eq{"question_id": q.TextObjectID})
		}
		cond = cond.And(textCond)
	}
	return cond
}

// Page returns one page of rows (newest first) and the total count
func (r *activityLogRepo) Page(ctx context.Context, q *activity_log.Query, page, pageSize int) (rows []*entity.ActivityLog, total int64, err error) {
	rows = make([]*entity.ActivityLog, 0)
	session := r.data.DB.Context(ctx).Where(r.cond(q)).Desc("created_at").Desc("id")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	total, err = session.Limit(pageSize, (page-1)*pageSize).FindAndCount(&rows)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// Iterate walks matching rows newest first in batches (same order as Page); stops when fn returns false or limit is reached
func (r *activityLogRepo) Iterate(ctx context.Context, q *activity_log.Query, limit int, fn func(rows []*entity.ActivityLog) bool) error {
	const batch = 1000
	sent := 0
	for sent < limit {
		rows := make([]*entity.ActivityLog, 0, batch)
		size := batch
		if limit-sent < size {
			size = limit - sent
		}
		err := r.data.DB.Context(ctx).Where(r.cond(q)).Desc("created_at").Desc("id").Limit(size, sent).Find(&rows)
		if err != nil {
			return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		if len(rows) == 0 {
			return nil
		}
		if !fn(rows) {
			return nil
		}
		sent += len(rows)
		if len(rows) < size {
			return nil
		}
	}
	return nil
}

// ActionCounts returns action → number of rows within the date range
func (r *activityLogRepo) ActionCounts(ctx context.Context, q *activity_log.Query) (map[string]int64, error) {
	type row struct {
		Action string `xorm:"action"`
		Cnt    int64  `xorm:"cnt"`
	}
	rows := make([]*row, 0)
	dateOnly := &activity_log.Query{From: q.From, To: q.To}
	err := r.data.DB.Context(ctx).Table(new(entity.ActivityLog)).Select("action, COUNT(*) AS cnt").
		Where(r.cond(dateOnly)).GroupBy("action").Find(&rows)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	out := make(map[string]int64, len(rows))
	for _, x := range rows {
		out[x.Action] = x.Cnt
	}
	return out, nil
}

// DeletePageViewsBefore removes page.view rows older than t (retention)
func (r *activityLogRepo) DeletePageViewsBefore(ctx context.Context, t time.Time) (int64, error) {
	n, err := r.data.DB.Context(ctx).Where("action = ? AND created_at < ?", activity_log.ActionPageView, t).Delete(&entity.ActivityLog{})
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return n, nil
}

// SearchUserIDs ids of users whose username / display name / e-mail contains text (for the free-text filter)
func (r *activityLogRepo) SearchUserIDs(ctx context.Context, text string, limit int) ([]string, error) {
	like := "%" + text + "%"
	users := make([]*entity.User, 0)
	err := r.data.DB.Context(ctx).Cols("id").
		Where(builder.Or(builder.Like{"username", like}, builder.Like{"display_name", like}, builder.Like{"e_mail", like})).
		Limit(limit).Find(&users)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	ids := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids, nil
}

// ListActionsSince created_at + action only, for the dashboard's daily buckets
func (r *activityLogRepo) ListActionsSince(ctx context.Context, t time.Time) ([]*entity.ActivityLog, error) {
	rows := make([]*entity.ActivityLog, 0)
	err := r.data.DB.Context(ctx).Cols("created_at", "action").Where("created_at >= ?", t).Find(&rows)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return rows, nil
}
