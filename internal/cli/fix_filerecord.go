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
	"strings"

	"github.com/apache/answer/internal/entity"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// scanFileRecords visits matching file records in id order, ignoring --limit.
func scanFileRecords(x *xorm.Engine, label string, cond builder.Cond, visit func(*entity.FileRecord) error) error {
	return runBatchedFix(newFixRunState(0), func(afterID int64) ([]*entity.FileRecord, int64, error) {
		records := make([]*entity.FileRecord, 0)
		err := x.Select("id, file_url").Where(builder.Gt{"id": afterID}.And(cond)).OrderBy("id").Limit(fixBatchSize).Find(&records)
		if err != nil {
			return nil, 0, fmt.Errorf("query %s file_record failed: %w", label, err)
		}
		var lastID int64
		if len(records) > 0 {
			lastID = int64(records[len(records)-1].ID)
		}
		return records, lastID, nil
	}, visit)
}

// fixFileRecordBySources ignores --limit and keeps unreferenced records (see
// findUnreferencedRecords) on their old URL.
func fixFileRecordBySources(x *xorm.Engine, opts *FixURLPrefixOptions, sources []string, label string) (fixResult, error) {
	held, err := findUnreferencedRecords(x, opts, sources, label)
	if err != nil {
		return fixResult{}, err
	}

	result := fixResult{}
	state := newFixRunState(0)
	err = scanFileRecords(x, label, builder.In("source", sources), func(record *entity.FileRecord) error {
		target := fmt.Sprintf("file_record id=%d", record.ID)
		if held[record.ID] {
			fmt.Printf("[%s] WARN %s: no revision references the new URL yet, keeping: %s\n", label, target, record.FileURL)
			result.Skipped++
			return nil
		}
		return applyValueFix(&result, opts, state, label, target, record.FileURL, func(newURL string) error {
			_, err := x.ID(record.ID).Cols("file_url").NoAutoTime().Update(&entity.FileRecord{FileURL: newURL})
			if err != nil {
				return fmt.Errorf("update %s file_record (id=%d) failed: %w", label, record.ID, err)
			}
			return nil
		})
	})
	if err != nil {
		return fixResult{}, err
	}
	return result, nil
}

// findUnreferencedRecords returns unlinked records whose new URL is in no
// revision yet; the cleaner would delete them. Dry runs assume revisions fixed.
func findUnreferencedRecords(x *xorm.Engine, opts *FixURLPrefixOptions, sources []string, label string) (map[int]bool, error) {
	pending := make(map[string][]int) // new URL -> record IDs
	maxLen := 0
	cond := builder.In("source", sources).And(builder.Eq{"object_id": 0, "status": entity.FileRecordStatusAvailable})
	err := scanFileRecords(x, label, cond, func(record *entity.FileRecord) error {
		if classifyValue(record.FileURL, opts.SrcPrefix, opts.DstPrefix) == matchFix {
			newURL := swapPrefix(record.FileURL, opts.SrcPrefix, opts.DstPrefix)
			pending[newURL] = append(pending[newURL], record.ID)
			maxLen = max(maxLen, len(newURL))
		}
		return nil
	})
	if err != nil || len(pending) == 0 {
		return nil, err
	}

	err = runBatchedFix(newFixRunState(0), func(afterID int64) ([]*entity.Revision, int64, error) {
		revisions := make([]*entity.Revision, 0)
		err := x.Select("id, content").Where("id > ?", afterID).OrderBy("id").Limit(fixBatchSize).Find(&revisions)
		if err != nil {
			return nil, 0, fmt.Errorf("query revisions for %s file_record failed: %w", label, err)
		}
		lastID, err := lastStringID(revisions, func(r *entity.Revision) string { return r.ID })
		return revisions, lastID, err
	}, func(revision *entity.Revision) error {
		content := revision.Content
		if opts.DryRun && classifyText(content, opts.SrcPrefix, opts.DstPrefix, jsonURLMarkers) == matchFix {
			content = replaceURLPrefix(content, opts.SrcPrefix, opts.DstPrefix, jsonURLMarkers)
		}
		// Pending URLs start with DST_PREFIX: try each such substring up to maxLen.
		for rest := content; ; rest = rest[1:] {
			i := strings.Index(rest, opts.DstPrefix)
			if i < 0 {
				break
			}
			rest = rest[i:]
			for n := len(opts.DstPrefix) + 1; n <= min(maxLen, len(rest)); n++ {
				delete(pending, rest[:n])
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	held := make(map[int]bool)
	for _, ids := range pending {
		for _, id := range ids {
			held[id] = true
		}
	}
	return held, nil
}

// fileRecordSync moves a source's records off SRC_PREFIX once their value is on
// DST_PREFIX, which also repairs records left behind by earlier runs.
type fileRecordSync struct {
	x     *xorm.Engine
	opts  *FixURLPrefixOptions
	label string
	ids   map[string][]int // SRC_PREFIX URL -> record IDs; file_url is not indexed
}

func newFileRecordSync(x *xorm.Engine, opts *FixURLPrefixOptions, label, source string) (*fileRecordSync, error) {
	s := &fileRecordSync{x: x, opts: opts, label: label, ids: make(map[string][]int)}
	if opts.DryRun {
		return s, nil
	}
	cond := builder.Eq{"source": source}.And(builder.Expr("file_url LIKE ?", opts.SrcPrefix+"%"))
	err := scanFileRecords(x, label, cond, func(record *entity.FileRecord) error {
		s.ids[record.FileURL] = append(s.ids[record.FileURL], record.ID)
		return nil
	})
	return s, err
}

func (s *fileRecordSync) sync(value string) error {
	if s.opts.DryRun || classifyValue(value, s.opts.SrcPrefix, s.opts.DstPrefix) != matchSkip {
		return nil
	}
	oldURL := swapPrefix(value, s.opts.DstPrefix, s.opts.SrcPrefix)
	ids, ok := s.ids[oldURL]
	if !ok {
		return nil
	}
	_, err := s.x.In("id", ids).Cols("file_url").NoAutoTime().Update(&entity.FileRecord{FileURL: value})
	if err != nil {
		return fmt.Errorf("update %s file_record (url=%s) failed: %w", s.label, oldURL, err)
	}
	delete(s.ids, oldURL)
	return nil
}
