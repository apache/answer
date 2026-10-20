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

// FollowReq follow object request
type FollowReq struct {
	// object id
	ObjectID string `validate:"required" form:"object_id" json:"object_id"`
	// is cancel
	IsCancel bool `validate:"omitempty" form:"is_cancel" json:"is_cancel"`
}

// FollowResp response object's follows and current user follow status
type FollowResp struct {
	// the followers of object
	Follows int `json:"follows"`
	// if user is followed object will be true,otherwise false
	IsFollowed bool `json:"is_followed"`
}

type FollowDTO struct {
	// object TagID
	ObjectID string
	// is cancel
	IsCancel bool
	// user TagID
	UserID string
	// whether the requester is an administrator or moderator
	IsAdminModerator bool
}

// UpdateFollowTagsReq update user follow tags
type UpdateFollowTagsReq struct {
	// tag slug name list
	SlugNameList []string `json:"slug_name_list"`
	// user id
	UserID string `json:"-"`
}
