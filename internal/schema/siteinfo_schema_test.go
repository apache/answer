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

import (
	"encoding/json"
	"testing"

	"github.com/apache/answer/internal/base/validator"
	"github.com/segmentfault/pacman/i18n"
	"github.com/stretchr/testify/require"
)

func TestSiteLoginReqRequireEmailVerificationValidation(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		expectError bool
	}{
		{
			name:        "omitted is invalid",
			payload:     `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[]}`,
			expectError: true,
		},
		{
			name:        "null is invalid",
			payload:     `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[],"require_email_verification":null}`,
			expectError: true,
		},
		{
			name:        "false is valid",
			payload:     `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[],"require_email_verification":false}`,
			expectError: false,
		},
		{
			name:        "true is valid",
			payload:     `{"allow_new_registrations":true,"allow_email_registrations":true,"allow_password_login":true,"allow_email_domains":[],"require_email_verification":true}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SiteLoginReq{}
			require.NoError(t, json.Unmarshal([]byte(tt.payload), req))

			_, err := validator.GetValidatorByLang(i18n.DefaultLanguage).Check(req)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
