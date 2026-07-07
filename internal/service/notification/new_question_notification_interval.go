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

package notification

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apache/answer/internal/schema"
)

const newQuestionNotificationEmailSendIntervalEnv = "NEW_QUESTION_NOTIFICATION_EMAIL_SEND_INTERVAL_SECONDS"

const maxNewQuestionNotificationEmailSendInterval = 5 * time.Minute

const maxNewQuestionNotificationEmailSendIntervalSeconds = int64(maxNewQuestionNotificationEmailSendInterval / time.Second)

type newQuestionNotificationEmailSender func(context.Context, string, *schema.NewQuestionTemplateRawData)

func newQuestionNotificationEmailSendInterval() time.Duration {
	return parseNewQuestionNotificationEmailSendInterval(os.Getenv(newQuestionNotificationEmailSendIntervalEnv))
}

func parseNewQuestionNotificationEmailSendInterval(value string) time.Duration {
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		return 0
	}
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil || seconds < 0 {
		return 0
	}
	if seconds > maxNewQuestionNotificationEmailSendIntervalSeconds {
		return maxNewQuestionNotificationEmailSendInterval
	}
	return time.Duration(seconds) * time.Second
}
