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
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apache/answer/internal/schema"
)

func TestNewQuestionEmailWorkerDelaysBetweenAttempts(t *testing.T) {
	timerFactory := newFakeNewQuestionEmailTimerFactory()
	sendCh := make(chan newQuestionEmailSendEvent, 2)
	worker := newQuestionEmailWorkerWithBuffer(
		func() time.Duration { return 3 * time.Second },
		newQuestionEmailSendRecorder(sendCh),
		timerFactory.New,
		2,
	)
	defer worker.Close()

	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-1", "user-1", "user-2")) {
		t.Fatalf("TryEnqueue() = false, want true")
	}

	first := receiveNewQuestionEmailSend(t, sendCh)
	if first.userID != "user-1" {
		t.Fatalf("first send user = %s, want user-1", first.userID)
	}
	timer := timerFactory.WaitForTimer(t)
	assertNoNewQuestionEmailSend(t, sendCh)
	timer.Fire()

	second := receiveNewQuestionEmailSend(t, sendCh)
	if second.userID != "user-2" {
		t.Fatalf("second send user = %s, want user-2", second.userID)
	}
	assertUniqueNewQuestionUnsubscribeCodes(t, []string{
		first.rawData.UnsubscribeCode,
		second.rawData.UnsubscribeCode,
	})
	if got := timerFactory.Durations(); !reflect.DeepEqual(got, []time.Duration{3 * time.Second}) {
		t.Fatalf("timer durations = %v, want [3s]", got)
	}
}

func TestNewQuestionEmailWorkerDelayContinuesAcrossTaskBoundaries(t *testing.T) {
	timerFactory := newFakeNewQuestionEmailTimerFactory()
	sendCh := make(chan newQuestionEmailSendEvent, 2)
	worker := newQuestionEmailWorkerWithBuffer(
		func() time.Duration { return 5 * time.Second },
		newQuestionEmailSendRecorder(sendCh),
		timerFactory.New,
		2,
	)
	defer worker.Close()

	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-1", "user-1")) {
		t.Fatalf("TryEnqueue() task 1 = false, want true")
	}
	first := receiveNewQuestionEmailSend(t, sendCh)
	if first.userID != "user-1" {
		t.Fatalf("first send user = %s, want user-1", first.userID)
	}

	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-2", "user-2")) {
		t.Fatalf("TryEnqueue() task 2 = false, want true")
	}
	timer := timerFactory.WaitForTimer(t)
	assertNoNewQuestionEmailSend(t, sendCh)
	timer.Fire()

	second := receiveNewQuestionEmailSend(t, sendCh)
	if second.userID != "user-2" {
		t.Fatalf("second send user = %s, want user-2", second.userID)
	}
}

func TestNewQuestionEmailWorkerZeroIntervalSendsWithoutTimers(t *testing.T) {
	var timerCount int
	sendCh := make(chan newQuestionEmailSendEvent, 3)
	worker := newQuestionEmailWorkerWithBuffer(
		func() time.Duration { return 0 },
		newQuestionEmailSendRecorder(sendCh),
		func(time.Duration) newQuestionEmailTimer {
			timerCount++
			return newFakeNewQuestionEmailTimer()
		},
		2,
	)
	defer worker.Close()

	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-1", "user-1", "user-2", "user-3")) {
		t.Fatalf("TryEnqueue() = false, want true")
	}

	gotUsers := []string{
		receiveNewQuestionEmailSend(t, sendCh).userID,
		receiveNewQuestionEmailSend(t, sendCh).userID,
		receiveNewQuestionEmailSend(t, sendCh).userID,
	}
	if !reflect.DeepEqual(gotUsers, []string{"user-1", "user-2", "user-3"}) {
		t.Fatalf("send users = %v, want [user-1 user-2 user-3]", gotUsers)
	}
	if timerCount != 0 {
		t.Fatalf("timer count = %d, want 0", timerCount)
	}
}

func TestNewQuestionEmailWorkerCloseCancelsPendingWaitAndDropsQueuedTasks(t *testing.T) {
	timerFactory := newFakeNewQuestionEmailTimerFactory()
	sendCh := make(chan newQuestionEmailSendEvent, 3)
	worker := newQuestionEmailWorkerWithBuffer(
		func() time.Duration { return time.Hour },
		newQuestionEmailSendRecorder(sendCh),
		timerFactory.New,
		2,
	)

	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-1", "user-1", "user-2")) {
		t.Fatalf("TryEnqueue() task 1 = false, want true")
	}
	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-2", "user-3")) {
		t.Fatalf("TryEnqueue() task 2 = false, want true")
	}

	first := receiveNewQuestionEmailSend(t, sendCh)
	if first.userID != "user-1" {
		t.Fatalf("first send user = %s, want user-1", first.userID)
	}
	timer := timerFactory.WaitForTimer(t)

	worker.Close()
	timer.AssertStopped(t)
	assertNoNewQuestionEmailSend(t, sendCh)
	if got := len(worker.tasks); got != 0 {
		t.Fatalf("pending tasks after Close() = %d, want 0", got)
	}
	if worker.TryEnqueue(newQuestionEmailWorkerTask("question-3", "user-4")) {
		t.Fatalf("TryEnqueue() after Close() = true, want false")
	}
}

func TestNewQuestionEmailWorkerProcessesSerially(t *testing.T) {
	entered := make(chan string, 2)
	releaseFirst := make(chan struct{})
	worker := newQuestionEmailWorkerWithBuffer(
		func() time.Duration { return 0 },
		func(_ context.Context, userID string, _ *schema.NewQuestionTemplateRawData) {
			entered <- userID
			if userID == "user-1" {
				<-releaseFirst
			}
		},
		nil,
		2,
	)
	defer worker.Close()

	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-1", "user-1", "user-2")) {
		t.Fatalf("TryEnqueue() = false, want true")
	}
	if got := receiveString(t, entered); got != "user-1" {
		t.Fatalf("first send user = %s, want user-1", got)
	}
	assertNoString(t, entered)

	close(releaseFirst)
	if got := receiveString(t, entered); got != "user-2" {
		t.Fatalf("second send user = %s, want user-2", got)
	}
}

func TestNewQuestionEmailWorkerBuildsFreshRawDataPerAttempt(t *testing.T) {
	sendCh := make(chan newQuestionEmailSendEvent, 2)
	worker := newQuestionEmailWorkerWithBuffer(
		func() time.Duration { return 0 },
		func(_ context.Context, userID string, rawData *schema.NewQuestionTemplateRawData) {
			if rawData.QuestionAuthorUserID != "" {
				t.Errorf("QuestionAuthorUserID = %q, want empty", rawData.QuestionAuthorUserID)
			}
			if userID == "user-1" {
				rawData.Tags[0] = "mutated"
				rawData.TagIDs[0] = "mutated"
			}
			sendCh <- newQuestionEmailSendEvent{userID: userID, rawData: rawData}
		},
		nil,
		2,
	)
	defer worker.Close()

	if !worker.TryEnqueue(newQuestionEmailTask{
		UserIDs:       []string{"user-1", "user-2"},
		QuestionTitle: "Question",
		QuestionID:    "question-1",
		Tags:          []string{"go"},
		TagIDs:        []string{"tag-1"},
	}) {
		t.Fatalf("TryEnqueue() = false, want true")
	}

	first := receiveNewQuestionEmailSend(t, sendCh)
	second := receiveNewQuestionEmailSend(t, sendCh)
	if first.rawData.UnsubscribeCode == "" || second.rawData.UnsubscribeCode == "" {
		t.Fatalf("unsubscribe codes must be non-empty: %q %q",
			first.rawData.UnsubscribeCode, second.rawData.UnsubscribeCode)
	}
	if first.rawData.UnsubscribeCode == second.rawData.UnsubscribeCode {
		t.Fatalf("unsubscribe codes must be unique, both were %q", first.rawData.UnsubscribeCode)
	}
	if !reflect.DeepEqual(second.rawData.Tags, []string{"go"}) ||
		!reflect.DeepEqual(second.rawData.TagIDs, []string{"tag-1"}) {
		t.Fatalf("second raw data tags = %v/%v, want original values",
			second.rawData.Tags, second.rawData.TagIDs)
	}
}

func TestNewQuestionEmailWorkerTryEnqueueCopiesTaskAndFailsFast(t *testing.T) {
	worker := newUnstartedNewQuestionEmailWorkerForTest(1)
	task := newQuestionEmailTask{
		UserIDs:       []string{"user-1"},
		QuestionTitle: "Question",
		QuestionID:    "question-1",
		Tags:          []string{"go"},
		TagIDs:        []string{"tag-1"},
	}
	if !worker.TryEnqueue(task) {
		t.Fatalf("TryEnqueue() = false, want true")
	}
	task.UserIDs[0] = "mutated-user"
	task.Tags[0] = "mutated-tag"
	task.TagIDs[0] = "mutated-tag-id"

	queuedTask := <-worker.tasks
	if !reflect.DeepEqual(queuedTask.UserIDs, []string{"user-1"}) ||
		!reflect.DeepEqual(queuedTask.Tags, []string{"go"}) ||
		!reflect.DeepEqual(queuedTask.TagIDs, []string{"tag-1"}) {
		t.Fatalf("queued task was mutated: %+v", queuedTask)
	}

	if !worker.TryEnqueue(newQuestionEmailWorkerTask("question-2", "user-2")) {
		t.Fatalf("TryEnqueue() refill = false, want true")
	}
	if worker.TryEnqueue(newQuestionEmailWorkerTask("question-3", "user-3")) {
		t.Fatalf("TryEnqueue() with full queue = true, want false")
	}

	worker.Close()
	if worker.TryEnqueue(newQuestionEmailWorkerTask("question-4", "user-4")) {
		t.Fatalf("TryEnqueue() after Close() = true, want false")
	}

	canceledWorker := newUnstartedNewQuestionEmailWorkerForTest(1)
	canceledWorker.cancel()
	if canceledWorker.TryEnqueue(newQuestionEmailWorkerTask("question-5", "user-5")) {
		t.Fatalf("TryEnqueue() after cancel = true, want false")
	}
}

func TestNewQuestionEmailWorkerTryEnqueueConcurrentClose(t *testing.T) {
	const (
		iterations = 100
		senders    = 32
	)

	for iteration := 0; iteration < iterations; iteration++ {
		worker := newUnstartedNewQuestionEmailWorkerForTest(1)
		if !worker.TryEnqueue(newQuestionEmailWorkerTask("already-queued", "queued-user")) {
			t.Fatalf("iteration %d: pre-fill TryEnqueue() = false, want true", iteration)
		}

		start := make(chan struct{})
		ready := make(chan struct{}, senders)
		panicCh := make(chan any, senders)
		var closeObserved atomic.Bool
		var acceptedAfterCloseObserved atomic.Int64
		var wg sync.WaitGroup

		for sender := 0; sender < senders; sender++ {
			wg.Add(1)
			go func(sender int) {
				defer wg.Done()
				defer func() {
					if recovered := recover(); recovered != nil {
						panicCh <- recovered
					}
				}()

				ready <- struct{}{}
				<-start
				for {
					accepted := worker.TryEnqueue(newQuestionEmailWorkerTask("question", "user"))
					if accepted && closeObserved.Load() {
						acceptedAfterCloseObserved.Add(1)
					}
					if closeObserved.Load() {
						return
					}
					runtime.Gosched()
				}
			}(sender)
		}
		for sender := 0; sender < senders; sender++ {
			<-ready
		}

		closeDoneObserved := make(chan struct{})
		go func() {
			<-worker.ctx.Done()
			closeObserved.Store(true)
			close(closeDoneObserved)
		}()

		close(start)
		runtime.Gosched()

		closeDone := make(chan struct{})
		go func() {
			worker.Close()
			close(closeDone)
		}()

		select {
		case <-closeDone:
		case <-time.After(time.Second):
			t.Fatalf("iteration %d: Close() did not return", iteration)
		}
		select {
		case <-closeDoneObserved:
		case <-time.After(time.Second):
			t.Fatalf("iteration %d: close was not observed", iteration)
		}

		wgDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(wgDone)
		}()
		select {
		case <-wgDone:
		case <-time.After(time.Second):
			t.Fatalf("iteration %d: TryEnqueue goroutines did not return", iteration)
		}

		select {
		case recovered := <-panicCh:
			t.Fatalf("iteration %d: TryEnqueue panicked during Close(): %v", iteration, recovered)
		default:
		}
		if got := acceptedAfterCloseObserved.Load(); got != 0 {
			t.Fatalf("iteration %d: accepted %d enqueue attempts after close was observed, want 0",
				iteration, got)
		}
		if worker.TryEnqueue(newQuestionEmailWorkerTask("after-close", "user")) {
			t.Fatalf("iteration %d: TryEnqueue() after Close() = true, want false", iteration)
		}
		if got := len(worker.tasks); got != 0 {
			t.Fatalf("iteration %d: pending tasks after Close() = %d, want 0", iteration, got)
		}
	}
}

func TestWaitNewQuestionEmailIntervalCancel(t *testing.T) {
	timerFactory := newFakeNewQuestionEmailTimerFactory()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool, 1)
	go func() {
		done <- waitNewQuestionEmailInterval(ctx, time.Minute, timerFactory.New)
	}()

	timer := timerFactory.WaitForTimer(t)
	cancel()

	select {
	case got := <-done:
		if got {
			t.Fatalf("waitNewQuestionEmailInterval() = true, want false")
		}
	case <-time.After(time.Second):
		t.Fatalf("waitNewQuestionEmailInterval() did not return after cancellation")
	}
	timer.AssertStopped(t)
}

type newQuestionEmailSendEvent struct {
	userID  string
	rawData *schema.NewQuestionTemplateRawData
}

func newQuestionEmailSendRecorder(sendCh chan<- newQuestionEmailSendEvent) newQuestionNotificationEmailSender {
	return func(_ context.Context, userID string, rawData *schema.NewQuestionTemplateRawData) {
		sendCh <- newQuestionEmailSendEvent{userID: userID, rawData: rawData}
	}
}

func newQuestionEmailWorkerTask(questionID string, userIDs ...string) newQuestionEmailTask {
	return newQuestionEmailTask{
		UserIDs:       userIDs,
		QuestionTitle: "Question",
		QuestionID:    questionID,
		Tags:          []string{"go"},
		TagIDs:        []string{"tag-1"},
	}
}

func newUnstartedNewQuestionEmailWorkerForTest(bufferSize int) *newQuestionEmailWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &newQuestionEmailWorker{
		tasks:        make(chan newQuestionEmailTask, bufferSize),
		interval:     func() time.Duration { return 0 },
		timerFactory: newRealNewQuestionEmailTimer,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func receiveNewQuestionEmailSend(t *testing.T, sendCh <-chan newQuestionEmailSendEvent) newQuestionEmailSendEvent {
	t.Helper()
	select {
	case event := <-sendCh:
		return event
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for new question email send")
		return newQuestionEmailSendEvent{}
	}
}

func assertNoNewQuestionEmailSend(t *testing.T, sendCh <-chan newQuestionEmailSendEvent) {
	t.Helper()
	select {
	case event := <-sendCh:
		t.Fatalf("unexpected new question email send: %+v", event)
	default:
	}
}

func receiveString(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for string")
		return ""
	}
}

func assertNoString(t *testing.T, ch <-chan string) {
	t.Helper()
	select {
	case value := <-ch:
		t.Fatalf("unexpected string: %s", value)
	default:
	}
}

type fakeNewQuestionEmailTimerFactory struct {
	timers    chan *fakeNewQuestionEmailTimer
	mu        sync.Mutex
	durations []time.Duration
}

func newFakeNewQuestionEmailTimerFactory() *fakeNewQuestionEmailTimerFactory {
	return &fakeNewQuestionEmailTimerFactory{
		timers: make(chan *fakeNewQuestionEmailTimer, 16),
	}
}

func (f *fakeNewQuestionEmailTimerFactory) New(duration time.Duration) newQuestionEmailTimer {
	timer := newFakeNewQuestionEmailTimer()

	f.mu.Lock()
	f.durations = append(f.durations, duration)
	f.mu.Unlock()

	f.timers <- timer
	return timer
}

func (f *fakeNewQuestionEmailTimerFactory) WaitForTimer(t *testing.T) *fakeNewQuestionEmailTimer {
	t.Helper()
	select {
	case timer := <-f.timers:
		return timer
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for timer")
		return nil
	}
}

func (f *fakeNewQuestionEmailTimerFactory) Durations() []time.Duration {
	f.mu.Lock()
	defer f.mu.Unlock()

	durations := make([]time.Duration, len(f.durations))
	copy(durations, f.durations)
	return durations
}

type fakeNewQuestionEmailTimer struct {
	ch      chan time.Time
	stopped chan struct{}
	once    sync.Once
}

func newFakeNewQuestionEmailTimer() *fakeNewQuestionEmailTimer {
	return &fakeNewQuestionEmailTimer{
		ch:      make(chan time.Time, 1),
		stopped: make(chan struct{}),
	}
}

func (t *fakeNewQuestionEmailTimer) C() <-chan time.Time {
	return t.ch
}

func (t *fakeNewQuestionEmailTimer) Stop() {
	t.once.Do(func() {
		close(t.stopped)
	})
}

func (t *fakeNewQuestionEmailTimer) Fire() {
	t.ch <- time.Now()
}

func (t *fakeNewQuestionEmailTimer) AssertStopped(tb testing.TB) {
	tb.Helper()
	select {
	case <-t.stopped:
	case <-time.After(time.Second):
		tb.Fatalf("timer was not stopped")
	}
}
