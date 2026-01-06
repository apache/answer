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

package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testMessage struct {
	ID   int
	Data string
}

func TestQueue_SendAndReceive(t *testing.T) {
	q := New[*testMessage]("test", 10)
	defer q.Close()

	received := make(chan *testMessage, 1)
	q.RegisterHandler(func(ctx context.Context, msg *testMessage) error {
		received <- msg
		return nil
	})

	msg := &testMessage{ID: 1, Data: "hello"}
	q.Send(context.Background(), msg)

	select {
	case r := <-received:
		if r.ID != msg.ID || r.Data != msg.Data {
			t.Errorf("received message mismatch: got %+v, want %+v", r, msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestQueue_MultipleMessages(t *testing.T) {
	q := New[*testMessage]("test", 10)
	defer q.Close()

	var count atomic.Int32
	var wg sync.WaitGroup
	numMessages := 100
	wg.Add(numMessages)

	q.RegisterHandler(func(ctx context.Context, msg *testMessage) error {
		count.Add(1)
		wg.Done()
		return nil
	})

	for i := range numMessages {
		q.Send(context.Background(), &testMessage{ID: i})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if int(count.Load()) != numMessages {
			t.Errorf("expected %d messages, got %d", numMessages, count.Load())
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout: only received %d of %d messages", count.Load(), numMessages)
	}
}

func TestQueue_NoHandlerDropsMessage(t *testing.T) {
	q := New[*testMessage]("test", 10)
	defer q.Close()

	// Send without handler - should not panic
	q.Send(context.Background(), &testMessage{ID: 1})

	// Give time for the message to be processed (dropped)
	time.Sleep(100 * time.Millisecond)
}

func TestQueue_RegisterHandlerAfterSend(t *testing.T) {
	q := New[*testMessage]("test", 10)
	defer q.Close()

	received := make(chan *testMessage, 1)

	// Send first
	q.Send(context.Background(), &testMessage{ID: 1})

	// Small delay then register handler
	time.Sleep(50 * time.Millisecond)
	q.RegisterHandler(func(ctx context.Context, msg *testMessage) error {
		received <- msg
		return nil
	})

	// Send another message that should be received
	q.Send(context.Background(), &testMessage{ID: 2})

	select {
	case r := <-received:
		if r.ID != 2 {
			// First message was dropped (no handler), second should be received
			t.Logf("received message ID: %d", r.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestQueue_Close(t *testing.T) {
	q := New[*testMessage]("test", 10)

	var count atomic.Int32
	q.RegisterHandler(func(ctx context.Context, msg *testMessage) error {
		count.Add(1)
		return nil
	})

	// Send some messages
	for i := range 5 {
		q.Send(context.Background(), &testMessage{ID: i})
	}

	// Close and wait
	q.Close()

	// All messages should have been processed
	if count.Load() != 5 {
		t.Errorf("expected 5 messages processed, got %d", count.Load())
	}

	// Sending after close should not panic
	q.Send(context.Background(), &testMessage{ID: 99})
}

func TestQueue_ConcurrentSend(t *testing.T) {
	q := New[*testMessage]("test", 100)
	defer q.Close()

	var count atomic.Int32
	q.RegisterHandler(func(ctx context.Context, msg *testMessage) error {
		count.Add(1)
		return nil
	})

	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 100

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range messagesPerGoroutine {
				q.Send(context.Background(), &testMessage{ID: id*1000 + j})
			}
		}(i)
	}

	wg.Wait()

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	expected := int32(numGoroutines * messagesPerGoroutine)
	if count.Load() != expected {
		t.Errorf("expected %d messages, got %d", expected, count.Load())
	}
}

func TestQueue_ConcurrentRegisterHandler(t *testing.T) {
	q := New[*testMessage]("test", 10)
	defer q.Close()

	// Concurrently register handlers - should not race
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.RegisterHandler(func(ctx context.Context, msg *testMessage) error {
				return nil
			})
		}()
	}
	wg.Wait()
}

// TestQueue_SendCloseRace is a regression test for the race condition between
// Send and Close. Without proper synchronization, concurrent Send and Close
// calls could cause a "send on closed channel" panic.
// Run with: go test -race -run TestQueue_SendCloseRace
func TestQueue_SendCloseRace(t *testing.T) {
	for i := range 100 {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Use large buffer to avoid blocking on channel send while holding RLock
			q := New[*testMessage]("test-race", 1000)
			q.RegisterHandler(func(ctx context.Context, msg *testMessage) error {
				return nil
			})

			var wg sync.WaitGroup

			// Use cancellable context so senders can exit when Close is called
			ctx, cancel := context.WithCancel(context.Background())

			// Start multiple senders
			for j := range 10 {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for k := range 100 {
						q.Send(ctx, &testMessage{ID: id*1000 + k})
					}
				}(j)
			}

			// Close while senders are still running
			go func() {
				time.Sleep(time.Microsecond * 10)
				cancel() // Cancel context to unblock any waiting senders
				q.Close()
			}()

			wg.Wait()
		})
	}
}
