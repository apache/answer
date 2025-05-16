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
	"strings"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/converter"
	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// userRepo user repository
type userRepo struct {
	data *data.Data
}

// NewUserRepo new repository
func NewUserRepo(data *data.Data) usercommon.UserRepo {
	return &userRepo{
		data: data,
	}
}

// AddUser add user
func (ur *userRepo) AddUser(ctx context.Context, user *entity.User) (err error) {
	_, err = ur.data.DB.Transaction(func(session *xorm.Session) (interface{}, error) {
		session = session.Context(ctx)
		userInfo := &entity.User{}
		exist, err := session.Where("username = ?", user.Username).Get(userInfo)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		if exist {
			return nil, errors.InternalServer(reason.UsernameDuplicate)
		}
		_, err = session.Insert(user)
		if err != nil {
			return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		}
		return nil, nil
	})
	return
}

// IncreaseAnswerCount increase answer count
func (ur *userRepo) IncreaseAnswerCount(ctx context.Context, userID string, amount int) (err error) {
	user := &entity.User{}
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userID).Incr("answer_count", amount).Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// IncreaseQuestionCount increase question count
func (ur *userRepo) IncreaseQuestionCount(ctx context.Context, userID string, amount int) (err error) {
	user := &entity.User{}
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userID).Incr("question_count", amount).Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *userRepo) UpdateQuestionCount(ctx context.Context, userID string, count int64) (err error) {
	user := &entity.User{}
	user.QuestionCount = int(count)
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userID).Cols("question_count").Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *userRepo) UpdateAnswerCount(ctx context.Context, userID string, count int) (err error) {
	user := &entity.User{}
	user.AnswerCount = count
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userID).Cols("answer_count").Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// UpdateLastLoginDate update last login date
func (ur *userRepo) UpdateLastLoginDate(ctx context.Context, userID string) (err error) {
	user := &entity.User{LastLoginDate: time.Now()}
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userID).Cols("last_login_date").Update(user)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

// UpdateEmailStatus update email status
func (ur *userRepo) UpdateEmailStatus(ctx context.Context, userID string, emailStatus int) error {
	cond := &entity.User{MailStatus: emailStatus}
	_, err := ur.data.DB.Context(ctx).Where("id = ?", userID).Cols("mail_status").Update(cond)
	if err != nil {
		return err
	}
	return nil
}

// UpdateNoticeStatus update notice status
func (ur *userRepo) UpdateNoticeStatus(ctx context.Context, userID string, noticeStatus int) error {
	cond := &entity.User{NoticeStatus: noticeStatus}
	_, err := ur.data.DB.Context(ctx).Where("id = ?", userID).Cols("notice_status").Update(cond)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *userRepo) UpdatePass(ctx context.Context, userID, pass string) error {
	_, err := ur.data.DB.Context(ctx).Where("id = ?", userID).Cols("pass").Update(&entity.User{Pass: pass})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *userRepo) UpdateEmail(ctx context.Context, userID, email string) (err error) {
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userID).Update(&entity.User{EMail: email})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ur *userRepo) UpdateUserInterface(ctx context.Context, userID, language, colorSchema string) (err error) {
	session := ur.data.DB.Context(ctx).Where("id = ?", userID)
	_, err = session.Cols("language", "color_scheme").Update(&entity.User{Language: language, ColorScheme: colorSchema})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateInfo update user info
func (ur *userRepo) UpdateInfo(ctx context.Context, userInfo *entity.User) (err error) {
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userInfo.ID).
		Cols("username", "display_name", "avatar", "bio", "bio_html", "website", "location").Update(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateUserProfile update user profile
func (ur *userRepo) UpdateUserProfile(ctx context.Context, userInfo *entity.User) (err error) {
	_, err = ur.data.DB.Context(ctx).Where("id = ?", userInfo.ID).
		Cols("username", "e_mail", "mail_status", "display_name").Update(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetByUserID get user info by user id
func (ur *userRepo) GetByUserID(ctx context.Context, userID string) (userInfo *entity.User, exist bool, err error) {
	userInfo = &entity.User{}
	exist, err = ur.data.DB.Context(ctx).Where("id = ?", userID).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	err = tryToDecorateUserInfoFromUserCenter(ctx, ur.data, userInfo)
	if err != nil {
		return nil, false, err
	}
	return
}

func (ur *userRepo) BatchGetByID(ctx context.Context, ids []string) ([]*entity.User, error) {
	list := make([]*entity.User, 0)
	err := ur.data.DB.Context(ctx).In("id", ids).Find(&list)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	tryToDecorateUserListFromUserCenter(ctx, ur.data, list)
	return list, nil
}

// GetByUsername get user by username
func (ur *userRepo) GetByUsername(ctx context.Context, username string) (userInfo *entity.User, exist bool, err error) {
	userInfo = &entity.User{}
	exist, err = ur.data.DB.Context(ctx).Where("username = ?", username).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	err = tryToDecorateUserInfoFromUserCenter(ctx, ur.data, userInfo)
	if err != nil {
		return nil, false, err
	}
	return
}

func (ur *userRepo) GetByUsernames(ctx context.Context, usernames []string) ([]*entity.User, error) {
	list := make([]*entity.User, 0)
	err := ur.data.DB.Context(ctx).Where("status =?", entity.UserStatusAvailable).In("username", usernames).Find(&list)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return list, err
	}
	tryToDecorateUserListFromUserCenter(ctx, ur.data, list)
	return list, nil
}

// GetByEmail get user by email
func (ur *userRepo) GetByEmail(ctx context.Context, email string) (userInfo *entity.User, exist bool, err error) {
	userInfo = &entity.User{}
	exist, err = ur.data.DB.Context(ctx).Where("e_mail = ?", email).
		Where("status != ?", entity.UserStatusDeleted).Get(userInfo)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (ur *userRepo) GetUserCount(ctx context.Context) (count int64, err error) {
	session := ur.data.DB.Context(ctx)
	session.Where("status = ? OR status = ?", entity.UserStatusAvailable, entity.UserStatusSuspended)
	count, err = session.Count(&entity.User{})
	if err != nil {
		return count, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, nil
}

func (ur *userRepo) SearchUserListByName(ctx context.Context, name string, limit int,
	onlyStaff bool) (userList []*entity.User, err error) {
	userList = make([]*entity.User, 0)
	session := ur.data.DB.Context(ctx)
	if onlyStaff {
		session.Join("INNER", "user_role_rel", "`user`.id = `user_role_rel`.user_id AND `user_role_rel`.role_id > 1")
	}
	session.Where("status = ?", entity.UserStatusAvailable)
	session.Where("username LIKE ? OR display_name LIKE ?", strings.ToLower(name)+"%", name+"%")
	session.OrderBy("username ASC, `user`.id DESC")
	session.Limit(limit)
	err = session.Find(&userList)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	tryToDecorateUserListFromUserCenter(ctx, ur.data, userList)
	return
}

func tryToDecorateUserInfoFromUserCenter(ctx context.Context, data *data.Data, original *entity.User) (err error) {
	if original == nil {
		return nil
	}
	uc, ok := plugin.GetUserCenter()
	if !ok {
		return nil
	}

	userInfo := &entity.UserExternalLogin{}
	session := data.DB.Context(ctx).Where("user_id = ?", original.ID)
	session.Where("provider = ?", uc.Info().SlugName)
	exist, err := session.Get(userInfo)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exist {
		return nil
	}

	userCenterBasicUserInfo, err := uc.UserInfo(userInfo.ExternalID)
	if err != nil {
		log.Error(err)
		return errors.BadRequest(reason.UserNotFound).WithError(err).WithStack()
	}

	decorateByUserCenterUser(original, userCenterBasicUserInfo)
	return nil
}

func tryToDecorateUserListFromUserCenter(ctx context.Context, data *data.Data, original []*entity.User) {
	uc, ok := plugin.GetUserCenter()
	if !ok {
		return
	}

	ids := make([]string, 0)
	originalUserIDMapping := make(map[string]*entity.User, 0)
	for _, user := range original {
		originalUserIDMapping[user.ID] = user
		ids = append(ids, user.ID)
	}

	userExternalLoginList := make([]*entity.UserExternalLogin, 0)
	session := data.DB.Context(ctx).Where("provider = ?", uc.Info().SlugName)
	session.In("user_id", ids)
	err := session.Find(&userExternalLoginList)
	if err != nil {
		log.Error(err)
		return
	}

	userExternalIDs := make([]string, 0)
	originalExternalIDMapping := make(map[string]*entity.User, 0)
	for _, u := range userExternalLoginList {
		originalExternalIDMapping[u.ExternalID] = originalUserIDMapping[u.UserID]
		userExternalIDs = append(userExternalIDs, u.ExternalID)
	}
	if len(userExternalIDs) == 0 {
		return
	}

	ucUsers, err := uc.UserList(userExternalIDs)
	if err != nil {
		log.Errorf("get user list from user center failed: %v, %v", err, userExternalIDs)
		return
	}

	for _, ucUser := range ucUsers {
		decorateByUserCenterUser(originalExternalIDMapping[ucUser.ExternalID], ucUser)
	}
}

func decorateByUserCenterUser(original *entity.User, ucUser *plugin.UserCenterBasicUserInfo) {
	if original == nil || ucUser == nil {
		return
	}
	// In general, usernames should be guaranteed unique by the User Center plugin, so there are no inconsistencies.
	if original.Username != ucUser.Username {
		log.Warnf("user %s username is inconsistent with user center", original.ID)
	}
	if len(ucUser.DisplayName) > 0 {
		original.DisplayName = ucUser.DisplayName
	}
	if len(ucUser.Email) > 0 {
		original.EMail = ucUser.Email
	}
	if len(ucUser.Avatar) > 0 {
		original.Avatar = schema.CustomAvatar(ucUser.Avatar).ToJsonString()
	}
	if len(ucUser.Mobile) > 0 {
		original.Mobile = ucUser.Mobile
	}
	if len(ucUser.Bio) > 0 {
		original.BioHTML = converter.Markdown2HTML(ucUser.Bio) + original.BioHTML
	}

	// If plugin enable rank agent, use rank from user center.
	if plugin.RankAgentEnabled() {
		original.Rank = ucUser.Rank
	}
	if ucUser.Status != plugin.UserStatusAvailable {
		original.Status = int(ucUser.Status)
	}
}

func (ur *userRepo) IsAvatarFileUsed(ctx context.Context, filePath string) (bool, error) {
	user := &entity.User{}
	count, err := ur.data.DB.Context(ctx).
		Table("user").
		Where(builder.Like{"avatar", "%" + filePath + "%"}).
		Count(&user)

	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}

	return count > 0, nil
}
