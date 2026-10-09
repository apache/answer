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

import "time"

const (
	CGDefault = 1
	CGDIY     = 2
)

// CollectionSwitchReq switch collection request
type CollectionSwitchReq struct {
	ObjectID         string `validate:"required" json:"object_id"`
	GroupID          string `validate:"required" json:"group_id"`
	Bookmark         bool   `validate:"omitempty" json:"bookmark"`
	UserID           string `json:"-"`
	IsAdminModerator bool   `json:"-"`
}

// CollectionSwitchResp switch collection response
type CollectionSwitchResp struct {
	ObjectCollectionCount int64 `json:"object_collection_count"`
}

// AddCollectionGroupReq add collection group request
type AddCollectionGroupReq struct {
	//
	UserID int64 `validate:"required" comment:"" json:"user_id"`
	// the collection group name
	Name string `validate:"required,gt=0,lte=50" comment:"the collection group name" json:"name"`
	// mark this group is default, default 1
	DefaultGroup int `validate:"required" comment:"mark this group is default, default 1" json:"default_group"`
	//
	CreateTime time.Time `validate:"required" comment:"" json:"create_time"`
	//
	UpdateTime time.Time `validate:"required" comment:"" json:"update_time"`
}

// UpdateCollectionGroupReq update collection group request
type UpdateCollectionGroupReq struct {
	//
	ID int64 `validate:"required" comment:"" json:"id"`
	//
	UserID int64 `validate:"omitempty" comment:"" json:"user_id"`
	// the collection group name
	Name string `validate:"omitempty,gt=0,lte=50" comment:"the collection group name" json:"name"`
	// mark this group is default, default 1
	DefaultGroup int `validate:"omitempty" comment:"mark this group is default, default 1" json:"default_group"`
	//
	CreateTime time.Time `validate:"omitempty" comment:"" json:"create_time"`
	//
	UpdateTime time.Time `validate:"omitempty" comment:"" json:"update_time"`
}

// GetCollectionGroupResp get collection group response
type GetCollectionGroupResp struct {
	//
	ID int64 `json:"id"`
	//
	UserID int64 `json:"user_id"`
	// the collection group name
	Name string `json:"name"`
	// mark this group is default, default 1
	DefaultGroup int `json:"default_group"`
	//
	CreateTime time.Time `json:"create_time"`
	//
	UpdateTime time.Time `json:"update_time"`
}
