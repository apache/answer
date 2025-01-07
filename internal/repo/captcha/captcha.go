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

package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/action"
	"github.com/segmentfault/pacman/errors"
	"github.com/segmentfault/pacman/log"
)

// captchaRepo captcha repository
type captchaRepo struct {
	data *data.Data
}

// NewCaptchaRepo new repository
func NewCaptchaRepo(data *data.Data) action.CaptchaRepo {
	return &captchaRepo{
		data: data,
	}
}

func (cr *captchaRepo) SetActionType(ctx context.Context, unit, actionType, config string, amount int) (err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	value := &entity.ActionRecordInfo{}
	value.LastTime = now.Unix()
	value.Num = amount
	valueStr, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	err = cr.data.Cache.SetString(ctx, cacheKey, string(valueStr), 6*time.Minute)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

func (cr *captchaRepo) GetActionType(ctx context.Context, unit, actionType string) (actionInfo *entity.ActionRecordInfo, err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	res, exist, err := cr.data.Cache.GetString(ctx, cacheKey)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	actionInfo = &entity.ActionRecordInfo{}
	_ = json.Unmarshal([]byte(res), actionInfo)
	return actionInfo, nil
}

func (cr *captchaRepo) DelActionType(ctx context.Context, unit, actionType string) (err error) {
	now := time.Now()
	cacheKey := fmt.Sprintf("ActionRecord:%s@%s@%s", unit, actionType, now.Format("2006-1-02"))
	err = cr.data.Cache.Del(ctx, cacheKey)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// SetCaptcha set captcha to cache
func (cr *captchaRepo) SetCaptcha(ctx context.Context, key, captcha string) (err error) {
	err = cr.data.Cache.SetString(ctx, key, captcha, 6*time.Minute)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetCaptcha get captcha from cache
func (cr *captchaRepo) GetCaptcha(ctx context.Context, key string) (captcha string, err error) {
	captcha, exist, err := cr.data.Cache.GetString(ctx, key)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", fmt.Errorf("captcha not exist")
	}
	return captcha, nil
}

func (cr *captchaRepo) DelCaptcha(ctx context.Context, key string) (err error) {
	err = cr.data.Cache.Del(ctx, key)
	if err != nil {
		log.Debug(err)
	}
	return nil
}
