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

package content

import (
	"context"
	"errors"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/revision"
)

type checkCanUpdateRevisionErrorRepo struct {
	revision.RevisionRepo
	err error
}

func (r *checkCanUpdateRevisionErrorRepo) ExistUnreviewedByObjectID(context.Context, string) (*entity.Revision, bool, error) {
	return nil, false, r.err
}

func TestCheckCanUpdateRevisionPropagatesLookupError(t *testing.T) {
	wantErr := errors.New("revision lookup failed")
	service := &RevisionService{
		revisionRepo: &checkCanUpdateRevisionErrorRepo{err: wantErr},
	}

	_, err := service.CheckCanUpdateRevision(context.Background(), &schema.CheckCanQuestionUpdate{ID: "1"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("CheckCanUpdateRevision() error = %v, want %v", err, wantErr)
	}
}
