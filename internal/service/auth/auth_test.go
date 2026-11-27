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

package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
)

func TestAuthAPIKeyDoesNotLogAccessKey(t *testing.T) {
	const accessKey = "sk_sensitive-api-key-must-not-be-logged"

	logger := &authTestLogger{}
	previousLogger := log.GetLogger()
	log.SetLogger(logger)
	t.Cleanup(func() { log.SetLogger(previousLogger) })

	service := NewAuthService(nil, &authTestAPIKeyRepo{
		key: &entity.APIKey{AccessKey: accessKey, Scope: "read-only"},
	})

	pass, err := service.AuthAPIKey(context.Background(), true, accessKey)
	if err != nil || !pass {
		t.Fatalf("read-only API key should authenticate read request: pass=%v err=%v", pass, err)
	}

	pass, err = service.AuthAPIKey(context.Background(), false, accessKey)
	if err != nil || pass {
		t.Fatalf("read-only API key should not authenticate write request: pass=%v err=%v", pass, err)
	}

	if logs := logger.String(); strings.Contains(logs, accessKey) {
		t.Fatalf("authentication logs contain API key: %s", logs)
	}
}

type authTestAPIKeyRepo struct {
	key *entity.APIKey
}

func (r *authTestAPIKeyRepo) GetAPIKeyList(context.Context) ([]*entity.APIKey, error) {
	return nil, nil
}

func (r *authTestAPIKeyRepo) GetAPIKey(context.Context, string) (*entity.APIKey, bool, error) {
	return r.key, true, nil
}

func (r *authTestAPIKeyRepo) UpdateAPIKey(context.Context, entity.APIKey) error { return nil }

func (r *authTestAPIKeyRepo) AddAPIKey(context.Context, entity.APIKey) error { return nil }

func (r *authTestAPIKeyRepo) DeleteAPIKey(context.Context, int) error { return nil }

func (r *authTestAPIKeyRepo) DeleteAPIKeysByUserID(context.Context, string) error { return nil }

type authTestLogger struct {
	entries []string
}

func (l *authTestLogger) Debug(v ...any) { l.entries = append(l.entries, fmt.Sprint(v...)) }
func (l *authTestLogger) Debugf(format string, v ...any) {
	l.entries = append(l.entries, fmt.Sprintf(format, v...))
}
func (l *authTestLogger) Info(v ...any) { l.entries = append(l.entries, fmt.Sprint(v...)) }
func (l *authTestLogger) Infof(format string, v ...any) {
	l.entries = append(l.entries, fmt.Sprintf(format, v...))
}
func (l *authTestLogger) Warn(v ...any) { l.entries = append(l.entries, fmt.Sprint(v...)) }
func (l *authTestLogger) Warnf(format string, v ...any) {
	l.entries = append(l.entries, fmt.Sprintf(format, v...))
}
func (l *authTestLogger) Error(v ...any) { l.entries = append(l.entries, fmt.Sprint(v...)) }
func (l *authTestLogger) Errorf(format string, v ...any) {
	l.entries = append(l.entries, fmt.Sprintf(format, v...))
}
func (l *authTestLogger) String() string { return strings.Join(l.entries, "\n") }
