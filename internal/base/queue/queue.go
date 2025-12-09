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
	"sync"

	"github.com/segmentfault/pacman/log"
)

// Queue is a generic message queue service that processes messages asynchronously.
// It is thread-safe and supports graceful shutdown.
type Queue[T any] struct {
	name    string
	queue   chan T
	handler func(ctx context.Context, msg T) error
	mu      sync.RWMutex
	closed  bool
	wg      sync.WaitGroup
}

// New creates a new queue with the given name and buffer size.
func New[T any](name string, bufferSize int) *Queue[T] {
	q := &Queue[T]{
		name:  name,
		queue: make(chan T, bufferSize),
	}
	q.startWorker()
	return q
}

// Send enqueues a message to be processed asynchronously.
// It will block if the queue is full.
func (q *Queue[T]) Send(ctx context.Context, msg T) {
	q.mu.RLock()
	closed := q.closed
	q.mu.RUnlock()

	if closed {
		log.Warnf("[%s] queue is closed, dropping message", q.name)
		return
	}

	select {
	case q.queue <- msg:
		log.Debugf("[%s] enqueued message: %+v", q.name, msg)
	case <-ctx.Done():
		log.Warnf("[%s] context cancelled while sending message", q.name)
	}
}

// RegisterHandler sets the handler function for processing messages.
// This is thread-safe and can be called at any time.
func (q *Queue[T]) RegisterHandler(handler func(ctx context.Context, msg T) error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handler = handler
}

// Close gracefully shuts down the queue, waiting for pending messages to be processed.
func (q *Queue[T]) Close() {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return
	}
	q.closed = true
	q.mu.Unlock()

	close(q.queue)
	q.wg.Wait()
	log.Infof("[%s] queue closed", q.name)
}

// startWorker starts the background goroutine that processes messages.
func (q *Queue[T]) startWorker() {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for msg := range q.queue {
			q.processMessage(msg)
		}
	}()
}

// processMessage handles a single message with proper synchronization.
func (q *Queue[T]) processMessage(msg T) {
	q.mu.RLock()
	handler := q.handler
	q.mu.RUnlock()

	if handler == nil {
		log.Warnf("[%s] no handler registered, dropping message: %+v", q.name, msg)
		return
	}

	// Use background context for async processing
	// TODO: Consider adding timeout or using a derived context
	if err := handler(context.TODO(), msg); err != nil {
		log.Errorf("[%s] handler error: %v", q.name, err)
	}
}
