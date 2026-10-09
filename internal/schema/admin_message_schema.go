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

// SendAdminMessageReq admin / moderator writes to a user
type SendAdminMessageReq struct {
	Username string `validate:"required,gt=0,lte=64" json:"username"`
	Title    string `validate:"required,notblank,gte=2,lte=200" json:"title"`
	Body     string `validate:"required,notblank,gte=2,lte=4000" json:"body"`
	// ObjectID optional context: the question / answer / comment the message is about
	ObjectID    string `validate:"omitempty" json:"object_id"`
	LoginUserID string `json:"-"`
}

// AdminMessagePageReq list filters
type AdminMessagePageReq struct {
	Page        int    `validate:"omitempty,min=1" form:"page"`
	PageSize    int    `validate:"omitempty,min=1,max=200" form:"page_size"`
	From        int64  `validate:"omitempty" form:"from"`
	To          int64  `validate:"omitempty" form:"to"`
	Username    string `validate:"omitempty,lte=64" form:"username"`
	Q           string `validate:"omitempty,lte=200" form:"q"`
	LoginUserID string `json:"-"`
}

// AdminMessageUser sender / receiver
type AdminMessageUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Email       string `json:"email,omitempty"`
	Status      string `json:"status"`
}

// AdminMessageItem one row of the admin list
type AdminMessageItem struct {
	ID          string            `json:"id"`
	CreatedAt   int64             `json:"created_at"`
	ReadAt      int64             `json:"read_at"` // 0 = unread
	Sender      *AdminMessageUser `json:"sender"`
	Receiver    *AdminMessageUser `json:"receiver"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	ObjectType  string            `json:"object_type,omitempty"`
	ObjectID    string            `json:"object_id,omitempty"`
	QuestionID  string            `json:"question_id,omitempty"`
	AnswerID    string            `json:"answer_id,omitempty"`
	ObjectTitle string            `json:"object_title,omitempty"`
	UrlTitle    string            `json:"url_title,omitempty"`
}
