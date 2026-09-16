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

package admin_message

import (
	"context"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/admin_message"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
)

type adminMessageRepo struct {
	data *data.Data
}

// NewAdminMessageRepo new repository; the table is created on start-up
func NewAdminMessageRepo(data *data.Data) admin_message.AdminMessageRepo {
	if err := data.DB.Sync2(new(entity.AdminMessage)); err != nil {
		log.Errorf("admin_message: sync table: %v", err)
	}
	return &adminMessageRepo{data: data}
}

// Add inserts a message and fills its ID
func (r *adminMessageRepo) Add(ctx context.Context, m *entity.AdminMessage) error {
	if _, err := r.data.DB.Context(ctx).Insert(m); err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// Page newest first with filters; userIDs restricts sender or receiver, textUserIDs / text are the free-text search
func (r *adminMessageRepo) Page(ctx context.Context, from, to time.Time, userIDs []string, text string, textUserIDs []string,
	page, pageSize int) (rows []*entity.AdminMessage, total int64, err error) {
	cond := builder.NewCond()
	if !from.IsZero() {
		cond = cond.And(builder.Gte{"created_at": from})
	}
	if !to.IsZero() {
		cond = cond.And(builder.Lt{"created_at": to})
	}
	if len(userIDs) > 0 {
		cond = cond.And(builder.Or(builder.In("sender_user_id", userIDs), builder.In("receiver_user_id", userIDs)))
	}
	if text != "" {
		like := "%" + text + "%"
		tc := builder.Or(builder.Like{"title", like}, builder.Like{"body", like})
		if len(textUserIDs) > 0 {
			tc = tc.Or(builder.In("sender_user_id", textUserIDs), builder.In("receiver_user_id", textUserIDs))
		}
		cond = cond.And(tc)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	rows = make([]*entity.AdminMessage, 0)
	total, err = r.data.DB.Context(ctx).Where(cond).Desc("created_at").Desc("id").Limit(pageSize, (page-1)*pageSize).FindAndCount(&rows)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SearchUserIDs users whose username / display name / e-mail contains text
func (r *adminMessageRepo) SearchUserIDs(ctx context.Context, text string, limit int) ([]string, error) {
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

// EmailsByIDs id → e-mail (the admin list shows the receiver's address)
func (r *adminMessageRepo) EmailsByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	users := make([]*entity.User, 0)
	if len(ids) == 0 {
		return map[string]string{}, nil
	}
	if err := r.data.DB.Context(ctx).Cols("id", "e_mail").In("id", ids).Find(&users); err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	out := make(map[string]string, len(users))
	for _, u := range users {
		out[u.ID] = u.EMail
	}
	return out, nil
}
