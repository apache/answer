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

package unique

import (
	"context"
	"fmt"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/unique"
	"github.com/segmentfault/pacman/errors"
)

// uniqueIDRepo Unique id repository
type uniqueIDRepo struct {
	data *data.Data
}

// NewUniqueIDRepo new repository
func NewUniqueIDRepo(data *data.Data) unique.UniqueIDRepo {
	return &uniqueIDRepo{
		data: data,
	}
}

// GenUniqueIDStr generate unique id string
// 1 + 00x(objectType) + 000000000000x(id)
func (ur *uniqueIDRepo) GenUniqueIDStr(ctx context.Context, key string) (uniqueID string, err error) {
	objectType := constant.ObjectTypeStrMapping[key]
	bean := &entity.Uniqid{UniqidType: objectType}
	_, err = ur.data.DB.Context(ctx).Insert(bean)
	if err != nil {
		return "", errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return fmt.Sprintf("1%03d%013d", objectType, bean.ID), nil
}
