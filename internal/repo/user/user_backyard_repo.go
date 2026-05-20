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

package user

import (
	"context"
	"encoding/json"
	"time"

	"xorm.io/builder"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/auth"
	"github.com/apache/answer/internal/service/user_admin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// userAdminRepo user repository
type userAdminRepo struct {
	data     *data.Data
	authRepo auth.AuthRepo
}

// NewUserAdminRepo new repository
func NewUserAdminRepo(data *data.Data, authRepo auth.AuthRepo) user_admin.UserAdminRepo {
	return &userAdminRepo{
		data:     data,
		authRepo: authRepo,
	}
}

// UpdateUserStatus update user status
func (ur *userAdminRepo) UpdateUserStatus(ctx context.Context, userID string, userStatus, mailStatus int,
	email string, suspendedUntil time.Time,
) (err error) {
	cond := &entity.User{Status: userStatus, MailStatus: mailStatus, EMail: email}
	switch userStatus {
	case entity.UserStatusSuspended:
		cond.SuspendedAt = time.Now()
		cond.SuspendedUntil = suspendedUntil
	case entity.UserStatusDeleted:
		cond.DeletedAt = time.Now()
	case entity.UserStatusAvailable:
		// When restoring user status, clear suspended until time to zero
		cond.SuspendedUntil = time.Time{}
	}
	_, err = ur.data.DB.Context(ctx).ID(userID).MustCols("status", "mail_status", "e_mail", "suspended_at", "suspended_until", "deleted_at").Update(cond)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	userCacheInfo := &entity.UserCacheInfo{
		UserID:      userID,
		EmailStatus: mailStatus,
		UserStatus:  userStatus,
	}
	t, _ := json.Marshal(userCacheInfo)
	log.Infof("user change status: %s", string(t))
	err = ur.authRepo.SetUserStatus(ctx, userID, userCacheInfo)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// AddUser add user
func (ur *userAdminRepo) AddUser(ctx context.Context, user *entity.User) (err error) {
	_, err = ur.data.DB.Context(ctx).Insert(user)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// AddUsers add users
func (ur *userAdminRepo) AddUsers(ctx context.Context, users []*entity.User) (err error) {
	_, err = ur.data.DB.Context(ctx).Insert(users)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateUserPassword update user password
func (ur *userAdminRepo) UpdateUserPassword(ctx context.Context, userID string, password string) (err error) {
	_, err = ur.data.DB.Context(ctx).ID(userID).Update(&entity.User{Pass: password})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetUserInfo get user info
func (ur *userAdminRepo) GetUserInfo(ctx context.Context, userID string) (user *entity.User, exist bool, err error) {
	user = &entity.User{}
	exist, err = ur.data.DB.Context(ctx).ID(userID).Get(user)
	if err != nil {
		return nil, false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return
	}
	err = tryToDecorateUserInfoFromUserCenter(ctx, ur.data, user)
	if err != nil {
		return nil, false, err
	}
	return
}

// GetUserInfoByEmail get user info
func (ur *userAdminRepo) GetUserInfoByEmail(ctx context.Context, email string) (user *entity.User, exist bool, err error) {
	userInfo := &entity.User{}
	exist, err = ur.data.DB.Context(ctx).Where("e_mail = ?", email).
		Where("status != ?", entity.UserStatusDeleted).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if !exist {
		return
	}
	err = tryToDecorateUserInfoFromUserCenter(ctx, ur.data, user)
	if err != nil {
		return nil, false, err
	}
	return
}

// GetUserPage get user page
func (ur *userAdminRepo) GetUserPage(ctx context.Context, page, pageSize int, user *entity.User,
	usernameOrDisplayName string, isStaff bool) (users []*entity.User, total int64, err error) {
	users = make([]*entity.User, 0)
	session := ur.data.DB.Context(ctx)
	switch user.Status {
	case entity.UserStatusDeleted:
		session.Desc("`user`.deleted_at")
	case entity.UserStatusSuspended:
		session.Desc("`user`.suspended_at")
	default:
		session.Desc("`user`.created_at")
	}

	if len(usernameOrDisplayName) > 0 {
		session.And(builder.Or(
			builder.Like{"`user`.username", usernameOrDisplayName},
			builder.Like{"`user`.display_name", usernameOrDisplayName},
		))
	}
	if isStaff {
		session.Join("INNER", "user_role_rel", "`user`.id = `user_role_rel`.user_id AND `user_role_rel`.role_id > 1")
	}

	total, err = pager.Help(page, pageSize, &users, user, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	tryToDecorateUserListFromUserCenter(ctx, ur.data, users)
	return
}

// DeletePermanentlyUsers delete permanently users
func (ur *userAdminRepo) DeletePermanentlyUsers(ctx context.Context) (err error) {
	_, err = ur.data.DB.Context(ctx).Where("status = ?", entity.UserStatusDeleted).Delete(&entity.User{})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetExpiredSuspendedUsers gets all suspended users whose suspension has expired
func (ur *userAdminRepo) GetExpiredSuspendedUsers(ctx context.Context) (users []*entity.User, err error) {
	users = make([]*entity.User, 0)
	now := time.Now()

	err = ur.data.DB.Context(ctx).
		Where("status = ?", entity.UserStatusSuspended).
		Where("suspended_until < ?", now).
		Find(&users)

	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	return users, nil
}
