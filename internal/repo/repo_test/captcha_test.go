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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/repo/captcha"
	"github.com/stretchr/testify/assert"
)

var (
	ip         = "127.0.0.1"
	actionType = "actionType"
	amount     = 1
)

func Test_captchaRepo_DelActionType(t *testing.T) {
	captchaRepo := captcha.NewCaptchaRepo(testDataSource)
	err := captchaRepo.SetActionType(context.TODO(), ip, actionType, "", amount)
	assert.NoError(t, err)

	actionInfo, err := captchaRepo.GetActionType(context.TODO(), ip, actionType)
	assert.NoError(t, err)
	assert.Equal(t, amount, actionInfo.Num)

	err = captchaRepo.DelActionType(context.TODO(), ip, actionType)
	assert.NoError(t, err)
}

func Test_captchaRepo_SetCaptcha(t *testing.T) {
	captchaRepo := captcha.NewCaptchaRepo(testDataSource)
	key, capt := "key", "1234"
	err := captchaRepo.SetCaptcha(context.TODO(), key, capt)
	assert.NoError(t, err)

	gotCaptcha, err := captchaRepo.GetCaptcha(context.TODO(), key)
	assert.NoError(t, err)
	assert.Equal(t, capt, gotCaptcha)
}
