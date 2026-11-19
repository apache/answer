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

package repo_test

import (
	"context"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/repo/search_common"
	"github.com/apache/answer/internal/repo/site_info"
	"github.com/apache/answer/internal/repo/tag"
	"github.com/apache/answer/internal/repo/tag_common"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/apache/answer/internal/repo/user"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/search_parser"
	"github.com/apache/answer/internal/service/siteinfo_common"
	tagcommon "github.com/apache/answer/internal/service/tag_common"
	usercommon "github.com/apache/answer/internal/service/user_common"
	"github.com/stretchr/testify/require"
)

// A [tag] filter must keep working for slugs that contain characters other
// than letters, digits and underscore, e.g. "react-native" (spaces in new tag
// names are turned into "-") or "node.js".
func TestSearchTagFilterWithNonWordSlug(t *testing.T) {
	ctx := context.Background()
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	siteInfoService := siteinfo_common.NewSiteInfoCommonService(site_info.NewSiteInfo(testDataSource))
	tagCommonService := tagcommon.NewTagCommonService(
		tag_common.NewTagCommonRepo(testDataSource, uniqueIDRepo),
		tag.NewTagRelRepo(testDataSource, uniqueIDRepo),
		tag.NewTagRepo(testDataSource, uniqueIDRepo),
		nil, siteInfoService, nil,
	)
	userCommon := usercommon.NewUserCommon(user.NewUserRepo(testDataSource), nil, nil, siteInfoService)
	parser := search_parser.NewSearchParser(tagCommonService, userCommon)
	searchRepo := search_common.NewSearchRepo(testDataSource, uniqueIDRepo, userCommon, tagCommonService)
	questionRepo := question.NewQuestionRepo(testDataSource, uniqueIDRepo)

	for _, slug := range []string{"search-hyphen-tag", "search.dot.tag"} {
		t.Run(slug, func(t *testing.T) {
			tagged := &entity.Tag{SlugName: slug, DisplayName: slug, Status: entity.TagStatusAvailable}
			require.NoError(t, tag_common.NewTagCommonRepo(testDataSource, uniqueIDRepo).AddTagList(ctx, []*entity.Tag{tagged}))

			newQuestion := func(title string) *entity.Question {
				q := &entity.Question{
					UserID: "1", Title: title, OriginalText: title, ParsedText: title,
					Status: entity.QuestionStatusAvailable, Show: entity.QuestionShow,
				}
				require.NoError(t, questionRepo.AddQuestion(ctx, q))
				return q
			}
			inTag := newQuestion("tagfilter needle in " + slug)
			outOfTag := newQuestion("tagfilter needle elsewhere " + slug)
			require.NoError(t, tag.NewTagRelRepo(testDataSource, uniqueIDRepo).AddTagRelList(ctx, []*entity.TagRel{
				{ObjectID: inTag.ID, TagID: tagged.ID, Status: entity.TagRelStatusAvailable},
			}))
			t.Cleanup(func() {
				_, _ = testDataSource.DB.Context(ctx).Where("tag_id = ?", tagged.ID).Delete(&entity.TagRel{})
				_, _ = testDataSource.DB.Context(ctx).ID(tagged.ID).Delete(&entity.Tag{})
				_, _ = testDataSource.DB.Context(ctx).ID(inTag.ID).Delete(&entity.Question{})
				_, _ = testDataSource.DB.Context(ctx).ID(outOfTag.ID).Delete(&entity.Question{})
			})

			// Same path as GET /answer/api/v1/search: bind + Check, then parse.
			dto := &schema.SearchDTO{Query: "[" + slug + "] tagfilter", Page: 1, Size: 20, Order: "newest"}
			_, err := dto.Check()
			require.NoError(t, err)
			cond := parser.ParseStructure(ctx, dto)
			require.Len(t, cond.Tags, 1, "tag filter dropped, query became %q", dto.Query)

			res, _, err := searchRepo.SearchContents(ctx, cond.Words, cond.Tags, cond.UserID, cond.VoteAmount, dto.Page, dto.Size, dto.Order)
			require.NoError(t, err)
			gotIDs := make([]string, 0, len(res))
			for _, r := range res {
				gotIDs = append(gotIDs, r.Object.ID)
			}
			require.Equal(t, []string{inTag.ID}, gotIDs)
		})
	}
}
