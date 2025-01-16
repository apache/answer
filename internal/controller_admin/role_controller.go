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

package controller_admin

import (
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/schema"
	service "github.com/apache/answer/internal/service/role"
	"github.com/gin-gonic/gin"
)

// RoleController role controller
type RoleController struct {
	roleService *service.RoleService
}

// NewRoleController new controller
func NewRoleController(roleService *service.RoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

// GetRoleList get role list
// @Summary get role list
// @Description get role list
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Success 200 {object} handler.RespBody{data=[]schema.GetRoleResp}
// @Router /answer/admin/api/roles [get]
func (rc *RoleController) GetRoleList(ctx *gin.Context) {
	req := &schema.GetRoleResp{}
	if handler.BindAndCheck(ctx, req) {
		return
	}
	resp, err := rc.roleService.GetRoleList(ctx)
	handler.HandleResponse(ctx, err, resp)
}
