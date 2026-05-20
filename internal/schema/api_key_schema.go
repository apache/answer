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

// GetAPIKeyReq get api key request
type GetAPIKeyReq struct {
	UserID string `json:"-"`
}

// GetAPIKeyResp get api keys response
type GetAPIKeyResp struct {
	ID          int    `json:"id"`
	AccessKey   string `json:"access_key"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	CreatedAt   int64  `json:"created_at"`
	LastUsedAt  int64  `json:"last_used_at"`
}

// AddAPIKeyReq add api key request
type AddAPIKeyReq struct {
	Description string `validate:"required,notblank,lte=150" json:"description"`
	Scope       string `validate:"required,oneof=read-only global" json:"scope"`
	UserID      string `json:"-"`
}

// AddAPIKeyResp add api key response
type AddAPIKeyResp struct {
	AccessKey string `json:"access_key"`
}

// UpdateAPIKeyReq update api key request
type UpdateAPIKeyReq struct {
	ID          int    `validate:"required" json:"id"`
	Description string `validate:"required,notblank,lte=150" json:"description"`
	UserID      string `json:"-"`
}

// DeleteAPIKeyReq delete api key request
type DeleteAPIKeyReq struct {
	ID     int    `json:"id"`
	UserID string `json:"-"`
}
