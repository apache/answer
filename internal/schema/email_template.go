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
	"encoding/json"
	"github.com/apache/answer/internal/base/constant"
)

const (
	AccountActivationSourceType EmailSourceType = "account-activation"
	PasswordResetSourceType     EmailSourceType = "password-reset"
	ConfirmNewEmailSourceType   EmailSourceType = "password-reset"
	UnsubscribeSourceType       EmailSourceType = "unsubscribe"
	BindingSourceType           EmailSourceType = "binding"
)

type EmailSourceType string

type EmailCodeContent struct {
	SourceType EmailSourceType `json:"source_type"`
	Email      string          `json:"e_mail"`
	UserID     string          `json:"user_id"`
	// Used for unsubscribe notification
	NotificationSources []constant.NotificationSource `json:"notification_source,omitempty"`
	// Used for third-party login account binding
	BindingKey string `json:"binding_key,omitempty"`
	// Skip the validation of the latest code
	SkipValidationLatestCode bool `json:"skip_validation_latest_code"`
}

func (r *EmailCodeContent) ToJSONString() string {
	codeBytes, _ := json.Marshal(r)
	return string(codeBytes)
}

func (r *EmailCodeContent) FromJSONString(data string) error {
	return json.Unmarshal([]byte(data), &r)
}

type RegisterTemplateData struct {
	SiteName    string
	RegisterUrl string
}

type PassResetTemplateData struct {
	SiteName     string
	PassResetUrl string
}

type ChangeEmailTemplateData struct {
	SiteName       string
	ChangeEmailUrl string
}

type TestTemplateData struct {
	SiteName string
}

type NewAnswerTemplateRawData struct {
	AnswerUserDisplayName string
	QuestionTitle         string
	QuestionID            string
	AnswerID              string
	AnswerSummary         string
	UnsubscribeCode       string
}

type NewAnswerTemplateData struct {
	SiteName       string
	DisplayName    string
	QuestionTitle  string
	AnswerUrl      string
	AnswerSummary  string
	UnsubscribeUrl string
}

type NewInviteAnswerTemplateRawData struct {
	InviterDisplayName string
	QuestionTitle      string
	QuestionID         string
	UnsubscribeCode    string
}

type NewInviteAnswerTemplateData struct {
	SiteName       string
	DisplayName    string
	QuestionTitle  string
	InviteUrl      string
	UnsubscribeUrl string
}

type NewCommentTemplateRawData struct {
	CommentUserDisplayName string
	QuestionTitle          string
	QuestionID             string
	AnswerID               string
	CommentID              string
	CommentSummary         string
	UnsubscribeCode        string
}

type NewCommentTemplateData struct {
	SiteName       string
	DisplayName    string
	QuestionTitle  string
	CommentUrl     string
	CommentSummary string
	UnsubscribeUrl string
}

type NewQuestionTemplateRawData struct {
	QuestionAuthorUserID string
	QuestionTitle        string
	QuestionID           string
	UnsubscribeCode      string
	Tags                 []string
	TagIDs               []string
}

type NewQuestionTemplateData struct {
	SiteName       string
	QuestionTitle  string
	QuestionUrl    string
	Tags           string
	UnsubscribeUrl string
}
