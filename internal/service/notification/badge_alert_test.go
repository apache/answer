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

	"github.com/apache/answer/internal/base/constant"
	basedata "github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/noticequeue"
	notficationcommon "github.com/apache/answer/internal/service/notification_common"
)

// badgeAlertTestRepo a notification repository where the row behind the alert is gone
type badgeAlertTestRepo struct {
	notficationcommon.NotificationRepo
}

func (r *badgeAlertTestRepo) GetById(ctx context.Context, id string) (*entity.Notification, bool, error) {
	return &entity.Notification{}, false, nil
}

// The alert that the badge dialog reads lives in the cache. When the notification row behind it is
// gone - a revoked badge deletes it - dismissing the dialog used to return success while leaving the
// alert in place, so the dialog came back on every page load and could not be closed.
func TestClearIDUnRead_DropsTheBadgeAlertWhenTheNotificationIsGone(t *testing.T) {
	cache, cleanup, err := basedata.NewCache(&basedata.CacheConf{})
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}
	t.Cleanup(cleanup)

	data := &basedata.Data{Cache: cache}
	common := notficationcommon.NewNotificationCommon(data, nil, nil, nil, nil, nil, noticequeue.NewService(), nil, nil)
	service := &NotificationService{
		data:               data,
		notificationRepo:   &badgeAlertTestRepo{},
		notificationCommon: common,
	}

	ctx := context.TODO()
	const userID, notificationID = "10000000000000001", "10000000000000002"
	if err := common.AddBadgeAwardAlertCache(ctx, userID, notificationID, "10000000000000003"); err != nil {
		t.Fatalf("add badge award alert: %v", err)
	}

	if err := service.ClearIDUnRead(ctx, userID, notificationID); err != nil {
		t.Fatalf("clear notification: %v", err)
	}

	key := fmt.Sprintf(constant.RedDotCacheKey, constant.NotificationTypeBadgeAchievement, userID)
	if _, exist, err := cache.GetString(ctx, key); err != nil {
		t.Fatalf("read cache: %v", err)
	} else if exist {
		t.Fatal("the badge alert is still there, the dialog would come back")
	}
}
