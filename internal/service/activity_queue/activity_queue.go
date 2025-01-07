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

package activity_queue

import (
	"context"

	"github.com/apache/answer/internal/schema"
	"github.com/segmentfault/pacman/log"
)

type ActivityQueueService interface {
	Send(ctx context.Context, msg *schema.ActivityMsg)
	RegisterHandler(handler func(ctx context.Context, msg *schema.ActivityMsg) error)
}

type activityQueueService struct {
	Queue   chan *schema.ActivityMsg
	Handler func(ctx context.Context, msg *schema.ActivityMsg) error
}

func (ns *activityQueueService) Send(ctx context.Context, msg *schema.ActivityMsg) {
	ns.Queue <- msg
}

func (ns *activityQueueService) RegisterHandler(
	handler func(ctx context.Context, msg *schema.ActivityMsg) error) {
	ns.Handler = handler
}

func (ns *activityQueueService) working() {
	go func() {
		for msg := range ns.Queue {
			log.Debugf("received activity %+v", msg)
			if ns.Handler == nil {
				log.Warnf("no handler for activity")
				continue
			}
			if err := ns.Handler(context.Background(), msg); err != nil {
				log.Error(err)
			}
		}
	}()
}

// NewActivityQueueService create a new activity queue service
func NewActivityQueueService() ActivityQueueService {
	ns := &activityQueueService{}
	ns.Queue = make(chan *schema.ActivityMsg, 128)
	ns.working()
	return ns
}
