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

	"github.com/apache/answer/internal/base/data"
)

const fixBatchSize = 100

const (
	fixTypeAll      = "all"
	fixTypeBranding = "branding"
	fixTypeAvatar   = "avatar"
	fixTypePost     = "post"
)

// FixURLPrefixOptions holds options for the fix command.
type FixURLPrefixOptions struct {
	// FixType is one of: all, branding, avatar, post
	FixType   string
	SrcPrefix string
	DstPrefix string
	DryRun    bool
	// Limit is the max total number of rows fixed across all selected types.
	// 0 means unlimited.
	Limit int64
	// CacheFilePath is the path to the file-backed cache.
	// If non-empty, affected cache entries are cleared after DB writes.
	CacheFilePath string
}

type fixResult struct {
	Affected int64
	Fixed    int64
	Skipped  int64
}

func (r *fixResult) add(other fixResult) {
	r.Affected += other.Affected
	r.Fixed += other.Fixed
	r.Skipped += other.Skipped
}

func FixURLPrefix(dbConf *data.Database, opts *FixURLPrefixOptions) error {
	if opts == nil {
		return fmt.Errorf("fix options is nil")
	}
	if len(opts.SrcPrefix) == 0 {
		return fmt.Errorf("SRC_PREFIX must not be empty")
	}
	if len(opts.DstPrefix) == 0 {
		return fmt.Errorf("DST_PREFIX must not be empty")
	}
	if opts.SrcPrefix == opts.DstPrefix {
		return fmt.Errorf("SRC_PREFIX and DST_PREFIX must be different")
	}

	db, err := data.NewDB(false, dbConf)
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	fixTypes := resolveFixTypes(opts.FixType)
	if len(fixTypes) == 0 {
		return fmt.Errorf("unknown fix type %q, must be one of: %s, %s, %s, %s", opts.FixType, fixTypeAll, fixTypeBranding, fixTypeAvatar, fixTypePost)
	}

	results := make(map[string]*fixResult, len(fixTypes))
	state := newFixRunState(opts.Limit)

	for _, fixType := range fixTypes {
		if state.exhausted() {
			break
		}

		var result fixResult
		switch fixType {
		case fixTypeBranding:
			result, err = fixBranding(db, opts, state)
		case fixTypeAvatar:
			result, err = fixAvatar(db, opts, state)
		case fixTypePost:
			result, err = fixPost(db, opts, state)
		}
		if err != nil {
			return fmt.Errorf("fix %s failed: %w", fixType, err)
		}

		results[fixType] = &result
	}

	printFixSummary(fixTypes, results)

	if !opts.DryRun {
		invalidateSiteInfoCache(opts.CacheFilePath, fixTypes)
	}
	return nil
}

func resolveFixTypes(fixType string) []string {
	switch fixType {
	case fixTypeAll:
		return []string{fixTypeBranding, fixTypeAvatar, fixTypePost}
	case fixTypeBranding, fixTypeAvatar, fixTypePost:
		return []string{fixType}
	default:
		return nil
	}
}

func printFixSummary(fixTypes []string, results map[string]*fixResult) {
	fmt.Println()
	fmt.Printf("%-12s %10s %10s %10s\n", "TYPE", "AFFECTED", "FIXED", "SKIPPED")
	fmt.Println(strings.Repeat("-", 46))

	var total fixResult
	for _, fixType := range fixTypes {
		result := results[fixType]
		if result == nil {
			continue
		}
		fmt.Printf("%-12s %10d %10d %10d\n", fixType, result.Affected, result.Fixed, result.Skipped)
		total.add(*result)
	}

	fmt.Println(strings.Repeat("-", 46))
	fmt.Printf("%-12s %10d %10d %10d\n", "TOTAL", total.Affected, total.Fixed, total.Skipped)
}

// URL context markers used to anchor prefix matching/replacement to actual URL
// positions in text, preventing accidental rewrites of plain text content.
var (
	// textURLMarkers precede URLs in raw markdown [text](URL) and HTML attr="URL"
	textURLMarkers = []string{"](", `="`}
	// jsonURLMarkers precede URLs in JSON-encoded text (revision content stores
	// the post entity as JSON, so HTML quotes are escaped as \")
	jsonURLMarkers = []string{"](", `=\"`}
)

// containsURLPrefix reports whether text contains marker+prefix for any marker.
func containsURLPrefix(text, prefix string, markers []string) bool {
	for _, m := range markers {
		if strings.Contains(text, m+prefix) {
			return true
		}
	}
	return false
}

// prefixMatch classifies how stored data relates to the src/dst prefixes.
type prefixMatch int

const (
	matchNone prefixMatch = iota // neither prefix present → nothing to do
	matchFix                     // src prefix present → rewrite to dst
	matchSkip                    // already on dst prefix → skip
	matchBoth                    // both prefixes present → ambiguous, manual fix
)

// replaceURLPrefix replaces marker+src with marker+dst for each marker.
func replaceURLPrefix(text, src, dst string, markers []string) string {
	for _, m := range markers {
		text = strings.ReplaceAll(text, m+src, m+dst)
	}
	return text
}

// containsPrefixNotShadowed reports whether `prefix` occurs as a URL prefix
// that is not merely the head of the longer `shadow` prefix. Example: with
// prefix="/uploads/" and shadow="/uploads/new/", a URL "/uploads/new/x.png"
// does NOT count, but a standalone "/uploads/x.png" does.
func containsPrefixNotShadowed(text, prefix, shadow string, markers []string) bool {
	if !containsURLPrefix(text, prefix, markers) {
		return false
	}
	if len(shadow) > len(prefix) && strings.HasPrefix(shadow, prefix) {
		scrubbed := text
		for _, m := range markers {
			scrubbed = strings.ReplaceAll(scrubbed, m+shadow, "")
		}
		return containsURLPrefix(scrubbed, prefix, markers)
	}
	return true
}

func anyContainsPrefixNotShadowed(texts []string, prefix, shadow string, markers []string) bool {
	for _, t := range texts {
		if containsPrefixNotShadowed(t, prefix, shadow, markers) {
			return true
		}
	}
	return false
}

// classifyValue classifies a stored value whose entire content is the URL, so
// prefixes are matched at the start. It never returns matchBoth.
func classifyValue(val, src, dst string) prefixMatch {
	matchesSrc := strings.HasPrefix(val, src)
	matchesDst := strings.HasPrefix(val, dst)
	// When both match, prefer the longer prefix so a value already on dst is
	// skipped even if dst and src share a common head.
	if matchesDst && (!matchesSrc || len(dst) > len(src)) {
		return matchSkip
	}
	if matchesSrc {
		return matchFix
	}
	return matchNone
}

// classifyText classifies free text that may embed URLs at marker positions.
func classifyText(texts []string, src, dst string, markers []string) prefixMatch {
	hasSrc := anyContainsPrefixNotShadowed(texts, src, dst, markers)
	hasDst := anyContainsPrefixNotShadowed(texts, dst, src, markers)
	switch {
	case hasSrc && hasDst:
		return matchBoth
	case hasDst:
		return matchSkip
	case hasSrc:
		return matchFix
	default:
		return matchNone
	}
}

// fixRunState centralizes limit handling for all fix sub-routines.
type fixRunState struct {
	limit     int64
	remaining int64
}

func newFixRunState(limit int64) *fixRunState {
	return &fixRunState{limit: limit, remaining: limit}
}

func (s *fixRunState) exhausted() bool {
	return s.limit > 0 && s.remaining <= 0
}

func (s *fixRunState) markAffected(result *fixResult) {
	result.Affected++
	if s.limit > 0 {
		s.remaining--
	}
}

func applyValueFix(result *fixResult, opts *FixURLPrefixOptions, state *fixRunState,
	label, target, current string, update func(newValue string) error,
) error {
	switch classifyValue(current, opts.SrcPrefix, opts.DstPrefix) {
	case matchSkip:
		fmt.Printf("[%s] skip  %s: already has DST_PREFIX: %s\n", label, target, current)
		result.Skipped++
		return nil
	case matchNone:
		return nil
	}

	newValue := opts.DstPrefix + strings.TrimPrefix(current, opts.SrcPrefix)
	state.markAffected(result)
	fmt.Printf("[%s] %s: %s -> %s\n", label, target, current, newValue)

	if opts.DryRun {
		return nil
	}
	if err := update(newValue); err != nil {
		return err
	}
	result.Fixed++
	return nil
}

func runBatchedFix[T any](state *fixRunState, fetch func(offset int) ([]T, error), process func(item T) error) error {
	offset := 0
	for {
		if state.exhausted() {
			return nil
		}

		items, err := fetch(offset)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}

		for _, item := range items {
			if state.exhausted() {
				return nil
			}
			if err := process(item); err != nil {
				return err
			}
		}

		if len(items) < fixBatchSize {
			return nil
		}
		offset += fixBatchSize
	}
}
