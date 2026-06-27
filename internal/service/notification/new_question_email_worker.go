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
	"sync"
	"time"

	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/pkg/token"
	"github.com/segmentfault/pacman/log"
)

const newQuestionEmailWorkerQueueSize = 128

type newQuestionEmailTask struct {
	UserIDs       []string
	QuestionTitle string
	QuestionID    string
	Tags          []string
	TagIDs        []string
}

type newQuestionEmailIntervalProvider func() time.Duration

type newQuestionEmailTimer interface {
	C() <-chan time.Time
	Stop()
}

type newQuestionEmailTimerFactory func(time.Duration) newQuestionEmailTimer

type newQuestionEmailWorker struct {
	tasks        chan newQuestionEmailTask
	send         newQuestionNotificationEmailSender
	interval     newQuestionEmailIntervalProvider
	timerFactory newQuestionEmailTimerFactory
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	closed       bool
	wg           sync.WaitGroup
}

func newQuestionEmailWorkerWithDefaults(
	interval newQuestionEmailIntervalProvider,
	send newQuestionNotificationEmailSender,
) *newQuestionEmailWorker {
	return newQuestionEmailWorkerWithBuffer(
		interval,
		send,
		newRealNewQuestionEmailTimer,
		newQuestionEmailWorkerQueueSize,
	)
}

func newQuestionEmailWorkerWithBuffer(
	interval newQuestionEmailIntervalProvider,
	send newQuestionNotificationEmailSender,
	timerFactory newQuestionEmailTimerFactory,
	bufferSize int,
) *newQuestionEmailWorker {
	if interval == nil {
		interval = newQuestionNotificationEmailSendInterval
	}
	if timerFactory == nil {
		timerFactory = newRealNewQuestionEmailTimer
	}
	ctx, cancel := context.WithCancel(context.Background())
	w := &newQuestionEmailWorker{
		tasks:        make(chan newQuestionEmailTask, bufferSize),
		send:         send,
		interval:     interval,
		timerFactory: timerFactory,
		ctx:          ctx,
		cancel:       cancel,
	}
	w.wg.Add(1)
	go w.run()
	return w
}

func (w *newQuestionEmailWorker) TryEnqueue(task newQuestionEmailTask) bool {
	if w == nil {
		log.Warnf("[new_question_email] worker is nil, dropping new question email task")
		return false
	}

	task = copyNewQuestionEmailTask(task)

	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.closed {
		log.Warnf("[new_question_email] worker is closed, dropping new question email task for question %s", task.QuestionID)
		return false
	}

	if w.ctx == nil {
		log.Warnf("[new_question_email] worker context is nil, dropping new question email task for question %s", task.QuestionID)
		return false
	}

	select {
	case <-w.ctx.Done():
		log.Warnf("[new_question_email] worker is canceled, dropping new question email task for question %s", task.QuestionID)
		return false
	default:
	}

	select {
	case w.tasks <- task:
		log.Debugf("[new_question_email] enqueued task for question %s to %d users", task.QuestionID, len(task.UserIDs))
		return true
	case <-w.ctx.Done():
		log.Warnf("[new_question_email] worker canceled while enqueueing task for question %s", task.QuestionID)
		return false
	default:
		log.Warnf("[new_question_email] queue is full, dropping new question email task for question %s", task.QuestionID)
		return false
	}
}

func (w *newQuestionEmailWorker) Close() {
	if w == nil {
		return
	}

	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	w.closed = true
	if w.cancel != nil {
		w.cancel()
	}
	w.mu.Unlock()

	w.wg.Wait()
	if dropped := w.dropPendingTasks(); dropped > 0 {
		log.Warnf("[new_question_email] dropped %d pending tasks during shutdown", dropped)
	}
	log.Infof("[new_question_email] worker closed")
}

func (w *newQuestionEmailWorker) run() {
	defer w.wg.Done()

	emailAttemptSent := false
	for {
		if w.ctx.Err() != nil {
			return
		}

		select {
		case <-w.ctx.Done():
			return
		case task := <-w.tasks:
			if w.ctx.Err() != nil {
				return
			}
			if !w.processTask(task, &emailAttemptSent) {
				return
			}
		}
	}
}

func (w *newQuestionEmailWorker) processTask(task newQuestionEmailTask, emailAttemptSent *bool) bool {
	for _, userID := range task.UserIDs {
		if w.ctx.Err() != nil {
			return false
		}
		if *emailAttemptSent {
			interval := w.interval()
			if interval > 0 && !waitNewQuestionEmailInterval(w.ctx, interval, w.timerFactory) {
				return false
			}
		}
		if w.ctx.Err() != nil {
			return false
		}
		if w.send == nil {
			log.Errorf("[new_question_email] sender is nil, dropping email attempt for user %s question %s", userID, task.QuestionID)
			*emailAttemptSent = true
			continue
		}
		w.send(w.ctx, userID, task.newRawData())
		*emailAttemptSent = true
	}
	return true
}

func (w *newQuestionEmailWorker) dropPendingTasks() int {
	dropped := 0
	for {
		select {
		case <-w.tasks:
			dropped++
		default:
			return dropped
		}
	}
}

func waitNewQuestionEmailInterval(
	ctx context.Context,
	interval time.Duration,
	timerFactory newQuestionEmailTimerFactory,
) bool {
	if interval <= 0 {
		return true
	}
	if timerFactory == nil {
		timerFactory = newRealNewQuestionEmailTimer
	}
	timer := timerFactory(interval)
	defer timer.Stop()

	select {
	case <-timer.C():
		return true
	case <-ctx.Done():
		return false
	}
}

func (task newQuestionEmailTask) newRawData() *schema.NewQuestionTemplateRawData {
	return &schema.NewQuestionTemplateRawData{
		QuestionTitle:   task.QuestionTitle,
		QuestionID:      task.QuestionID,
		UnsubscribeCode: token.GenerateToken(),
		Tags:            copyStringSlice(task.Tags),
		TagIDs:          copyStringSlice(task.TagIDs),
	}
}

func newQuestionEmailTaskFromRawData(
	userIDs []string,
	rawData *schema.NewQuestionTemplateRawData,
) newQuestionEmailTask {
	if rawData == nil {
		return newQuestionEmailTask{UserIDs: copyStringSlice(userIDs)}
	}
	return newQuestionEmailTask{
		UserIDs:       copyStringSlice(userIDs),
		QuestionTitle: rawData.QuestionTitle,
		QuestionID:    rawData.QuestionID,
		Tags:          copyStringSlice(rawData.Tags),
		TagIDs:        copyStringSlice(rawData.TagIDs),
	}
}

func copyNewQuestionEmailTask(task newQuestionEmailTask) newQuestionEmailTask {
	task.UserIDs = copyStringSlice(task.UserIDs)
	task.Tags = copyStringSlice(task.Tags)
	task.TagIDs = copyStringSlice(task.TagIDs)
	return task
}

func copyStringSlice(values []string) []string {
	if values == nil {
		return nil
	}
	copied := make([]string, len(values))
	copy(copied, values)
	return copied
}

type realNewQuestionEmailTimer struct {
	timer *time.Timer
}

func newRealNewQuestionEmailTimer(interval time.Duration) newQuestionEmailTimer {
	return &realNewQuestionEmailTimer{timer: time.NewTimer(interval)}
}

func (t *realNewQuestionEmailTimer) C() <-chan time.Time {
	return t.timer.C
}

func (t *realNewQuestionEmailTimer) Stop() {
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
		}
	}
}
