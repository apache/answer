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
	"time"

	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func addActivityTimeline(ctx context.Context, x *xorm.Engine) (err error) {
	switch x.Dialect().URI().DBType {
	case schemas.MYSQL:
		_, err = x.Context(ctx).Exec("ALTER TABLE `answer` CHANGE `updated_at` `updated_at` TIMESTAMP NULL DEFAULT NULL")
		if err != nil {
			return err
		}
		_, err = x.Context(ctx).Exec("ALTER TABLE `question` CHANGE `updated_at` `updated_at` TIMESTAMP NULL DEFAULT NULL")
		if err != nil {
			return err
		}
	case schemas.POSTGRES:
		_, err = x.Context(ctx).Exec(`ALTER TABLE "answer" ALTER COLUMN "updated_at" DROP NOT NULL, ALTER COLUMN "updated_at" SET DEFAULT NULL`)
		if err != nil {
			return err
		}
		_, err = x.Context(ctx).Exec(`ALTER TABLE "question" ALTER COLUMN "updated_at" DROP NOT NULL, ALTER COLUMN "updated_at" SET DEFAULT NULL`)
		if err != nil {
			return err
		}
	case schemas.SQLITE:
		_, err = x.Context(ctx).Exec(`DROP INDEX "IDX_answer_user_id";

ALTER TABLE "answer" RENAME TO "_answer_old_v3";

CREATE TABLE "answer" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME DEFAULT NULL,
  "question_id" INTEGER NOT NULL DEFAULT 0,
  "user_id" INTEGER NOT NULL DEFAULT 0,
  "original_text" TEXT NOT NULL,
  "parsed_text" TEXT NOT NULL,
  "status" INTEGER NOT NULL DEFAULT 1,
  "adopted" INTEGER NOT NULL DEFAULT 1,
  "comment_count" INTEGER NOT NULL DEFAULT 0,
  "vote_count" INTEGER NOT NULL DEFAULT 0,
  "revision_id" INTEGER NOT NULL DEFAULT 0
);

INSERT INTO "answer" ("id", "created_at", "updated_at", "question_id", "user_id", "original_text", "parsed_text", "status", "adopted", "comment_count", "vote_count", "revision_id") SELECT "id", "created_at", "updated_at", "question_id", "user_id", "original_text", "parsed_text", "status", "adopted", "comment_count", "vote_count", "revision_id" FROM "_answer_old_v3";

CREATE INDEX "IDX_answer_user_id"
ON "answer" (
  "user_id" ASC
);
DROP INDEX "IDX_question_user_id";

ALTER TABLE "question" RENAME TO "_question_old_v3";

CREATE TABLE "question" (
  "id" INTEGER NOT NULL,
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" DATETIME DEFAULT NULL,
  "user_id" INTEGER NOT NULL DEFAULT 0,
  "title" TEXT NOT NULL DEFAULT '',
  "original_text" TEXT NOT NULL,
  "parsed_text" TEXT NOT NULL,
  "status" INTEGER NOT NULL DEFAULT 1,
  "view_count" INTEGER NOT NULL DEFAULT 0,
  "unique_view_count" INTEGER NOT NULL DEFAULT 0,
  "vote_count" INTEGER NOT NULL DEFAULT 0,
  "answer_count" INTEGER NOT NULL DEFAULT 0,
  "collection_count" INTEGER NOT NULL DEFAULT 0,
  "follow_count" INTEGER NOT NULL DEFAULT 0,
  "accepted_answer_id" INTEGER NOT NULL DEFAULT 0,
  "last_answer_id" INTEGER NOT NULL DEFAULT 0,
  "post_update_time" DATETIME DEFAULT CURRENT_TIMESTAMP,
  "revision_id" INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);

INSERT INTO "question" ("id", "created_at", "updated_at", "user_id", "title", "original_text", "parsed_text", "status", "view_count", "unique_view_count", "vote_count", "answer_count", "collection_count", "follow_count", "accepted_answer_id", "last_answer_id", "post_update_time", "revision_id") SELECT "id", "created_at", "updated_at", "user_id", "title", "original_text", "parsed_text", "status", "view_count", "unique_view_count", "vote_count", "answer_count", "collection_count", "follow_count", "accepted_answer_id", "last_answer_id", "post_update_time", "revision_id" FROM "_question_old_v3";

CREATE INDEX "IDX_question_user_id"
ON "question" (
  "user_id" ASC
);`)
		if err != nil {
			return err
		}
	}

	// only increasing field length to 128
	type Config struct {
		ID  int    `xorm:"not null pk autoincr INT(11) id"`
		Key string `xorm:"unique VARCHAR(128) key"`
	}
	if err := x.Context(ctx).Sync(new(Config)); err != nil {
		return fmt.Errorf("sync config table failed: %w", err)
	}
	defaultConfigTable := []*entity.Config{
		{ID: 36, Key: "rank.question.add", Value: `1`},
		{ID: 37, Key: "rank.question.edit", Value: `200`},
		{ID: 38, Key: "rank.question.delete", Value: `-1`},
		{ID: 39, Key: "rank.question.vote_up", Value: `15`},
		{ID: 40, Key: "rank.question.vote_down", Value: `125`},
		{ID: 41, Key: "rank.answer.add", Value: `1`},
		{ID: 42, Key: "rank.answer.edit", Value: `200`},
		{ID: 43, Key: "rank.answer.delete", Value: `-1`},
		{ID: 44, Key: "rank.answer.accept", Value: `-1`},
		{ID: 45, Key: "rank.answer.vote_up", Value: `15`},
		{ID: 46, Key: "rank.answer.vote_down", Value: `125`},
		{ID: 47, Key: "rank.comment.add", Value: `1`},
		{ID: 48, Key: "rank.comment.edit", Value: `-1`},
		{ID: 49, Key: "rank.comment.delete", Value: `-1`},
		{ID: 50, Key: "rank.report.add", Value: `1`},
		{ID: 51, Key: "rank.tag.add", Value: `1500`},
		{ID: 52, Key: "rank.tag.edit", Value: `100`},
		{ID: 53, Key: "rank.tag.delete", Value: `-1`},
		{ID: 54, Key: "rank.tag.synonym", Value: `20000`},
		{ID: 55, Key: "rank.link.url_limit", Value: `10`},
		{ID: 56, Key: "rank.vote.detail", Value: `0`},

		{ID: 87, Key: "question.asked", Value: `0`},
		{ID: 88, Key: "question.closed", Value: `0`},
		{ID: 89, Key: "question.reopened", Value: `0`},
		{ID: 90, Key: "question.answered", Value: `0`},
		{ID: 91, Key: "question.commented", Value: `0`},
		{ID: 92, Key: "question.accept", Value: `0`},
		{ID: 93, Key: "question.edited", Value: `0`},
		{ID: 94, Key: "question.rollback", Value: `0`},
		{ID: 95, Key: "question.deleted", Value: `0`},
		{ID: 96, Key: "question.undeleted", Value: `0`},
		{ID: 97, Key: "answer.answered", Value: `0`},
		{ID: 98, Key: "answer.commented", Value: `0`},
		{ID: 99, Key: "answer.edited", Value: `0`},
		{ID: 100, Key: "answer.rollback", Value: `0`},
		{ID: 101, Key: "answer.undeleted", Value: `0`},
		{ID: 102, Key: "tag.created", Value: `0`},
		{ID: 103, Key: "tag.edited", Value: `0`},
		{ID: 104, Key: "tag.rollback", Value: `0`},
		{ID: 105, Key: "tag.deleted", Value: `0`},
		{ID: 106, Key: "tag.undeleted", Value: `0`},

		{ID: 107, Key: "rank.comment.vote_up", Value: `1`},
		{ID: 108, Key: "rank.comment.vote_down", Value: `1`},
		{ID: 109, Key: "rank.question.edit_without_review", Value: `2000`},
		{ID: 110, Key: "rank.answer.edit_without_review", Value: `2000`},
		{ID: 111, Key: "rank.tag.edit_without_review", Value: `20000`},
		{ID: 112, Key: "rank.answer.audit", Value: `2000`},
		{ID: 113, Key: "rank.question.audit", Value: `2000`},
		{ID: 114, Key: "rank.tag.audit", Value: `20000`},
	}
	for _, c := range defaultConfigTable {
		exist, err := x.Context(ctx).Get(&entity.Config{ID: c.ID, Key: c.Key})
		if err != nil {
			return fmt.Errorf("get config failed: %w", err)
		}
		if exist {
			if _, err = x.Context(ctx).Update(c, &entity.Config{ID: c.ID, Key: c.Key}); err != nil {
				log.Errorf("update %+v config failed: %s", c, err)
				return fmt.Errorf("update config failed: %w", err)
			}
			continue
		}
		if _, err = x.Context(ctx).Insert(&entity.Config{ID: c.ID, Key: c.Key, Value: c.Value}); err != nil {
			log.Errorf("insert %+v config failed: %s", c, err)
			return fmt.Errorf("add config failed: %w", err)
		}
	}

	type Revision struct {
		ID           string `xorm:"not null pk autoincr BIGINT(20) id"`
		ObjectID     string `xorm:"not null default 0 BIGINT(20) INDEX object_id"`
		ReviewUserID int64  `xorm:"not null default 0 BIGINT(20) review_user_id"`
	}
	type Activity struct {
		ID               string    `xorm:"not null pk autoincr BIGINT(20) id"`
		CancelledAt      time.Time `xorm:"TIMESTAMP cancelled_at"`
		UserID           string    `xorm:"not null index BIGINT(20) user_id"`
		TriggerUserID    int64     `xorm:"not null default 0 index BIGINT(20) trigger_user_id"`
		ObjectID         string    `xorm:"not null default 0 index BIGINT(20) object_id"`
		RevisionID       int64     `xorm:"not null default 0 BIGINT(20) revision_id"`
		OriginalObjectID string    `xorm:"not null default 0 BIGINT(20) original_object_id"`
	}
	type Tag struct {
		ID       string `xorm:"not null pk comment('tag_id') BIGINT(20) id"`
		SlugName string `xorm:"not null default '' unique VARCHAR(35) slug_name"`
		UserID   string `xorm:"not null default 0 BIGINT(20) user_id"`
	}
	type Question struct {
		ID             string    `xorm:"not null pk BIGINT(20) id"`
		UserID         string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
		UpdatedAt      time.Time `xorm:"updated_at TIMESTAMP"`
		LastEditUserID string    `xorm:"not null default 0 BIGINT(20) last_edit_user_id"`
		PostUpdateTime time.Time `xorm:"post_update_time TIMESTAMP"`
	}
	type Answer struct {
		ID             string    `xorm:"not null pk autoincr BIGINT(20) id"`
		UserID         string    `xorm:"not null default 0 BIGINT(20) INDEX user_id"`
		UpdatedAt      time.Time `xorm:"updated_at TIMESTAMP"`
		LastEditUserID string    `xorm:"not null default 0 BIGINT(20) last_edit_user_id"`
	}

	err = x.Context(ctx).Sync(new(Activity), new(Revision), new(Tag), new(Question), new(Answer))
	if err != nil {
		return fmt.Errorf("sync table failed %w", err)
	}
	return nil
}
