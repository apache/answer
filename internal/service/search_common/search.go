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

package search_common

import (
	"context"

	"github.com/apache/incubator-answer/internal/schema"
	"github.com/apache/incubator-answer/plugin"
)

type SearchRepo interface {
	SearchContents(ctx context.Context, words []string, tagIDs [][]string, userID string, votes, page, size int, order string) (resp []*schema.SearchResult, total int64, err error)
	SearchQuestions(ctx context.Context, words []string, tagIDs [][]string, notAccepted bool, views, answers int, page, size int, order string) (resp []*schema.SearchResult, total int64, err error)
	SearchAnswers(ctx context.Context, words []string, tagIDs [][]string, accepted bool, questionID string, page, size int, order string) (resp []*schema.SearchResult, total int64, err error)
	ParseSearchPluginResult(ctx context.Context, sres []plugin.SearchResult, words []string) (resp []*schema.SearchResult, err error)
}
