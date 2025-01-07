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
	"github.com/apache/answer/internal/service/role"
	"github.com/segmentfault/pacman/errors"
	"xorm.io/builder"
)

// rolePowerRelRepo rolePowerRel repository
type rolePowerRelRepo struct {
	data *data.Data
}

// NewRolePowerRelRepo new repository
func NewRolePowerRelRepo(data *data.Data) role.RolePowerRelRepo {
	return &rolePowerRelRepo{
		data: data,
	}
}

// GetRolePowerTypeList get role power type list
func (rr *rolePowerRelRepo) GetRolePowerTypeList(ctx context.Context, roleID int) (powers []string, err error) {
	powers = make([]string, 0)
	err = rr.data.DB.Context(ctx).Table("role_power_rel").
		Cols("power_type").Where(builder.Eq{"role_id": roleID}).Find(&powers)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
