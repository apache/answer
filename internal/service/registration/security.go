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

package registration

import (
	"context"
	"time"
)

const (
	EmailCodeTTL                 = 10 * time.Minute
	EmailCodeVerificationLockTTL = time.Minute
	EmailCooldown                = time.Minute
	EmailHourlyWindow            = time.Hour
	EmailHourlyLimit             = int64(5)
	IPHourlyWindow               = time.Hour
	IPHourlyLimit                = int64(100)
)

// SecurityRepo stores one-time registration codes and applies shared limits.
type SecurityRepo interface {
	CheckAndRecordSendLimit(ctx context.Context, email, ip string) (retryAfter time.Duration, err error)
	SaveEmailCode(ctx context.Context, email, code string, ttl time.Duration) error
	VerifyAndLockEmailCode(ctx context.Context, email, code string) (verificationToken string, matched bool, err error)
	DeleteEmailCodeIfMatches(ctx context.Context, email, code string) (bool, error)
	ReleaseEmailCodeLock(ctx context.Context, email, verificationToken string) error
}
