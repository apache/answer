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
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/pkg/uid"
)

type ExternalNotificationMsg struct {
	ReceiverUserID string `json:"receiver_user_id"`
	ReceiverEmail  string `json:"receiver_email"`
	ReceiverLang   string `json:"receiver_lang"`

	NewAnswerTemplateRawData       *NewAnswerTemplateRawData       `json:"new_answer_template_raw_data,omitempty"`
	NewInviteAnswerTemplateRawData *NewInviteAnswerTemplateRawData `json:"new_invite_answer_template_raw_data,omitempty"`
	NewCommentTemplateRawData      *NewCommentTemplateRawData      `json:"new_comment_template_raw_data,omitempty"`
	NewQuestionTemplateRawData     *NewQuestionTemplateRawData     `json:"new_question_template_raw_data,omitempty"`
}

func CreateNewQuestionNotificationMsg(
	questionID, questionTitle, questionAuthorUserID string, tags []*entity.Tag) *ExternalNotificationMsg {
	questionID = uid.DeShortID(questionID)
	msg := &ExternalNotificationMsg{
		NewQuestionTemplateRawData: &NewQuestionTemplateRawData{
			QuestionAuthorUserID: questionAuthorUserID,
			QuestionID:           questionID,
			QuestionTitle:        questionTitle,
		},
	}
	for _, tag := range tags {
		msg.NewQuestionTemplateRawData.Tags = append(msg.NewQuestionTemplateRawData.Tags, tag.SlugName)
		msg.NewQuestionTemplateRawData.TagIDs = append(msg.NewQuestionTemplateRawData.TagIDs, tag.ID)
	}
	return msg
}
