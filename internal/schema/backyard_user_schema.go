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

package schema

import (
	"context"
	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/base/validator"
	"github.com/segmentfault/pacman/errors"
	"strings"
)

// UpdateUserStatusReq update user request
type UpdateUserStatusReq struct {
	UserID           string `validate:"required" json:"user_id"`
	Status           string `validate:"required,oneof=normal suspended deleted inactive" json:"status" enums:"normal,suspended,deleted,inactive"`
	RemoveAllContent bool   `validate:"omitempty" json:"remove_all_content"`
	LoginUserID      string `json:"-"`
}

func (r *UpdateUserStatusReq) IsNormal() bool    { return r.Status == constant.UserNormal }
func (r *UpdateUserStatusReq) IsSuspended() bool { return r.Status == constant.UserSuspended }
func (r *UpdateUserStatusReq) IsDeleted() bool   { return r.Status == constant.UserDeleted }
func (r *UpdateUserStatusReq) IsInactive() bool  { return r.Status == constant.UserInactive }

// GetUserPageReq get user list page request
type GetUserPageReq struct {
	// page
	Page int `validate:"omitempty,min=1" form:"page"`
	// page size
	PageSize int `validate:"omitempty,min=1" form:"page_size"`
	// email
	Query string `validate:"omitempty,gt=0,lte=100" form:"query"`
	// user status
	Status string `validate:"omitempty,oneof=normal suspended deleted inactive" form:"status"`
	// staff, if staff is true means query admin or moderator
	Staff bool `validate:"omitempty" form:"staff"`
}

func (r *GetUserPageReq) IsSuspended() bool { return r.Status == constant.UserSuspended }
func (r *GetUserPageReq) IsDeleted() bool   { return r.Status == constant.UserDeleted }
func (r *GetUserPageReq) IsInactive() bool  { return r.Status == constant.UserInactive }

// GetUserPageResp get user response
type GetUserPageResp struct {
	// user id
	UserID string `json:"user_id"`
	// create time
	CreatedAt int64 `json:"created_at"`
	// delete time
	DeletedAt int64 `json:"deleted_at"`
	// suspended time
	SuspendedAt int64 `json:"suspended_at"`
	// username
	Username string `json:"username"`
	// email
	EMail string `json:"e_mail"`
	// rank
	Rank int `json:"rank"`
	// user status(normal,suspended,deleted,inactive)
	Status string `json:"status"`
	// display name
	DisplayName string `json:"display_name"`
	// avatar
	Avatar string `json:"avatar"`
	// role id
	RoleID int `json:"role_id"`
	// role name
	RoleName string `json:"role_name"`
}

// GetUserInfoReq get user request
type GetUserInfoReq struct {
	UserID string `validate:"required" json:"user_id"`
}

// GetUserInfoResp get user response
type GetUserInfoResp struct {
}

// UpdateUserRoleReq update user role request
type UpdateUserRoleReq struct {
	// user id
	UserID string `validate:"required" json:"user_id"`
	// role id
	RoleID int `validate:"required" json:"role_id"`
	// login user id
	LoginUserID string `json:"-"`
}

// EditUserProfileReq edit user profile request
type EditUserProfileReq struct {
	UserID      string `validate:"required" json:"user_id"`
	DisplayName string `validate:"required,gte=4,lte=30" json:"display_name"`
	Username    string `validate:"omitempty,gt=3,lte=30" json:"username"`
	Email       string `validate:"required,email,gt=0,lte=500" json:"email"`
	LoginUserID string `json:"-"`
	IsAdmin     bool   `json:"-"`
}

// AddUserReq add user request
type AddUserReq struct {
	DisplayName string `validate:"required,gte=2,lte=30" json:"display_name"`
	Email       string `validate:"required,email,gt=0,lte=500" json:"email"`
	Password    string `validate:"required,gte=8,lte=32" json:"password"`
	LoginUserID string `json:"-"`
}

// AddUsersReq add users request
type AddUsersReq struct {
	// users info line by line
	UsersStr string        `json:"users"`
	Users    []*AddUserReq `json:"-"`
}

// DeletePermanentlyReq delete permanently request
type DeletePermanentlyReq struct {
	Type string `validate:"required,oneof=users questions answers" json:"type"`
}

type AddUsersErrorData struct {
	// optional. error field name.
	Field string `json:"field"`
	// must. error line number.
	Line int `json:"line"`
	// must. error content.
	Content string `json:"content"`
	// optional. error message.
	ExtraMessage string `json:"extra_message"`
}

func (e *AddUsersErrorData) GetErrField(ctx context.Context) (errFields []*validator.FormErrorField) {
	return append([]*validator.FormErrorField{}, &validator.FormErrorField{
		ErrorField: "users",
		ErrorMsg:   translator.TrWithData(handler.GetLangByCtx(ctx), reason.AddBulkUsersFormatError, e),
	})
}

func (req *AddUsersReq) ParseUsers(ctx context.Context) (errFields []*validator.FormErrorField, err error) {
	req.UsersStr = strings.TrimSpace(req.UsersStr)
	lines := strings.Split(req.UsersStr, "\n")
	req.Users = make([]*AddUserReq, 0)
	for i, line := range lines {
		arr := strings.Split(line, ",")
		if len(arr) != 3 {
			errFields = append([]*validator.FormErrorField{}, &validator.FormErrorField{
				ErrorField: "users",
				ErrorMsg: translator.TrWithData(handler.GetLangByCtx(ctx), reason.AddBulkUsersFormatError,
					&AddUsersErrorData{
						Line:    i + 1,
						Content: line,
					}),
			})
			return errFields, errors.BadRequest(reason.RequestFormatError)
		}
		req.Users = append(req.Users, &AddUserReq{
			DisplayName: strings.TrimSpace(arr[0]),
			Email:       strings.TrimSpace(arr[1]),
			Password:    strings.TrimSpace(arr[2]),
		})
	}

	// check users amount
	if len(req.Users) <= 0 || len(req.Users) > constant.DefaultBulkUser {
		errFields = append([]*validator.FormErrorField{}, &validator.FormErrorField{
			ErrorField: "users",
			ErrorMsg: translator.TrWithData(handler.GetLangByCtx(ctx), reason.AddBulkUsersAmountError,
				map[string]int{
					"MaxAmount": constant.DefaultBulkUser,
				}),
		})
		return errFields, errors.BadRequest(reason.RequestFormatError)
	}
	return nil, nil
}

// UpdateUserPasswordReq update user password request
type UpdateUserPasswordReq struct {
	UserID      string `validate:"required" json:"user_id"`
	Password    string `validate:"required,gte=8,lte=32" json:"password"`
	LoginUserID string `json:"-"`
}

// GetUserActivationReq get user activation
type GetUserActivationReq struct {
	UserID string `validate:"required" form:"user_id"`
}

// GetUserActivationResp get user activation
type GetUserActivationResp struct {
	ActivationURL string `json:"activation_url"`
}

// SendUserActivationReq send user activation
type SendUserActivationReq struct {
	UserID string `validate:"required" json:"user_id"`
}
