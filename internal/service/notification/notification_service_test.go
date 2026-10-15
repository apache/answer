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
	"fmt"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/constant"
	basedata "github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/noticequeue"
	notficationcommon "github.com/apache/answer/internal/service/notification_common"
	"github.com/segmentfault/pacman/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// redDotTestNotificationRepo is an in-memory notification repo. Its ClearIDUnRead only changes
// the notification while it is still unread and reports the changed rows, the same way the
// database repo does.
type redDotTestNotificationRepo struct {
	notifications map[string]*entity.Notification
}

func (f *redDotTestNotificationRepo) AddNotification(_ context.Context, notification *entity.Notification) error {
	f.notifications[notification.ID] = notification
	return nil
}

func (f *redDotTestNotificationRepo) GetNotificationPage(_ context.Context, _ *schema.NotificationSearch) (
	[]*entity.Notification, int64, error) {
	return nil, 0, nil
}

func (f *redDotTestNotificationRepo) ClearUnRead(_ context.Context, _ string, _ int) error {
	return nil
}

func (f *redDotTestNotificationRepo) ClearIDUnRead(_ context.Context, userID string, id string) (int64, error) {
	notification, exist := f.notifications[id]
	if !exist || notification.UserID != userID || notification.IsRead != schema.NotificationNotRead {
		return 0, nil
	}
	notification.IsRead = schema.NotificationRead
	return 1, nil
}

func (f *redDotTestNotificationRepo) GetByUserIdObjectIdTypeId(_ context.Context, _, _ string, _ int) (
	*entity.Notification, bool, error) {
	return nil, false, nil
}

func (f *redDotTestNotificationRepo) UpdateNotificationContent(_ context.Context, _ *entity.Notification) error {
	return nil
}

func (f *redDotTestNotificationRepo) GetById(_ context.Context, id string) (*entity.Notification, bool, error) {
	notification, exist := f.notifications[id]
	return notification, exist, nil
}

func (f *redDotTestNotificationRepo) CountNotificationByUser(_ context.Context, _ *entity.Notification) (int64, error) {
	return 0, nil
}

func (f *redDotTestNotificationRepo) DeleteNotification(_ context.Context, _ string) error {
	return nil
}

func (f *redDotTestNotificationRepo) DeleteUserNotificationConfig(_ context.Context, _ string) error {
	return nil
}

// TestNotificationService_ClearIDUnReadDecreaseRedDotOnce checks that reading the same
// notification more than once only decreases the red dot cache once, because only the request
// that really changes the notification from unread to read may decrease it.
func TestNotificationService_ClearIDUnReadDecreaseRedDotOnce(t *testing.T) {
	ctx := context.Background()
	const (
		userID         = "red-dot-user"
		notificationID = "100"
	)

	cacheStore, cleanup, err := basedata.NewCache(&basedata.CacheConf{})
	require.NoError(t, err)
	t.Cleanup(cleanup)

	testData := &basedata.Data{Cache: cacheStore}
	repo := &redDotTestNotificationRepo{notifications: map[string]*entity.Notification{}}
	info := &entity.Notification{
		ID:       notificationID,
		UserID:   userID,
		ObjectID: "1",
		Content:  "{}",
		Type:     schema.NotificationTypeInbox,
		IsRead:   schema.NotificationNotRead,
		Status:   schema.NotificationStatusNormal,
	}
	require.NoError(t, repo.AddNotification(ctx, info))

	service := &NotificationService{
		data:             testData,
		notificationRepo: repo,
		notificationCommon: notficationcommon.NewNotificationCommon(
			testData, repo, nil, nil, nil, nil, noticequeue.NewService(), nil, nil),
	}

	// the user has 3 unread notifications and reads one of them
	redDotKey := fmt.Sprintf(constant.RedDotCacheKey, constant.NotificationTypeInbox, userID)
	require.NoError(t, cacheStore.SetInt64(ctx, redDotKey, 3, time.Hour))

	require.NoError(t, service.ClearIDUnRead(ctx, userID, notificationID))
	assertRedDotCount(t, ctx, cacheStore, redDotKey, 2)

	// a duplicate request must not decrease the red dot cache again
	require.NoError(t, service.ClearIDUnRead(ctx, userID, notificationID))
	assertRedDotCount(t, ctx, cacheStore, redDotKey, 2)

	// another user can neither read the notification nor change the red dot cache
	require.NoError(t, service.ClearIDUnRead(ctx, "another-user", notificationID))
	assertRedDotCount(t, ctx, cacheStore, redDotKey, 2)
}

func assertRedDotCount(t *testing.T, ctx context.Context, c cache.Cache, key string, expected int64) {
	count, exist, err := c.GetInt64(ctx, key)
	require.NoError(t, err)
	require.True(t, exist, "red dot cache key should exist")
	assert.Equal(t, expected, count)
}
