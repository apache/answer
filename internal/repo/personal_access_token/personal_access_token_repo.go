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

package personal_access_token

import (
	"context"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	personalaccesstoken "github.com/apache/answer/internal/service/personal_access_token"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

type repository struct {
	data *data.Data
}

func NewRepository(data *data.Data) personalaccesstoken.Repository {
	return &repository{data: data}
}

func (r *repository) Create(ctx context.Context, token *entity.PersonalAccessToken) error {
	_, err := r.data.DB.Context(ctx).Insert(token)
	return databaseError(err)
}

func (r *repository) FindByHash(ctx context.Context, hash string) (*entity.PersonalAccessToken, bool, error) {
	token := &entity.PersonalAccessToken{}
	exists, err := r.data.DB.Context(ctx).Where(builder.Eq{"token_hash": hash}).Get(token)
	return token, exists, databaseError(err)
}

func (r *repository) ListByUserID(ctx context.Context, userID string) ([]*entity.PersonalAccessToken, error) {
	tokens := make([]*entity.PersonalAccessToken, 0)
	err := r.data.DB.Context(ctx).Where(builder.Eq{"user_id": userID}).Desc("created_at").Find(&tokens)
	return tokens, databaseError(err)
}

func (r *repository) Revoke(ctx context.Context, userID string, id int64, revokedAt time.Time) (bool, error) {
	token := &entity.PersonalAccessToken{}
	exists, err := r.data.DB.Context(ctx).Where(builder.Eq{"id": id, "user_id": userID}).Get(token)
	if err != nil || !exists {
		return exists, databaseError(err)
	}
	if !token.RevokedAt.IsZero() {
		return true, nil
	}
	_, err = r.data.DB.Context(ctx).ID(id).Cols("revoked_at").Update(&entity.PersonalAccessToken{RevokedAt: revokedAt})
	return true, databaseError(err)
}

func (r *repository) RevokeAllByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	_, err := r.data.DB.Context(ctx).
		Where(builder.Eq{"user_id": userID}.And(builder.Eq{"revoked_at": time.Time{}})).
		Cols("revoked_at").Update(&entity.PersonalAccessToken{RevokedAt: revokedAt})
	return databaseError(err)
}

func databaseError(err error) error {
	if err == nil {
		return nil
	}
	return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
}
