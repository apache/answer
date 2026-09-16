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

// ActivityLogPageReq admin activity log filters
type ActivityLogPageReq struct {
	Page     int `validate:"omitempty,min=1" form:"page"`
	PageSize int `validate:"omitempty,min=1,max=200" form:"page_size"`
	// From / To unix seconds (To exclusive); both optional
	From int64 `validate:"omitempty" form:"from"`
	To   int64 `validate:"omitempty" form:"to"`
	// Username filters actor or target
	Username string `validate:"omitempty,lte=64" form:"username"`
	// Action comma separated list of action keys
	Action string `validate:"omitempty,lte=1024" form:"action"`
	// Q free text
	Q string `validate:"omitempty,lte=200" form:"q"`
}

// ActivityLogUser actor / target of an entry
type ActivityLogUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
	Status      string `json:"status"`
}

// ActivityLogItem one grid row
type ActivityLogItem struct {
	ID         int64            `json:"id"`
	CreatedAt  int64            `json:"created_at"`
	Action     string           `json:"action"`
	User       *ActivityLogUser `json:"user"`
	Target     *ActivityLogUser `json:"target,omitempty"`
	ObjectType string           `json:"object_type"`
	ObjectID   string           `json:"object_id"`
	QuestionID string           `json:"question_id,omitempty"`
	AnswerID   string           `json:"answer_id,omitempty"`
	Title      string           `json:"title,omitempty"`
	UrlTitle   string           `json:"url_title,omitempty"`
	RankDelta  int              `json:"rank_delta"`
	Detail     map[string]any   `json:"detail"`
	IP         string           `json:"ip,omitempty"`
}

// ActivityLogActionCount action key + number of entries in the range
type ActivityLogActionCount struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

// ActivityLogPageViewReq page view beacon from the UI
type ActivityLogPageViewReq struct {
	Path   string `validate:"required,lte=512" json:"path"`
	Title  string `validate:"omitempty,lte=300" json:"title"`
	UserID string `json:"-"`
}
