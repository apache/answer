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

// GetRankPersonalWithPageReq get rank list page request
type GetRankPersonalWithPageReq struct {
	// page
	Page int `validate:"omitempty,min=1" form:"page"`
	// page size
	PageSize int `validate:"omitempty,min=1" form:"page_size"`
	// username
	Username string `validate:"omitempty,gt=0,lte=100" form:"username"`
	// user id
	UserID string `json:"-" form:"-"`
	// login user id
	LoginUserID string `json:"-" form:"-"`
	// is admin or moderator
	IsAdminModerator bool `json:"-" form:"-"`
}

// GetRankPersonalPageResp rank response
type GetRankPersonalPageResp struct {
	// create time
	CreatedAt int64 `json:"created_at"`
	// object id
	ObjectID string `json:"object_id"`
	// question id
	QuestionID string `json:"question_id"`
	// answer id
	AnswerID string `json:"answer_id"`
	// object type
	ObjectType string `json:"object_type" enums:"question,answer,tag,comment"`
	// title
	Title string `json:"title"`
	// url title
	UrlTitle string `json:"url_title"`
	// content
	Content string `json:"content"`
	// reputation
	Reputation int `json:"reputation"`
	// rank type
	RankType string `json:"rank_type"`
}
