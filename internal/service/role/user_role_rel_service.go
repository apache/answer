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

	"github.com/apache/answer/internal/entity"
)

// UserRoleRelRepo userRoleRel repository
type UserRoleRelRepo interface {
	SaveUserRoleRel(ctx context.Context, userID string, roleID int) (err error)
	GetUserRoleRelList(ctx context.Context, userIDs []string) (userRoleRelList []*entity.UserRoleRel, err error)
	GetUserRoleRelListByRoleID(ctx context.Context, roleIDs []int) (
		userRoleRelList []*entity.UserRoleRel, err error)
	GetUserRoleRel(ctx context.Context, userID string) (rolePowerRel *entity.UserRoleRel, exist bool, err error)
}

// UserRoleRelService user service
type UserRoleRelService struct {
	userRoleRelRepo UserRoleRelRepo
	roleService     *RoleService
}

// NewUserRoleRelService new user role rel service
func NewUserRoleRelService(userRoleRelRepo UserRoleRelRepo, roleService *RoleService) *UserRoleRelService {
	return &UserRoleRelService{
		userRoleRelRepo: userRoleRelRepo,
		roleService:     roleService,
	}
}

// SaveUserRole save user role
func (us *UserRoleRelService) SaveUserRole(ctx context.Context, userID string, roleID int) (err error) {
	return us.userRoleRelRepo.SaveUserRoleRel(ctx, userID, roleID)
}

// GetUserRoleMapping get user role mapping
func (us *UserRoleRelService) GetUserRoleMapping(ctx context.Context, userIDs []string) (
	userRoleMapping map[string]*entity.Role, err error) {
	userRoleMapping = make(map[string]*entity.Role, 0)
	roleMapping, err := us.roleService.GetRoleMapping(ctx)
	if err != nil {
		return userRoleMapping, err
	}
	if len(roleMapping) == 0 {
		return userRoleMapping, nil
	}

	relMapping, err := us.GetUserRoleRelMapping(ctx, userIDs)
	if err != nil {
		return userRoleMapping, err
	}

	// default role is user
	defaultRole := roleMapping[1]
	for _, userID := range userIDs {
		roleID, ok := relMapping[userID]
		if !ok {
			userRoleMapping[userID] = defaultRole
			continue
		}
		userRoleMapping[userID] = roleMapping[roleID]
		if userRoleMapping[userID] == nil {
			userRoleMapping[userID] = defaultRole
		}
	}
	return userRoleMapping, nil
}

// GetUserRoleRelMapping get user role rel mapping
func (us *UserRoleRelService) GetUserRoleRelMapping(ctx context.Context, userIDs []string) (
	userRoleRelMapping map[string]int, err error) {
	userRoleRelMapping = make(map[string]int, 0)

	relList, err := us.userRoleRelRepo.GetUserRoleRelList(ctx, userIDs)
	if err != nil {
		return userRoleRelMapping, err
	}

	for _, rel := range relList {
		userRoleRelMapping[rel.UserID] = rel.RoleID
	}
	return userRoleRelMapping, nil
}

// GetUserRole get user role
func (us *UserRoleRelService) GetUserRole(ctx context.Context, userID string) (roleID int, err error) {
	rolePowerRel, exist, err := us.userRoleRelRepo.GetUserRoleRel(ctx, userID)
	if err != nil {
		return 0, err
	}
	if !exist {
		// set default role
		return 1, nil
	}
	return rolePowerRel.RoleID, nil
}

// GetUserByRoleID get user by role id
func (us *UserRoleRelService) GetUserByRoleID(ctx context.Context, roleIDs []int) (rel []*entity.UserRoleRel, err error) {
	rolePowerRels, err := us.userRoleRelRepo.GetUserRoleRelListByRoleID(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	return rolePowerRels, nil
}
