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
	service "github.com/apache/answer/internal/service/role"
	"github.com/segmentfault/pacman/errors"
)

// roleRepo role repository
type roleRepo struct {
	data *data.Data
}

// NewRoleRepo new repository
func NewRoleRepo(data *data.Data) service.RoleRepo {
	return &roleRepo{
		data: data,
	}
}

// GetRoleAllList get role list all
func (rr *roleRepo) GetRoleAllList(ctx context.Context) (roleList []*entity.Role, err error) {
	roleList = make([]*entity.Role, 0)
	err = rr.data.DB.Context(ctx).Find(&roleList)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetRoleAllMapping get role all mapping
func (rr *roleRepo) GetRoleAllMapping(ctx context.Context) (roleMapping map[int]*entity.Role, err error) {
	roleList, err := rr.GetRoleAllList(ctx)
	if err != nil {
		return nil, err
	}
	roleMapping = make(map[int]*entity.Role, 0)
	for _, role := range roleList {
		roleMapping[role.ID] = role
	}
	return roleMapping, nil
}
