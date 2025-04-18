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

package activity_common

import "context"

type FollowRepo interface {
	GetFollowIDs(ctx context.Context, userID, objectType string) (followIDs []string, err error)
	GetFollowAmount(ctx context.Context, objectID string) (followAmount int, err error)
	GetFollowUserIDs(ctx context.Context, objectID string) (userIDs []string, err error)
	IsFollowed(ctx context.Context, userId, objectId string) (bool, error)
	MigrateFollowers(ctx context.Context, sourceObjectID, targetObjectID, action string) error
}
