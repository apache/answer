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
	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/user_admin"
	"github.com/apache/answer/plugin"
	"github.com/gin-gonic/gin"
	"github.com/segmentfault/pacman/errors"
)

// UserAdminController user controller
type UserAdminController struct {
	userService *user_admin.UserAdminService
}

// NewUserAdminController new controller
func NewUserAdminController(userService *user_admin.UserAdminService) *UserAdminController {
	return &UserAdminController{userService: userService}
}

// UpdateUserStatus update user
// @Summary update user
// @Description update user
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.UpdateUserStatusReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user/status [put]
func (uc *UserAdminController) UpdateUserStatus(ctx *gin.Context) {
	if u, ok := plugin.GetUserCenter(); ok && u.Description().UserStatusAgentEnabled {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		return
	}
	req := &schema.UpdateUserStatusReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := uc.userService.UpdateUserStatus(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// UpdateUserRole update user role
// @Summary update user role
// @Description update user role
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.UpdateUserRoleReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user/role [put]
func (uc *UserAdminController) UpdateUserRole(ctx *gin.Context) {
	req := &schema.UpdateUserRoleReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := uc.userService.UpdateUserRole(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AddUser add user
// @Summary add user
// @Description add user
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.AddUserReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user [post]
func (uc *UserAdminController) AddUser(ctx *gin.Context) {
	req := &schema.AddUserReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := uc.userService.AddUser(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// AddUsers add users
// @Summary add users
// @Description add users
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.AddUsersReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/users [post]
func (uc *UserAdminController) AddUsers(ctx *gin.Context) {
	req := &schema.AddUsersReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := uc.userService.AddUsers(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// UpdateUserPassword update user password
// @Summary update user password
// @Description update user password
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.UpdateUserPasswordReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user/password [put]
func (uc *UserAdminController) UpdateUserPassword(ctx *gin.Context) {
	req := &schema.UpdateUserPasswordReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.LoginUserID = middleware.GetLoginUserIDFromContext(ctx)

	err := uc.userService.UpdateUserPassword(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// EditUserProfile edit user profile
// @Summary edit user profile
// @Description edit user profile
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.EditUserProfileReq true "user"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/user/profile [put]
func (uc *UserAdminController) EditUserProfile(ctx *gin.Context) {
	req := &schema.EditUserProfileReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	req.IsAdmin = middleware.GetUserIsAdminModerator(ctx)
	if !req.IsAdmin {
		handler.HandleResponse(ctx, errors.Forbidden(reason.ForbiddenError), nil)
		return
	}

	errFields, err := uc.userService.EditUserProfile(ctx, req)
	for _, field := range errFields {
		field.ErrorMsg = translator.Tr(handler.GetLang(ctx), field.ErrorMsg)
	}
	handler.HandleResponse(ctx, err, errFields)
}

// GetUserPage get user page
// @Summary get user page
// @Description get user page
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param page query int false "page size"
// @Param page_size query int false "page size"
// @Param query query string false "search query: email, username or id:[id]"
// @Param staff query bool false "staff user"
// @Param status query string false "user status" Enums(suspended, deleted, inactive)
// @Success 200 {object} handler.RespBody{data=pager.PageModel{records=[]schema.GetUserPageResp}}
// @Router /answer/admin/api/users/page [get]
func (uc *UserAdminController) GetUserPage(ctx *gin.Context) {
	req := &schema.GetUserPageReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := uc.userService.GetUserPage(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// GetUserActivation get user activation
// @Summary get user activation
// @Description get user activation
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param user_id query string true "user id"
// @Success 200 {object} handler.RespBody{data=schema.GetUserActivationResp}
// @Router /answer/admin/api/user/activation [get]
func (uc *UserAdminController) GetUserActivation(ctx *gin.Context) {
	req := &schema.GetUserActivationReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	resp, err := uc.userService.GetUserActivation(ctx, req)
	handler.HandleResponse(ctx, err, resp)
}

// SendUserActivation send user activation
// @Summary send user activation
// @Description send user activation
// @Security ApiKeyAuth
// @Tags admin
// @Produce json
// @Param data body schema.SendUserActivationReq true "SendUserActivationReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/users/activation [post]
func (uc *UserAdminController) SendUserActivation(ctx *gin.Context) {
	req := &schema.SendUserActivationReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	err := uc.userService.SendUserActivation(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}

// DeletePermanently delete permanently
// @Summary delete permanently
// @Description delete permanently
// @Security ApiKeyAuth
// @Tags admin
// @Accept json
// @Produce json
// @Param data body schema.DeletePermanentlyReq true "DeletePermanentlyReq"
// @Success 200 {object} handler.RespBody
// @Router /answer/admin/api/delete/permanently [delete]
func (uc *UserAdminController) DeletePermanently(ctx *gin.Context) {
	req := &schema.DeletePermanentlyReq{}
	if handler.BindAndCheck(ctx, req) {
		return
	}

	err := uc.userService.DeletePermanently(ctx, req)
	handler.HandleResponse(ctx, err, nil)
}
