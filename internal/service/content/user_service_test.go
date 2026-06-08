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

package content

import (
	"errors"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyRegistrationVerification(t *testing.T) {
	t.Run("required sends activation email and leaves email pending", func(t *testing.T) {
		userInfo := &entity.User{}
		calls := map[string]int{}

		err := applyRegistrationVerification(userInfo, true, registrationVerificationActions{
			sendActivationEmail: func() error {
				calls["sendActivationEmail"]++
				return nil
			},
			activateUser: func() error {
				calls["activateUser"]++
				return nil
			},
			markEmailAvailable: func() error {
				calls["markEmailAvailable"]++
				return nil
			},
		})

		require.NoError(t, err)
		assert.Equal(t, entity.EmailStatusToBeVerified, userInfo.MailStatus)
		assert.Equal(t, 1, calls["sendActivationEmail"])
		assert.Zero(t, calls["activateUser"])
		assert.Zero(t, calls["markEmailAvailable"])
	})

	t.Run("not required activates once and marks email available", func(t *testing.T) {
		userInfo := &entity.User{}
		calls := map[string]int{}

		err := applyRegistrationVerification(userInfo, false, registrationVerificationActions{
			sendActivationEmail: func() error {
				calls["sendActivationEmail"]++
				return nil
			},
			activateUser: func() error {
				calls["activateUser"]++
				return nil
			},
			markEmailAvailable: func() error {
				calls["markEmailAvailable"]++
				return nil
			},
		})

		require.NoError(t, err)
		assert.Equal(t, entity.EmailStatusAvailable, userInfo.MailStatus)
		assert.Zero(t, calls["sendActivationEmail"])
		assert.Equal(t, 1, calls["activateUser"])
		assert.Equal(t, 1, calls["markEmailAvailable"])
	})

	t.Run("not required user activation failure falls back to email verification", func(t *testing.T) {
		userInfo := &entity.User{}
		calls := map[string]int{}

		err := applyRegistrationVerification(userInfo, false, registrationVerificationActions{
			sendActivationEmail: func() error {
				calls["sendActivationEmail"]++
				return nil
			},
			activateUser: func() error {
				calls["activateUser"]++
				return errors.New("activate failed")
			},
			markEmailAvailable: func() error {
				calls["markEmailAvailable"]++
				return nil
			},
		})

		require.NoError(t, err)
		assert.Equal(t, entity.EmailStatusToBeVerified, userInfo.MailStatus)
		assert.Equal(t, 1, calls["sendActivationEmail"])
		assert.Equal(t, 1, calls["activateUser"])
		assert.Zero(t, calls["markEmailAvailable"])
	})

	t.Run("not required email status failure falls back to email verification", func(t *testing.T) {
		userInfo := &entity.User{}
		calls := map[string]int{}

		err := applyRegistrationVerification(userInfo, false, registrationVerificationActions{
			sendActivationEmail: func() error {
				calls["sendActivationEmail"]++
				return nil
			},
			activateUser: func() error {
				calls["activateUser"]++
				return nil
			},
			markEmailAvailable: func() error {
				calls["markEmailAvailable"]++
				return errors.New("update failed")
			},
		})

		require.NoError(t, err)
		assert.Equal(t, entity.EmailStatusToBeVerified, userInfo.MailStatus)
		assert.Equal(t, 1, calls["sendActivationEmail"])
		assert.Equal(t, 1, calls["activateUser"])
		assert.Equal(t, 1, calls["markEmailAvailable"])
	})

	t.Run("fallback email failure returns before available email status", func(t *testing.T) {
		userInfo := &entity.User{}
		expectedErr := errors.New("email failed")

		err := applyRegistrationVerification(userInfo, false, registrationVerificationActions{
			sendActivationEmail: func() error {
				return expectedErr
			},
			activateUser: func() error {
				return errors.New("activate failed")
			},
			markEmailAvailable: func() error {
				return nil
			},
		})

		require.ErrorIs(t, err, expectedErr)
		assert.Equal(t, entity.EmailStatusToBeVerified, userInfo.MailStatus)
	})
}
