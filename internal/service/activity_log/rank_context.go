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

package activity_log

import "context"

type rankCtxKey struct{}

// RankContext tells the reputation hook which object / activity caused a rank change.
// Callers that change reputation (vote, accept, review, first-login activities) wrap ctx with WithRankContext.
type RankContext struct {
	ObjectID     string
	ActivityType int // config id of the activity (e.g. answer.voted_up), 0 when unknown
}

// WithRankContext attaches object / activity info for the reputation log entry
func WithRankContext(ctx context.Context, objectID string, activityType int) context.Context {
	return context.WithValue(ctx, rankCtxKey{}, RankContext{ObjectID: objectID, ActivityType: activityType})
}

// RankContextFrom reads what WithRankContext stored (zero value when absent)
func RankContextFrom(ctx context.Context) RankContext {
	if v, ok := ctx.Value(rankCtxKey{}).(RankContext); ok {
		return v
	}
	return RankContext{}
}
