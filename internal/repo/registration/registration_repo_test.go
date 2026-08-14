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
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/apache/answer/internal/base/data"
	"github.com/stretchr/testify/require"
)

func newRegistrationTestRepo(t *testing.T) (*registrationRepo, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	port, err := strconv.Atoi(server.Port())
	require.NoError(t, err)
	redisCache, err := data.NewRedisCache(data.RedisCacheConf{
		Host: server.Host(), Port: port, KeyPrefix: "hnu-forum:test:",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, redisCache.Close()) })
	return &registrationRepo{data: &data.Data{Cache: redisCache}}, server
}

func TestRegistrationEmailCodeIsKeptUntilDeletedAfterSuccessfulRegistration(t *testing.T) {
	repo, _ := newRegistrationTestRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.SaveEmailCode(ctx, "Student@HAINANU.EDU.CN", "012345", time.Minute))

	verificationToken, matched, err := repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "999999")
	require.NoError(t, err)
	require.False(t, matched)
	require.Empty(t, verificationToken)

	verificationToken, matched, err = repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.True(t, matched)
	require.NotEmpty(t, verificationToken)

	// Simulate a database insertion failure: releasing the lock must retain the code.
	require.NoError(t, repo.ReleaseEmailCodeLock(ctx, "student@hainanu.edu.cn", verificationToken))

	verificationToken, matched, err = repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.True(t, matched)

	deleted, err := repo.DeleteEmailCodeIfMatches(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.True(t, deleted)
	require.NoError(t, repo.ReleaseEmailCodeLock(ctx, "student@hainanu.edu.cn", verificationToken))

	_, matched, err = repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.False(t, matched)
}

func TestRegistrationEmailCodeVerificationIsLockedPerEmail(t *testing.T) {
	repo, _ := newRegistrationTestRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.SaveEmailCode(ctx, "student@hainanu.edu.cn", "012345", time.Minute))

	verificationToken, matched, err := repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.True(t, matched)

	_, matched, err = repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.False(t, matched)

	require.NoError(t, repo.ReleaseEmailCodeLock(ctx, "student@hainanu.edu.cn", verificationToken))
}

func TestDeletingVerifiedCodeDoesNotDeleteNewerCode(t *testing.T) {
	repo, _ := newRegistrationTestRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.SaveEmailCode(ctx, "student@hainanu.edu.cn", "012345", time.Minute))

	verificationToken, matched, err := repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.True(t, matched)

	require.NoError(t, repo.SaveEmailCode(ctx, "student@hainanu.edu.cn", "654321", time.Minute))
	deleted, err := repo.DeleteEmailCodeIfMatches(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.False(t, deleted)
	require.NoError(t, repo.ReleaseEmailCodeLock(ctx, "student@hainanu.edu.cn", verificationToken))

	newToken, matched, err := repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "654321")
	require.NoError(t, err)
	require.True(t, matched)
	require.NoError(t, repo.ReleaseEmailCodeLock(ctx, "student@hainanu.edu.cn", newToken))
}

func TestRegistrationEmailCodeExpires(t *testing.T) {
	repo, server := newRegistrationTestRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.SaveEmailCode(ctx, "student@hainanu.edu.cn", "012345", time.Minute))
	server.FastForward(time.Minute + time.Second)

	_, matched, err := repo.VerifyAndLockEmailCode(ctx, "student@hainanu.edu.cn", "012345")
	require.NoError(t, err)
	require.False(t, matched)
}

func TestRegistrationSendLimitsUseEmailStrictlyAndIPAsFallback(t *testing.T) {
	repo, server := newRegistrationTestRepo(t)
	ctx := context.Background()

	retryAfter, err := repo.CheckAndRecordSendLimit(ctx, "first@hainanu.edu.cn", "203.0.113.10")
	require.NoError(t, err)
	require.Zero(t, retryAfter)

	retryAfter, err = repo.CheckAndRecordSendLimit(ctx, "first@hainanu.edu.cn", "203.0.113.10")
	require.NoError(t, err)
	require.Greater(t, retryAfter, time.Duration(0))

	// Another email sharing the same campus NAT remains allowed.
	retryAfter, err = repo.CheckAndRecordSendLimit(ctx, "second@hainanu.edu.cn", "203.0.113.10")
	require.NoError(t, err)
	require.Zero(t, retryAfter)

	// Five sends per rolling hour are allowed when the minute cooldown has elapsed.
	for send := 2; send <= 5; send++ {
		server.FastForward(time.Minute + time.Second)
		retryAfter, err = repo.CheckAndRecordSendLimit(ctx, "first@hainanu.edu.cn", "203.0.113.10")
		require.NoError(t, err)
		require.Zero(t, retryAfter)
	}
	server.FastForward(time.Minute + time.Second)
	retryAfter, err = repo.CheckAndRecordSendLimit(ctx, "first@hainanu.edu.cn", "203.0.113.10")
	require.NoError(t, err)
	require.Greater(t, retryAfter, time.Duration(0))
}
