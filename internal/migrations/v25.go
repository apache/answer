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

package migrations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func addFileRecord(ctx context.Context, x *xorm.Engine) error {
	if err := x.Context(ctx).Sync(new(entity.FileRecord)); err != nil {
		return err
	}

	// Set default external_content_display to always_display
	legalInfo := &entity.SiteInfo{Type: "legal"}
	exist, err := x.Context(ctx).Get(legalInfo)
	if err != nil {
		return fmt.Errorf("get legal config failed: %w", err)
	}
	legalConfig := make(map[string]any)
	if exist {
		if err := json.Unmarshal([]byte(legalInfo.Content), &legalConfig); err != nil {
			return fmt.Errorf("unmarshal legal config failed: %w", err)
		}
	}
	legalConfig["external_content_display"] = "always_display"
	legalConfigBytes, _ := json.Marshal(legalConfig)
	if exist {
		legalInfo.Content = string(legalConfigBytes)
		_, err = x.Context(ctx).ID(legalInfo.ID).Cols("content").Update(legalInfo)
		if err != nil {
			return fmt.Errorf("update legal config failed: %w", err)
		}
	} else {
		legalInfo.Content = string(legalConfigBytes)
		legalInfo.Status = 1
		_, err = x.Context(ctx).Insert(legalInfo)
		if err != nil {
			return fmt.Errorf("insert legal config failed: %w", err)
		}
	}
	return nil
}
