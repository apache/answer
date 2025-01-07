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
	"time"

	"github.com/apache/answer/internal/repo/export"
	"github.com/stretchr/testify/assert"
)

func Test_emailRepo_VerifyCode(t *testing.T) {
	emailRepo := export.NewEmailRepo(testDataSource)
	code, content := "1111", "{\"source_type\":\"\",\"e_mail\":\"\",\"user_id\":\"1\",\"skip_validation_latest_code\":false}"
	err := emailRepo.SetCode(context.TODO(), "1", code, content, time.Minute)
	assert.NoError(t, err)

	verifyContent, err := emailRepo.VerifyCode(context.TODO(), code)
	assert.NoError(t, err)
	assert.Equal(t, content, verifyContent)
}
