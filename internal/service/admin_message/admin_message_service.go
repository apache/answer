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

package admin_message

import (
	"context"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/noticequeue"
	"github.com/apache/answer/internal/service/object_info"
	"github.com/apache/answer/internal/service/role"
	"github.com/apache/answer/internal/service/unique"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/apache/answer/pkg/htmltext"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// AdminMessageRepo persistence
type AdminMessageRepo interface {
	Add(ctx context.Context, m *entity.AdminMessage) error
	Page(ctx context.Context, from, to time.Time, userIDs []string, text string, textUserIDs []string, page, pageSize int) ([]*entity.AdminMessage, int64, error)
	SearchUserIDs(ctx context.Context, text string, limit int) ([]string, error)
	EmailsByIDs(ctx context.Context, ids []string) (map[string]string, error)
}

// AdminMessageService messages from admins / moderators to users
type AdminMessageService struct {
	repo              AdminMessageRepo
	uniqueIDRepo      unique.UniqueIDRepo
	userCommon        *usercommon.UserCommon
	userRoleService   *role.UserRoleRelService
	objectService     *object_info.ObjService
	notificationQueue noticequeue.Service
}

// NewAdminMessageService new service
func NewAdminMessageService(
	repo AdminMessageRepo,
	uniqueIDRepo unique.UniqueIDRepo,
	userCommon *usercommon.UserCommon,
	userRoleService *role.UserRoleRelService,
	objectService *object_info.ObjService,
	notificationQueue noticequeue.Service,
) *AdminMessageService {
	return &AdminMessageService{repo: repo, uniqueIDRepo: uniqueIDRepo, userCommon: userCommon, userRoleService: userRoleService,
		objectService: objectService, notificationQueue: notificationQueue}
}

// checkStaff only admins and moderators may send / list
func (s *AdminMessageService) checkStaff(ctx context.Context, userID string) error {
	roleID, err := s.userRoleService.GetUserRole(ctx, userID)
	if err != nil {
		return err
	}
	if roleID != role.RoleAdminID && roleID != role.RoleModeratorID {
		return errors.Forbidden(reason.ForbiddenError)
	}
	return nil
}

// Send stores the message and delivers it as an inbox notification of the "messages" kind
func (s *AdminMessageService) Send(ctx context.Context, req *schema.SendAdminMessageReq) (*schema.AdminMessageItem, error) {
	if err := s.checkStaff(ctx, req.LoginUserID); err != nil {
		return nil, err
	}
	receiver, exist, err := s.userCommon.GetUserBasicInfoByUserName(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errors.BadRequest(reason.UserNotFound)
	}
	id, err := s.uniqueIDRepo.GenUniqueIDStr(ctx, constant.AdminMessageObjectType)
	if err != nil {
		return nil, err
	}
	m := &entity.AdminMessage{ID: id, SenderUserID: req.LoginUserID, ReceiverUserID: receiver.ID,
		Title: strings.TrimSpace(req.Title), Body: strings.TrimSpace(req.Body), ObjectID: "0", QuestionID: "0", AnswerID: "0"}
	if req.ObjectID != "" {
		if info, err := s.objectService.GetInfo(ctx, uid.DeShortID(req.ObjectID)); err == nil && info != nil {
			m.ObjectType, m.ObjectID = info.ObjectType, info.ObjectID
			if info.QuestionID != "" {
				m.QuestionID = info.QuestionID
			}
			if info.AnswerID != "" {
				m.AnswerID = info.AnswerID
			}
		}
	}
	if err = s.repo.Add(ctx, m); err != nil {
		return nil, err
	}
	// the notification: object = the message itself, the body travels in the object map
	s.notificationQueue.Send(ctx, &schema.NotificationMsg{
		TriggerUserID:       req.LoginUserID,
		ReceiverUserID:      receiver.ID,
		Type:                schema.NotificationTypeInbox,
		Title:               m.Title,
		ObjectID:            m.ID,
		ObjectType:          constant.AdminMessageObjectType,
		NotificationAction:  constant.NotificationAdminMessage,
		NoNeedPushAllFollow: true,
		ExtraInfo: map[string]string{"message": m.ID, "body": m.Body,
			"question": zeroToEmpty(m.QuestionID), "answer": zeroToEmpty(m.AnswerID), "context_type": m.ObjectType},
	})
	items := s.decorate(ctx, []*entity.AdminMessage{m})
	return items[0], nil
}

// Page the admin list
func (s *AdminMessageService) Page(ctx context.Context, req *schema.AdminMessagePageReq) (*pager.PageModel, error) {
	if err := s.checkStaff(ctx, req.LoginUserID); err != nil {
		return nil, err
	}
	var from, to time.Time
	if req.From > 0 {
		from = time.Unix(req.From, 0)
	}
	if req.To > 0 {
		to = time.Unix(req.To, 0)
	}
	var userIDs, textUserIDs []string
	if u := strings.TrimSpace(req.Username); u != "" {
		info, exist, err := s.userCommon.GetUserBasicInfoByUserName(ctx, u)
		if err != nil {
			return nil, err
		}
		if !exist {
			userIDs = []string{"-1"}
		} else {
			userIDs = []string{info.ID}
		}
	}
	text := strings.TrimSpace(req.Q)
	if text != "" {
		ids, err := s.repo.SearchUserIDs(ctx, text, 50)
		if err != nil {
			return nil, err
		}
		textUserIDs = ids
	}
	rows, total, err := s.repo.Page(ctx, from, to, userIDs, text, textUserIDs, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return pager.NewPageModel(total, s.decorate(ctx, rows)), nil
}

func (s *AdminMessageService) decorate(ctx context.Context, rows []*entity.AdminMessage) []*schema.AdminMessageItem {
	ids := make([]string, 0, len(rows)*2)
	seen := map[string]bool{}
	for _, r := range rows {
		for _, id := range []string{r.SenderUserID, r.ReceiverUserID} {
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	users := map[string]*schema.UserBasicInfo{}
	if len(ids) > 0 {
		if m, err := s.userCommon.BatchUserBasicInfoByID(ctx, ids); err == nil {
			users = m
		} else {
			log.Error(err)
		}
	}
	emails, err := s.repo.EmailsByIDs(ctx, ids)
	if err != nil {
		log.Error(err)
	}
	toUser := func(id string) *schema.AdminMessageUser {
		u := users[id]
		if u == nil {
			return &schema.AdminMessageUser{ID: id, DisplayName: "#" + id}
		}
		return &schema.AdminMessageUser{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName, Avatar: u.Avatar, Email: emails[id], Status: u.Status}
	}
	items := make([]*schema.AdminMessageItem, 0, len(rows))
	for _, r := range rows {
		it := &schema.AdminMessageItem{ID: r.ID, CreatedAt: r.CreatedAt.Unix(), Sender: toUser(r.SenderUserID), Receiver: toUser(r.ReceiverUserID),
			Title: r.Title, Body: r.Body, ObjectType: r.ObjectType, ObjectID: zeroToEmpty(r.ObjectID),
			QuestionID: zeroToEmpty(r.QuestionID), AnswerID: zeroToEmpty(r.AnswerID)}
		if r.ReadAt != nil {
			it.ReadAt = r.ReadAt.Unix()
		}
		if it.ObjectID != "" && s.objectService != nil {
			if info, err := s.objectService.GetInfo(ctx, it.ObjectID); err == nil && info != nil {
				it.ObjectTitle = info.Title
				it.UrlTitle = htmltext.UrlTitle(info.Title)
			}
		}
		items = append(items, it)
	}
	return items
}

func zeroToEmpty(id string) string {
	if id == "0" {
		return ""
	}
	return id
}
