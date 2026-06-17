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

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/token"
)

const newQuestionNotificationEmailSendIntervalEnv = "NEW_QUESTION_NOTIFICATION_EMAIL_SEND_INTERVAL_SECONDS"

const maxNewQuestionNotificationEmailSendIntervalSeconds = int64(1<<63-1) / int64(time.Second)

type newQuestionNotificationEmailSleeper func(time.Duration)

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
	if err != nil || seconds < 0 || seconds > maxNewQuestionNotificationEmailSendIntervalSeconds {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func sendNewQuestionNotificationEmailsWithInterval(
	ctx context.Context,
	subscribers []*NewQuestionSubscriber,
	rawData *schema.NewQuestionTemplateRawData,
	interval time.Duration,
	sleep newQuestionNotificationEmailSleeper,
	send newQuestionNotificationEmailSender,
) {
	if rawData == nil || send == nil {
		return
	}
	if sleep == nil {
		sleep = time.Sleep
	}

	emailAttempts := 0
	for _, subscriber := range subscribers {
		for _, channel := range subscriber.Channels {
			if !channel.Enable || channel.Key != constant.EmailChannel {
				continue
			}
			if interval > 0 && emailAttempts > 0 {
				sleep(interval)
			}
			send(ctx, subscriber.UserID, &schema.NewQuestionTemplateRawData{
				QuestionTitle:   rawData.QuestionTitle,
				QuestionID:      rawData.QuestionID,
				UnsubscribeCode: token.GenerateToken(),
				Tags:            rawData.Tags,
				TagIDs:          rawData.TagIDs,
			})
			emailAttempts++
		}
	}
}
