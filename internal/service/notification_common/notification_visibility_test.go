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

package notificationcommon

import (
	"context"
	"reflect"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/role"
)

type notificationVisibilityRoleService struct{ roles map[string]int }

func (s notificationVisibilityRoleService) GetUserRoleRelMapping(context.Context, []string) (map[string]int, error) {
	return s.roles, nil
}

func TestVisibleFollowersFiltersRestrictedQuestion(t *testing.T) {
	service := &NotificationCommon{userRoleService: notificationVisibilityRoleService{roles: map[string]int{
		"moderator": role.RoleModeratorID,
	}}}
	question := &schema.SimpleObjectInfo{
		ObjectType:            constant.QuestionObjectType,
		QuestionCreatorUserID: "author",
		QuestionStatus:        entity.QuestionStatusAvailable,
		QuestionShow:          entity.QuestionHide,
	}

	visible, err := service.visibleFollowers(context.Background(), question,
		[]string{"ordinary-user", "author", "moderator"})
	if err != nil {
		t.Fatalf("visibleFollowers() error: %v", err)
	}
	if want := []string{"author", "moderator"}; !reflect.DeepEqual(visible, want) {
		t.Fatalf("visible followers = %v, want %v", visible, want)
	}
}
