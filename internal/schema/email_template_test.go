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

import "testing"

func TestEmailCodeContentIsSourceType(t *testing.T) {
	if PasswordResetSourceType == ConfirmNewEmailSourceType {
		t.Fatal("password reset and confirm new email source types must be distinct")
	}

	passwordResetCode := &EmailCodeContent{SourceType: PasswordResetSourceType}
	if !passwordResetCode.IsSourceType(PasswordResetSourceType) {
		t.Fatal("password reset code should match the password reset source type")
	}
	if passwordResetCode.IsSourceType(UnsubscribeSourceType, ConfirmNewEmailSourceType) {
		t.Fatal("password reset code must not match another source type")
	}

	unsubscribeCode := &EmailCodeContent{SourceType: UnsubscribeSourceType}
	if unsubscribeCode.IsSourceType(PasswordResetSourceType, AccountActivationSourceType) {
		t.Fatal("unsubscribe code must not match an account credential source type")
	}
}
