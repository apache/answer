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

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/jinzhu/copier"
)

const (
	// Since there is currently no need to edit roles to add roles and other operations,
	// the current role information is translated directly.
	// Later on, when the relevant ability is available, it can be adjusted by the user himself.

	RoleUserID      = 1
	RoleAdminID     = 2
	RoleModeratorID = 3

	roleUserName      = "User"
	roleAdminName     = "Admin"
	roleModeratorName = "Moderator"

	trRoleNameUser      = "role.name.user"
	trRoleNameAdmin     = "role.name.admin"
	trRoleNameModerator = "role.name.moderator"

	trRoleDescriptionUser      = "role.description.user"
	trRoleDescriptionAdmin     = "role.description.admin"
	trRoleDescriptionModerator = "role.description.moderator"
)

// RoleRepo role repository
type RoleRepo interface {
	GetRoleAllList(ctx context.Context) (roles []*entity.Role, err error)
	GetRoleAllMapping(ctx context.Context) (roleMapping map[int]*entity.Role, err error)
}

// RoleService user service
type RoleService struct {
	roleRepo RoleRepo
}

func NewRoleService(roleRepo RoleRepo) *RoleService {
	return &RoleService{
		roleRepo: roleRepo,
	}
}

// GetRoleList get role list all
func (rs *RoleService) GetRoleList(ctx context.Context) (resp []*schema.GetRoleResp, err error) {
	roles, err := rs.roleRepo.GetRoleAllList(ctx)
	if err != nil {
		return
	}

	for _, role := range roles {
		rs.translateRole(ctx, role)
	}

	resp = []*schema.GetRoleResp{}
	_ = copier.Copy(&resp, roles)
	return
}

func (rs *RoleService) GetRoleMapping(ctx context.Context) (roleMapping map[int]*entity.Role, err error) {
	return rs.roleRepo.GetRoleAllMapping(ctx)
}

func (rs *RoleService) translateRole(ctx context.Context, role *entity.Role) {
	switch role.Name {
	case roleUserName:
		role.Name = translator.Tr(handler.GetLangByCtx(ctx), trRoleNameUser)
		role.Description = translator.Tr(handler.GetLangByCtx(ctx), trRoleDescriptionUser)
	case roleAdminName:
		role.Name = translator.Tr(handler.GetLangByCtx(ctx), trRoleNameAdmin)
		role.Description = translator.Tr(handler.GetLangByCtx(ctx), trRoleDescriptionAdmin)
	case roleModeratorName:
		role.Name = translator.Tr(handler.GetLangByCtx(ctx), trRoleNameModerator)
		role.Description = translator.Tr(handler.GetLangByCtx(ctx), trRoleDescriptionModerator)
	}
}
