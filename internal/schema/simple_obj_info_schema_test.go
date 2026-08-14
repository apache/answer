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
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestSimpleObjectInfo_IsDeletedOrParentDeleted(t *testing.T) {
	tests := []struct {
		name string
		info *SimpleObjectInfo
		want bool
	}{
		{name: "missing object", info: nil, want: true},
		{
			name: "available question",
			info: &SimpleObjectInfo{ObjectType: constant.QuestionObjectType, QuestionStatus: entity.QuestionStatusAvailable},
			want: false,
		},
		{
			name: "deleted question",
			info: &SimpleObjectInfo{ObjectType: constant.QuestionObjectType, QuestionStatus: entity.QuestionStatusDeleted},
			want: true,
		},
		{
			name: "available answer on available question",
			info: &SimpleObjectInfo{
				ObjectType:     constant.AnswerObjectType,
				AnswerStatus:   entity.AnswerStatusAvailable,
				QuestionStatus: entity.QuestionStatusAvailable,
			},
			want: false,
		},
		{
			name: "deleted answer",
			info: &SimpleObjectInfo{
				ObjectType:     constant.AnswerObjectType,
				AnswerStatus:   entity.AnswerStatusDeleted,
				QuestionStatus: entity.QuestionStatusAvailable,
			},
			want: true,
		},
		{
			name: "answer on deleted question",
			info: &SimpleObjectInfo{
				ObjectType:     constant.AnswerObjectType,
				AnswerStatus:   entity.AnswerStatusAvailable,
				QuestionStatus: entity.QuestionStatusDeleted,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.info.IsDeletedOrParentDeleted())
		})
	}
}

func TestUserResponsesExposeForumCountAliases(t *testing.T) {
	userInfo := &entity.User{
		AnswerCount:   7,
		QuestionCount: 3,
	}

	loginResp := &UserLoginResp{}
	loginResp.ConvertFromUserEntity(userInfo)
	assert.Equal(t, 7, loginResp.CommentCount)
	assert.Equal(t, 3, loginResp.PostCount)

	currentResp := &GetCurrentLoginUserInfoResp{}
	currentResp.ConvertFromUserEntity(userInfo)
	assert.Equal(t, 7, currentResp.CommentCount)
	assert.Equal(t, 3, currentResp.PostCount)

	otherResp := &GetOtherUserInfoByUsernameResp{}
	otherResp.ConvertFromUserEntity(userInfo)
	assert.Equal(t, 7, otherResp.CommentCount)
	assert.Equal(t, 3, otherResp.PostCount)
}
