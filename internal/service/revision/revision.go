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

package revision

import (
	"context"

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

// RevisionRepo revision repository
type RevisionRepo interface {
	AddRevision(ctx context.Context, revision *entity.Revision, autoUpdateRevisionID bool) (err error)
	GetRevisionByID(ctx context.Context, revisionID string) (revision *entity.Revision, exist bool, err error)
	GetLastRevisionByObjectID(ctx context.Context, objectID string) (revision *entity.Revision, exist bool, err error)
	GetLastRevisionByFileURL(ctx context.Context, fileURL string) (revision *entity.Revision, exist bool, err error)
	GetRevisionList(ctx context.Context, revision *entity.Revision) (revisionList []entity.Revision, err error)
	UpdateObjectRevisionId(ctx context.Context, revision *entity.Revision, session *xorm.Session) (err error)
	ExistUnreviewedByObjectID(ctx context.Context, objectID string) (revision *entity.Revision, exist bool, err error)
	GetUnreviewedRevisionPage(ctx context.Context, page, pageSize int, objectTypes []int) ([]*entity.Revision, int64, error)
	CountUnreviewedRevision(ctx context.Context, objectTypeList []int) (count int64, err error)
	UpdateStatus(ctx context.Context, id string, status int, reviewUserID string) (err error)
}
