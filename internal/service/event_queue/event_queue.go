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

package event_queue

import (
	"context"

	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/log"
)

type EventQueueService interface {
	Send(ctx context.Context, msg *schema.EventMsg)
	RegisterHandler(handler func(ctx context.Context, msg *schema.EventMsg) error)
}

type eventQueueService struct {
	Queue   chan *schema.EventMsg
	Handler func(ctx context.Context, msg *schema.EventMsg) error
}

func (ns *eventQueueService) Send(ctx context.Context, msg *schema.EventMsg) {
	ns.Queue <- msg
}

func (ns *eventQueueService) RegisterHandler(
	handler func(ctx context.Context, msg *schema.EventMsg) error) {
	ns.Handler = handler
}

func (ns *eventQueueService) working() {
	go func() {
		for msg := range ns.Queue {
			log.Debugf("received badge %+v", msg)
			if ns.Handler == nil {
				log.Warnf("no handler for badge")
				continue
			}
			if err := ns.Handler(context.Background(), msg); err != nil {
				log.Error(err)
			}
		}
	}()
}

// NewEventQueueService create a new badge queue service
func NewEventQueueService() EventQueueService {
	ns := &eventQueueService{}
	ns.Queue = make(chan *schema.EventMsg, 128)
	ns.working()
	return ns
}
