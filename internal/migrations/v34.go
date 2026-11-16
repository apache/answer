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

package migrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func addRequireEmailVerification(ctx context.Context, x *xorm.Engine) error {
	loginSiteInfo := &entity.SiteInfo{}
	exist, err := x.Context(ctx).Where("type = ?", constant.SiteTypeLogin).Get(loginSiteInfo)
	if err != nil {
		return fmt.Errorf("get login config failed: %w", err)
	}
	if !exist {
		return nil
	}

	content, err := backfillRequireEmailVerification(loginSiteInfo.Content)
	if err != nil {
		return fmt.Errorf("backfill login config failed: %w", err)
	}
	loginSiteInfo.Content = content
	_, err = x.Context(ctx).ID(loginSiteInfo.ID).Cols("content").Update(loginSiteInfo)
	if err != nil {
		return fmt.Errorf("update login config failed: %w", err)
	}
	return nil
}

func backfillRequireEmailVerification(content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		content = "{}"
	}

	loginConfig := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(content), &loginConfig); err != nil {
		return "", err
	}
	if loginConfig == nil {
		loginConfig = map[string]json.RawMessage{}
	}

	requireEmailVerification, exists := loginConfig["require_email_verification"]
	// Legacy configs that predate this setting should keep the safer behavior.
	// Treat a missing or null value as requiring email verification.
	if !exists || bytes.Equal(bytes.TrimSpace(requireEmailVerification), []byte("null")) {
		loginConfig["require_email_verification"] = json.RawMessage("true")
	} else {
		var value bool
		if err := json.Unmarshal(requireEmailVerification, &value); err != nil {
			return "", err
		}
		loginConfig["require_email_verification"] = requireEmailVerification
	}

	data, err := json.Marshal(loginConfig)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
