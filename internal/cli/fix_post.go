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

package cli

import (
	"fmt"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/plugin"
	"xorm.io/xorm"
)

// postTable is a table with original_text/parsed_text; objectType 0 means its
// rows have no revisions.
type postTable struct {
	name       string
	objectType int
}

// postTables lists the tables that may link uploads. Comments have no upload
// button but keep pasted image URLs.
func postTables() []postTable {
	return []postTable{
		{"question", constant.ObjectTypeStrMapping[constant.QuestionObjectType]},
		{"answer", constant.ObjectTypeStrMapping[constant.AnswerObjectType]},
		{"tag", constant.ObjectTypeStrMapping[constant.TagObjectType]},
		{"comment", 0},
	}
}

// fixPost fixes each post with its revisions as one unit, so --limit never
// leaves old revisions that approving a pending edit would write back. It
// returns the units, counted by --limit, and the post file records, not counted.
func fixPost(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState) (units, files fixResult, err error) {
	for _, t := range postTables() {
		r, err := fixPostTable(x, opts, state, t)
		if err != nil {
			return fixResult{}, fixResult{}, err
		}
		units.add(r)
		if r, err = fixOrphanRevisions(x, opts, state, t); err != nil {
			return fixResult{}, fixResult{}, err
		}
		units.add(r)
	}

	// Must run after revisions, which it checks.
	files, err = fixFileRecordBySources(x, opts,
		[]string{string(plugin.UserPost), string(plugin.UserPostAttachment)}, fixTypePost)
	if err != nil {
		return fixResult{}, fixResult{}, err
	}
	return units, files, nil
}

type postRow struct {
	ID           string `xorm:"id"`
	OriginalText string `xorm:"original_text"`
	ParsedText   string `xorm:"parsed_text"`
}

type postUnit struct {
	row       *postRow
	revisions []*entity.Revision
}

func fixPostTable(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState, t postTable) (fixResult, error) {
	result := fixResult{}
	err := runBatchedFix(state,
		func(afterID int64) ([]*postUnit, int64, error) { return queryPostUnits(x, t, afterID) },
		func(u *postUnit) error {
			items := []textFix{{
				target:  fmt.Sprintf("%s id=%s", t.name, u.row.ID),
				texts:   []string{u.row.OriginalText, u.row.ParsedText},
				markers: textURLMarkers,
				update: func(db xorm.Interface) error {
					_, err := db.Table(t.name).Where("id = ?", u.row.ID).Cols("original_text", "parsed_text").
						Update(&postRow{
							OriginalText: fixTextField(u.row.OriginalText, opts, textURLMarkers),
							ParsedText:   fixTextField(u.row.ParsedText, opts, textURLMarkers),
						})
					if err != nil {
						return fmt.Errorf("update %s (id=%s) failed: %w", t.name, u.row.ID, err)
					}
					return nil
				},
			}}
			items = append(items, revisionFixes(opts, u.revisions)...)
			return applyUnitFix(x, &result, opts, state, items)
		})
	if err != nil {
		return fixResult{}, err
	}
	return result, nil
}

// queryPostUnits loads a page of posts and, in one more query, their revisions.
func queryPostUnits(x *xorm.Engine, t postTable, afterID int64) ([]*postUnit, int64, error) {
	rows := make([]*postRow, 0)
	err := x.Table(t.name).Select("id, original_text, parsed_text").
		Where("id > ?", afterID).
		OrderBy("id").
		Limit(fixBatchSize).
		Find(&rows)
	if err != nil {
		return nil, 0, fmt.Errorf("query %s failed: %w", t.name, err)
	}
	lastID, err := lastStringID(rows, func(r *postRow) string { return r.ID })
	if err != nil || len(rows) == 0 {
		return nil, lastID, err
	}

	units := make([]*postUnit, len(rows))
	byID := make(map[string]*postUnit, len(rows))
	ids := make([]string, len(rows))
	for i, row := range rows {
		units[i] = &postUnit{row: row}
		byID[row.ID] = units[i]
		ids[i] = row.ID
	}
	if t.objectType == 0 {
		return units, lastID, nil
	}

	revisions := make([]*entity.Revision, 0)
	err = x.Select("id, object_id, content").
		Where("object_type = ?", t.objectType).
		In("object_id", ids).
		OrderBy("id").
		Find(&revisions)
	if err != nil {
		return nil, 0, fmt.Errorf("query %s revision failed: %w", t.name, err)
	}
	for _, rev := range revisions {
		if u := byID[rev.ObjectID]; u != nil {
			u.revisions = append(u.revisions, rev)
		}
	}
	return units, lastID, nil
}

// fixOrphanRevisions fixes revisions whose post row is gone, one unit each.
func fixOrphanRevisions(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState, t postTable) (fixResult, error) {
	result := fixResult{}
	if t.objectType == 0 {
		return result, nil
	}
	err := runBatchedFix(state,
		func(afterID int64) ([]*entity.Revision, int64, error) {
			revisions := make([]*entity.Revision, 0)
			err := x.Select("id, content").
				Where("id > ?", afterID).
				And("object_type = ?", t.objectType).
				And(fmt.Sprintf("object_id NOT IN (SELECT id FROM %s)", t.name)).
				OrderBy("id").
				Limit(fixBatchSize).
				Find(&revisions)
			if err != nil {
				return nil, 0, fmt.Errorf("query orphan %s revision failed: %w", t.name, err)
			}
			lastID, err := lastStringID(revisions, func(r *entity.Revision) string { return r.ID })
			return revisions, lastID, err
		},
		func(rev *entity.Revision) error {
			return applyUnitFix(x, &result, opts, state, revisionFixes(opts, []*entity.Revision{rev}))
		})
	if err != nil {
		return fixResult{}, err
	}
	return result, nil
}

// fixTextField leaves fields not classified as fixable alone, so a field on a
// DST_PREFIX that starts with SRC_PREFIX is not nested (e.g. /uploads/cdn/cdn/).
func fixTextField(text string, opts *FixURLPrefixOptions, markers []string) string {
	if classifyText(text, opts.SrcPrefix, opts.DstPrefix, markers) != matchFix {
		return text
	}
	return replaceURLPrefix(text, opts.SrcPrefix, opts.DstPrefix, markers)
}

func revisionFixes(opts *FixURLPrefixOptions, revisions []*entity.Revision) []textFix {
	items := make([]textFix, 0, len(revisions))
	for _, rev := range revisions {
		items = append(items, textFix{
			target:  fmt.Sprintf("revision id=%s", rev.ID),
			texts:   []string{rev.Content},
			markers: jsonURLMarkers,
			update: func(db xorm.Interface) error {
				content := replaceURLPrefix(rev.Content, opts.SrcPrefix, opts.DstPrefix, jsonURLMarkers)
				if _, err := db.ID(rev.ID).Cols("content").NoAutoTime().Update(&entity.Revision{Content: content}); err != nil {
					return fmt.Errorf("update revision (id=%s) failed: %w", rev.ID, err)
				}
				return nil
			},
		})
	}
	return items
}

// textFix is one record of a unit: its text fields and how to write them back.
type textFix struct {
	target  string
	texts   []string
	markers []string
	update  func(db xorm.Interface) error
}

// applyUnitFix writes a unit's fixable records in one transaction and counts
// the unit once: affected if any record is fixable, skipped if a record is
// ambiguous (left for manual fix; the rest is still fixed) or none is fixable
// but some are already on DST_PREFIX.
func applyUnitFix(x *xorm.Engine, result *fixResult, opts *FixURLPrefixOptions, state *fixRunState, items []textFix) error {
	var fixable []textFix
	ambiguous, onDst := false, false
	for _, item := range items {
		switch classifyFields(item.texts, opts.SrcPrefix, opts.DstPrefix, item.markers) {
		case matchBoth:
			fmt.Printf("[%s] WARN %s: a field has BOTH prefixes and DST_PREFIX starts with SRC_PREFIX, skipping (manual fix needed)\n", fixTypePost, item.target)
			ambiguous = true
		case matchSkip:
			fmt.Printf("[%s] skip %s: already has DST_PREFIX\n", fixTypePost, item.target)
			onDst = true
		case matchFix:
			fixable = append(fixable, item)
		}
	}
	if ambiguous || (len(fixable) == 0 && onDst) {
		result.Skipped++
	}
	if len(fixable) == 0 {
		return nil
	}

	state.markAffected(result)
	for _, item := range fixable {
		fmt.Printf("[%s] %s: replacing %q with %q\n", fixTypePost, item.target, opts.SrcPrefix, opts.DstPrefix)
	}
	if opts.DryRun {
		return nil
	}
	_, err := x.Transaction(func(session *xorm.Session) (interface{}, error) {
		for _, item := range fixable {
			if err := item.update(session); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	result.Fixed++
	return nil
}
