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

package fake_username

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/fake_username"
	"github.com/segmentfault/pacman/errors"
)

type fakeUsernameRepo struct {
	data *data.Data
}

func NewFakeUsernameRepo(data *data.Data) fake_username.FakeUsernameRepo {
	return &fakeUsernameRepo{
		data: data,
	}
}

func (ur *fakeUsernameRepo) Add(ctx context.Context, userID, questionID string, fakeName string) (err error) {
	fakeUsername := &entity.FakeUsername{
		UserID:     userID,
		QuestionID: questionID,
		FakeName:   fakeName,
	}
	_, err = ur.data.DB.Context(ctx).Insert(fakeUsername)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *fakeUsernameRepo) GetByUserIDAndQuestionID(ctx context.Context, userID, questionID string) (
	fu *entity.FakeUsername, exist bool, err error,
) {
	fu = &entity.FakeUsername{UserID: userID, QuestionID: questionID}
	exist, err = ur.data.DB.Context(ctx).Get(fu)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (fr *fakeUsernameRepo) BatchGetByUserIDs(
	ctx context.Context, userIDs []string, questionID string,
) ([]*entity.FakeUsername, error) {
	list := make([]*entity.FakeUsername, 0)

	err := fr.data.DB.Context(ctx).In("user_id", userIDs).And("question_id = ?", questionID).Find(&list)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	return list, nil
}
