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

package migrations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func addNewAnswerNotification(ctx context.Context, x *xorm.Engine) error {
	cond := &entity.Config{Key: "email.config"}
	exists, err := x.Context(ctx).Get(cond)
	if err != nil {
		return fmt.Errorf("get email config failed: %w", err)
	}
	if !exists {
		// This should be impossible except that the config was deleted manually by user.
		_, err = x.Context(ctx).Insert(&entity.Config{
			Key:   "email.config",
			Value: `{"from_name":"","from_email":"","smtp_host":"","smtp_port":465,"smtp_password":"","smtp_username":"","smtp_authentication":true,"encryption":"","register_title":"[{{.SiteName}}] Confirm your new account","register_body":"Welcome to {{.SiteName}}<br><br>\n\nClick the following link to confirm and activate your new account:<br>\n<a href='{{.RegisterUrl}}' target='_blank'>{{.RegisterUrl}}</a><br><br>\n\nIf the above link is not clickable, try copying and pasting it into the address bar of your web browser.\n","pass_reset_title":"[{{.SiteName }}] Password reset","pass_reset_body":"Somebody asked to reset your password on [{{.SiteName}}].<br><br>\n\nIf it was not you, you can safely ignore this email.<br><br>\n\nClick the following link to choose a new password:<br>\n<a href='{{.PassResetUrl}}' target='_blank'>{{.PassResetUrl}}</a>\n","change_title":"[{{.SiteName}}] Confirm your new email address","change_body":"Confirm your new email address for {{.SiteName}}  by clicking on the following link:<br><br>\n\n<a href='{{.ChangeEmailUrl}}' target='_blank'>{{.ChangeEmailUrl}}</a><br><br>\n\nIf you did not request this change, please ignore this email.\n","test_title":"[{{.SiteName}}] Test Email","test_body":"This is a test email.","new_answer_title":"[{{.SiteName}}] {{.DisplayName}} answered your question","new_answer_body":"<strong><a href='{{.AnswerUrl}}'>{{.QuestionTitle}}</a></strong><br><br>\n\n<small>{{.DisplayName}}:</small><br>\n<blockquote>{{.AnswerSummary}}</blockquote><br>\n<a href='{{.AnswerUrl}}'>View it on {{.SiteName}}</a><br><br>\n\n<small>You are receiving this because you authored the thread. <a href='{{.UnsubscribeUrl}}'>Unsubscribe</a></small>","new_comment_title":"[{{.SiteName}}] {{.DisplayName}} commented on your post","new_comment_body":"<strong><a href='{{.CommentUrl}}'>{{.QuestionTitle}}</a></strong><br><br>\n\n<small>{{.DisplayName}}:</small><br>\n<blockquote>{{.CommentSummary}}</blockquote><br>\n<a href='{{.CommentUrl}}'>View it on {{.SiteName}}</a><br><br>\n\n<small>You are receiving this because you authored the thread. <a href='{{.UnsubscribeUrl}}'>Unsubscribe</a></small>"}`,
		})
		if err != nil {
			return fmt.Errorf("add email config failed: %v", err)
		}
	}

	m := make(map[string]interface{})
	_ = json.Unmarshal([]byte(cond.Value), &m)
	m["new_answer_title"] = "[{{.SiteName}}] {{.DisplayName}} answered your question"
	m["new_answer_body"] = "<strong><a href='{{.AnswerUrl}}'>{{.QuestionTitle}}</a></strong><br><br>\n\n<small>{{.DisplayName}}:</small><br>\n<blockquote>{{.AnswerSummary}}</blockquote><br>\n<a href='{{.AnswerUrl}}'>View it on {{.SiteName}}</a><br><br>\n\n<small>You are receiving this because you authored the thread. <a href='{{.UnsubscribeUrl}}'>Unsubscribe</a></small>"
	m["new_comment_title"] = "[{{.SiteName}}] {{.DisplayName}} commented on your post"
	m["new_comment_body"] = "<strong><a href='{{.CommentUrl}}'>{{.QuestionTitle}}</a></strong><br><br>\n\n<small>{{.DisplayName}}:</small><br>\n<blockquote>{{.CommentSummary}}</blockquote><br>\n<a href='{{.CommentUrl}}'>View it on {{.SiteName}}</a><br><br>\n\n<small>You are receiving this because you authored the thread. <a href='{{.UnsubscribeUrl}}'>Unsubscribe</a></small>"

	val, _ := json.Marshal(m)
	_, err = x.Context(ctx).ID(cond.ID).Update(&entity.Config{Value: string(val)})
	if err != nil {
		return fmt.Errorf("update email config failed: %v", err)
	}
	return nil
}
