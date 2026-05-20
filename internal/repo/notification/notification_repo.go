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

package notification

import (
	"context"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	notficationcommon "github.com/apache/answer/internal/service/notification_common"
	"github.com/apache/answer/pkg/uid"
	"github.com/segmentfault/pacman/errors"
)

// notificationRepo notification repository
type notificationRepo struct {
	data *data.Data
}

// NewNotificationRepo new repository
func NewNotificationRepo(data *data.Data) notficationcommon.NotificationRepo {
	return &notificationRepo{
		data: data,
	}
}

// AddNotification add notification
func (nr *notificationRepo) AddNotification(ctx context.Context, notification *entity.Notification) (err error) {
	notification.ObjectID = uid.DeShortID(notification.ObjectID)
	_, err = nr.data.DB.Context(ctx).Insert(notification)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) UpdateNotificationContent(ctx context.Context, notification *entity.Notification) (err error) {
	now := time.Now()
	notification.UpdatedAt = now
	notification.ObjectID = uid.DeShortID(notification.ObjectID)
	_, err = nr.data.DB.Context(ctx).Where("id =?", notification.ID).Cols("content", "updated_at").Update(notification)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) ClearUnRead(ctx context.Context, userID string, notificationType int) (err error) {
	info := &entity.Notification{}
	info.IsRead = schema.NotificationRead
	_, err = nr.data.DB.Context(ctx).Where("user_id = ?", userID).And("type = ?", notificationType).Cols("is_read").Update(info)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) ClearIDUnRead(ctx context.Context, userID string, id string) (err error) {
	info := &entity.Notification{}
	info.IsRead = schema.NotificationRead
	_, err = nr.data.DB.Context(ctx).Where("user_id = ?", userID).And("id = ?", id).Cols("is_read").Update(info)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) GetById(ctx context.Context, id string) (*entity.Notification, bool, error) {
	info := &entity.Notification{}
	exist, err := nr.data.DB.Context(ctx).Where("id = ? ", id).Get(info)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return info, false, err
	}
	return info, exist, nil
}

func (nr *notificationRepo) GetByUserIdObjectIdTypeId(ctx context.Context, userID, objectID string, notificationType int) (*entity.Notification, bool, error) {
	info := &entity.Notification{}
	exist, err := nr.data.DB.Context(ctx).Where("user_id = ?", userID).And("object_id = ?", objectID).And("type = ?", notificationType).Get(info)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return info, false, err
	}
	return info, exist, nil
}

func (nr *notificationRepo) GetNotificationPage(ctx context.Context, searchCond *schema.NotificationSearch) (
	notificationList []*entity.Notification, total int64, err error) {
	notificationList = make([]*entity.Notification, 0)
	if searchCond.UserID == "" {
		return notificationList, 0, nil
	}

	session := nr.data.DB.Context(ctx)
	session = session.Desc("updated_at")

	cond := &entity.Notification{
		UserID: searchCond.UserID,
		Type:   searchCond.Type,
	}
	if searchCond.InboxType > 0 {
		cond.MsgType = searchCond.InboxType
	}
	total, err = pager.Help(searchCond.Page, searchCond.PageSize, &notificationList, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) CountNotificationByUser(ctx context.Context, cond *entity.Notification) (int64, error) {
	count, err := nr.data.DB.Context(ctx).Count(cond)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return count, err
}

func (nr *notificationRepo) DeleteNotification(ctx context.Context, userID string) (err error) {
	_, err = nr.data.DB.Context(ctx).Where("user_id = ?", userID).Delete(&entity.Notification{})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (nr *notificationRepo) DeleteUserNotificationConfig(ctx context.Context, userID string) (err error) {
	_, err = nr.data.DB.Context(ctx).Where("user_id = ?", userID).Delete(&entity.UserNotificationConfig{})
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
