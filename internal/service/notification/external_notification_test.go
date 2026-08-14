// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package notification

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/schema"
)

func TestExternalNotificationHandlerSkipsContentEmails(t *testing.T) {
	service := &ExternalNotificationService{}
	tests := []struct {
		name string
		msg  *schema.ExternalNotificationMsg
	}{
		{
			name: "new question",
			msg: &schema.ExternalNotificationMsg{
				NewQuestionTemplateRawData: &schema.NewQuestionTemplateRawData{},
			},
		},
		{
			name: "new answer",
			msg: &schema.ExternalNotificationMsg{
				NewAnswerTemplateRawData: &schema.NewAnswerTemplateRawData{},
			},
		},
		{
			name: "new comment",
			msg: &schema.ExternalNotificationMsg{
				NewCommentTemplateRawData: &schema.NewCommentTemplateRawData{},
			},
		},
		{
			name: "invite to answer",
			msg: &schema.ExternalNotificationMsg{
				NewInviteAnswerTemplateRawData: &schema.NewInviteAnswerTemplateRawData{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := service.Handler(context.Background(), tt.msg); err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
		})
	}
}
