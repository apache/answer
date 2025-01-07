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

package action

import (
	"context"
	"time"

	"github.com/apache/answer/plugin"
	"github.com/segmentfault/pacman/log"

	"github.com/apache/answer/internal/entity"
)

// ValidationStrategy
// true pass
// false need captcha
func (cs *CaptchaService) ValidationStrategy(ctx context.Context, unit, actionType string) bool {
	// If the captcha is not enabled, the verification is passed directly
	if !plugin.CaptchaEnabled() {
		return true
	}
	info, err := cs.captchaRepo.GetActionType(ctx, unit, actionType)
	if err != nil {
		log.Error(err)
		return false
	}
	switch actionType {
	case entity.CaptchaActionEmail:
		return cs.CaptchaActionEmail(ctx, unit, info)
	case entity.CaptchaActionPassword:
		return cs.CaptchaActionPassword(ctx, unit, info)
	case entity.CaptchaActionEditUserinfo:
		return cs.CaptchaActionEditUserinfo(ctx, unit, info)
	case entity.CaptchaActionQuestion:
		return cs.CaptchaActionQuestion(ctx, unit, info)
	case entity.CaptchaActionAnswer:
		return cs.CaptchaActionAnswer(ctx, unit, info)
	case entity.CaptchaActionComment:
		return cs.CaptchaActionComment(ctx, unit, info)
	case entity.CaptchaActionEdit:
		return cs.CaptchaActionEdit(ctx, unit, info)
	case entity.CaptchaActionInvitationAnswer:
		return cs.CaptchaActionInvitationAnswer(ctx, unit, info)
	case entity.CaptchaActionSearch:
		return cs.CaptchaActionSearch(ctx, unit, info)
	case entity.CaptchaActionReport:
		return cs.CaptchaActionReport(ctx, unit, info)
	case entity.CaptchaActionDelete:
		return cs.CaptchaActionDelete(ctx, unit, info)
	case entity.CaptchaActionVote:
		return cs.CaptchaActionVote(ctx, unit, info)

	}
	//actionType not found
	return false
}

func (cs *CaptchaService) CaptchaActionEmail(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	// You need a verification code every time
	return false
}

func (cs *CaptchaService) CaptchaActionPassword(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 3
	setTime := int64(60 * 30) //seconds
	now := time.Now().Unix()
	if now-actionInfo.LastTime <= setTime && actionInfo.Num >= setNum {
		return false
	}
	if now-actionInfo.LastTime != 0 && now-actionInfo.LastTime > setTime {
		cs.captchaRepo.SetActionType(ctx, unit, entity.CaptchaActionPassword, "", 0)
	}
	return true
}

func (cs *CaptchaService) CaptchaActionEditUserinfo(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 3
	setTime := int64(60 * 30) //seconds
	now := time.Now().Unix()
	if now-actionInfo.LastTime <= setTime && actionInfo.Num >= setNum {
		return false
	}
	if now-actionInfo.LastTime != 0 && now-actionInfo.LastTime > setTime {
		cs.captchaRepo.SetActionType(ctx, unit, entity.CaptchaActionEditUserinfo, "", 0)
	}
	return true
}

func (cs *CaptchaService) CaptchaActionQuestion(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 10
	setTime := int64(5) //seconds
	now := time.Now().Unix()
	if now-actionInfo.LastTime <= setTime || actionInfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionAnswer(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 10
	setTime := int64(5) //seconds
	now := time.Now().Unix()
	if now-actionInfo.LastTime <= setTime || actionInfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionComment(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 30
	setTime := int64(1) //seconds
	now := time.Now().Unix()
	if now-actionInfo.LastTime <= setTime || actionInfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionEdit(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 10
	if actionInfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionInvitationAnswer(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 30
	if actionInfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionSearch(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	now := time.Now().Unix()
	setNum := 20
	setTime := int64(60) //seconds
	if now-int64(actionInfo.LastTime) <= setTime && actionInfo.Num >= setNum {
		return false
	}
	if now-actionInfo.LastTime > setTime {
		cs.captchaRepo.SetActionType(ctx, unit, entity.CaptchaActionSearch, "", 0)
	}
	return true
}

func (cs *CaptchaService) CaptchaActionReport(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 30
	setTime := int64(1) //seconds
	now := time.Now().Unix()
	if now-actionInfo.LastTime <= setTime || actionInfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionDelete(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 5
	setTime := int64(5) //seconds
	now := time.Now().Unix()
	if now-actionInfo.LastTime <= setTime || actionInfo.Num >= setNum {
		return false
	}
	return true
}

func (cs *CaptchaService) CaptchaActionVote(ctx context.Context, unit string, actionInfo *entity.ActionRecordInfo) bool {
	if actionInfo == nil {
		return true
	}
	setNum := 40
	if actionInfo.Num >= setNum {
		return false
	}
	return true
}
