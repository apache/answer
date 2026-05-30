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
	"fmt"

	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

func addQuestionPrivateLevel(ctx context.Context, x *xorm.Engine) error {
	type QuestionPrivateLevel struct {
		PrivateLevel string `xorm:"not null default 'public' VARCHAR(20) private_level"`
	}

	if err := x.Context(ctx).Sync(new(QuestionPrivateLevel)); err != nil {
		log.Errorf("sync QuestionPrivateLevel failed: %s", err)
	}

	switch x.Dialect().URI().DBType {
	case "mysql", "MYSQL":
		_, err := x.Context(ctx).Exec(
			"ALTER TABLE `question` ADD COLUMN IF NOT EXISTS `private_level` VARCHAR(20) NOT NULL DEFAULT 'public'",
		)
		if err != nil {
			return fmt.Errorf("add private_level column to question (mysql): %w", err)
		}
	case "postgres", "POSTGRES":
		_, err := x.Context(ctx).Exec(
			`ALTER TABLE "question" ADD COLUMN IF NOT EXISTS "private_level" VARCHAR(20) NOT NULL DEFAULT 'public'`,
		)
		if err != nil {
			return fmt.Errorf("add private_level column to question (postgres): %w", err)
		}
	default:
		// sqlite3 - check if column exists first
		rows, err := x.Context(ctx).QueryString("PRAGMA table_info(question)")
		if err != nil {
			return fmt.Errorf("check question columns (sqlite): %w", err)
		}
		exists := false
		for _, row := range rows {
			if row["name"] == "private_level" {
				exists = true
				break
			}
		}
		if !exists {
			_, err = x.Context(ctx).Exec(
				"ALTER TABLE `question` ADD COLUMN `private_level` VARCHAR(20) NOT NULL DEFAULT 'public'",
			)
			if err != nil {
				return fmt.Errorf("add private_level column to question (sqlite): %w", err)
			}
		}
	}
	return nil
}
