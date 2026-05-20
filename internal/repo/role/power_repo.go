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

package role

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/role"
	"github.com/segmentfault/pacman/errors"
)

// powerRepo power repository
type powerRepo struct {
	data *data.Data
}

// NewPowerRepo new repository
func NewPowerRepo(data *data.Data) role.PowerRepo {
	return &powerRepo{
		data: data,
	}
}

// GetPowerList get  list all
func (pr *powerRepo) GetPowerList(ctx context.Context, power *entity.Power) (powerList []*entity.Power, err error) {
	powerList = make([]*entity.Power, 0)
	err = pr.data.DB.Context(ctx).Find(&powerList, power)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
