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

package badge

import (
	"context"
	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/event_queue"
	"github.com/segmentfault/pacman/log"
)

type BadgeEventService struct {
	data              *data.Data
	eventQueueService event_queue.EventQueueService
	badgeAwardRepo    BadgeAwardRepo
	badgeRepo         BadgeRepo
	eventRuleRepo     EventRuleRepo
	badgeAwardService *BadgeAwardService
}

type EventRuleHandler func(ctx context.Context, event *schema.EventMsg) (awards []*entity.BadgeAward, err error)

type EventRuleRepo interface {
	HandleEventWithRule(ctx context.Context, msg *schema.EventMsg) (awards []*entity.BadgeAward)
}

func NewBadgeEventService(
	data *data.Data,
	eventQueueService event_queue.EventQueueService,
	badgeRepo BadgeRepo,
	eventRuleRepo EventRuleRepo,
	badgeAwardService *BadgeAwardService,
) *BadgeEventService {
	n := &BadgeEventService{
		data:              data,
		eventQueueService: eventQueueService,
		badgeRepo:         badgeRepo,
		eventRuleRepo:     eventRuleRepo,
		badgeAwardService: badgeAwardService,
	}
	eventQueueService.RegisterHandler(n.Handler)
	return n
}

func (ns *BadgeEventService) Handler(ctx context.Context, msg *schema.EventMsg) error {
	awards := ns.eventRuleRepo.HandleEventWithRule(ctx, msg)
	if len(awards) == 0 {
		return nil
	}

	for _, award := range awards {
		err := ns.badgeAwardService.Award(ctx, award.BadgeID, award.UserID, award.AwardKey)
		if err != nil {
			log.Debugf("error awarding badge %s: %v", award.BadgeID, err)
		}
	}
	return nil
}
