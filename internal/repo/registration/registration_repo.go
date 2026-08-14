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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	registrationservice "github.com/apache/answer/internal/service/registration"
	"github.com/apache/answer/pkg/token"
	"github.com/segmentfault/pacman/errors"
)

const (
	emailCodeKeyPrefix       = "answer:register:email-code:"
	emailCodeLockKeyPrefix   = "answer:register:email-code-lock:"
	emailMinuteRateKeyPrefix = "answer:register:rate:email-minute:"
	emailHourRateKeyPrefix   = "answer:register:rate:email-hour:"
	ipHourRateKeyPrefix      = "answer:register:rate:ip-hour:"
)

type registrationRepo struct {
	data *data.Data
}

// NewRegistrationRepo creates the Redis-backed registration security repository.
func NewRegistrationRepo(dataSource *data.Data) registrationservice.SecurityRepo {
	return &registrationRepo{data: dataSource}
}

func (r *registrationRepo) CheckAndRecordSendLimit(
	ctx context.Context,
	email, ip string,
) (time.Duration, error) {
	atomicCache, ok := r.data.Cache.(data.AtomicCache)
	if !ok {
		return 0, errors.InternalServer(reason.DatabaseError).
			WithError(fmt.Errorf("registration rate limiting requires an atomic cache")).
			WithStack()
	}

	emailHash := hashValue(normalizeEmail(email))
	ipHash := hashValue(strings.TrimSpace(ip))
	rules := []data.SlidingWindowRule{
		{Key: emailMinuteRateKeyPrefix + emailHash, Limit: 1, Window: registrationservice.EmailCooldown},
		{Key: emailHourRateKeyPrefix + emailHash, Limit: registrationservice.EmailHourlyLimit, Window: registrationservice.EmailHourlyWindow},
		{Key: ipHourRateKeyPrefix + ipHash, Limit: registrationservice.IPHourlyLimit, Window: registrationservice.IPHourlyWindow},
	}

	retryAfter, err := atomicCache.CheckAndRecordSlidingWindows(ctx, token.GenerateToken(), rules)
	if err != nil {
		return 0, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return retryAfter, nil
}

func (r *registrationRepo) SaveEmailCode(
	ctx context.Context,
	email, code string,
	ttl time.Duration,
) error {
	err := r.data.Cache.SetString(ctx, emailCodeKey(email), codeDigest(email, code), ttl)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *registrationRepo) VerifyAndLockEmailCode(
	ctx context.Context,
	email, code string,
) (string, bool, error) {
	atomicCache, ok := r.data.Cache.(data.AtomicCache)
	if !ok {
		return "", false, errors.InternalServer(reason.DatabaseError).
			WithError(fmt.Errorf("registration code verification requires an atomic cache")).
			WithStack()
	}

	verificationToken := token.GenerateToken()
	locked, err := atomicCache.SetIfAbsent(
		ctx,
		emailCodeLockKey(email),
		verificationToken,
		registrationservice.EmailCodeVerificationLockTTL,
	)
	if err != nil {
		return "", false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !locked {
		return "", false, nil
	}

	storedDigest, exists, err := r.data.Cache.GetString(ctx, emailCodeKey(email))
	if err != nil {
		_ = r.ReleaseEmailCodeLock(ctx, email, verificationToken)
		return "", false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !exists || storedDigest != codeDigest(email, code) {
		if err = r.ReleaseEmailCodeLock(ctx, email, verificationToken); err != nil {
			return "", false, err
		}
		return "", false, nil
	}
	return verificationToken, true, nil
}

func (r *registrationRepo) DeleteEmailCodeIfMatches(ctx context.Context, email, code string) (bool, error) {
	atomicCache, ok := r.data.Cache.(data.AtomicCache)
	if !ok {
		return false, errors.InternalServer(reason.DatabaseError).
			WithError(fmt.Errorf("registration code deletion requires an atomic cache")).
			WithStack()
	}

	matched, err := atomicCache.CompareAndDelete(ctx, emailCodeKey(email), codeDigest(email, code))
	if err != nil {
		return false, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return matched, nil
}

func (r *registrationRepo) ReleaseEmailCodeLock(
	ctx context.Context,
	email, verificationToken string,
) error {
	atomicCache, ok := r.data.Cache.(data.AtomicCache)
	if !ok {
		return errors.InternalServer(reason.DatabaseError).
			WithError(fmt.Errorf("registration code lock release requires an atomic cache")).
			WithStack()
	}

	_, err := atomicCache.CompareAndDelete(ctx, emailCodeLockKey(email), verificationToken)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func emailCodeKey(email string) string {
	return emailCodeKeyPrefix + hashValue(normalizeEmail(email))
}

func emailCodeLockKey(email string) string {
	return emailCodeLockKeyPrefix + hashValue(normalizeEmail(email))
}

func codeDigest(email, code string) string {
	return hashValue(normalizeEmail(email) + "\x00" + code)
}

func hashValue(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
