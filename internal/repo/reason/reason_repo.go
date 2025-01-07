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

package reason

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/answer/internal/base/handler"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/config"
	"github.com/apache/answer/internal/service/reason_common"
	"github.com/segmentfault/pacman/log"
)

type reasonRepo struct {
	configService *config.ConfigService
}

func NewReasonRepo(configService *config.ConfigService) reason_common.ReasonRepo {
	return &reasonRepo{
		configService: configService,
	}
}

func (rr *reasonRepo) ListReasons(ctx context.Context, objectType, action string) (resp []*schema.ReasonItem, err error) {
	lang := handler.GetLangByCtx(ctx)
	reasonAction := fmt.Sprintf("%s.%s.reasons", objectType, action)
	resp = make([]*schema.ReasonItem, 0)

	reasonKeys, err := rr.configService.GetArrayStringValue(ctx, reasonAction)
	if err != nil {
		return nil, err
	}
	for _, reasonKey := range reasonKeys {
		cfg, err := rr.configService.GetConfigByKey(ctx, reasonKey)
		if err != nil {
			log.Error(err)
			continue
		}

		reason := &schema.ReasonItem{}
		err = json.Unmarshal(cfg.GetByteValue(), reason)
		if err != nil {
			log.Error(err)
			continue
		}
		reason.Translate(reasonKey, lang)
		reason.ReasonType = cfg.ID
		resp = append(resp, reason)
	}
	return resp, nil
}
