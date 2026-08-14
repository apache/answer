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

package data

import (
	"context"
	"time"
)

// SlidingWindowRule describes one reusable Redis sliding-window limit.
type SlidingWindowRule struct {
	Key    string
	Limit  int64
	Window time.Duration
}

// AtomicCache exposes cache operations that must remain atomic across app instances.
type AtomicCache interface {
	SetIfAbsent(ctx context.Context, key, value string, ttl time.Duration) (stored bool, err error)
	CheckAndRecordSlidingWindows(
		ctx context.Context,
		member string,
		rules []SlidingWindowRule,
	) (retryAfter time.Duration, err error)
	CompareAndDelete(ctx context.Context, key, expected string) (matched bool, err error)
}
