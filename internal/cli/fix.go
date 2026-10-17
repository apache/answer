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
	"strconv"
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
	// Limit caps fixed rows across all types (a post with its revisions is one
	// row; file records are not counted). 0 means unlimited.
	Limit int64
	// CacheFilePath is the file cache whose branding entry is cleared, if set.
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
	// A slash on one side only joins path segments: /uploads/a.png -> /cdna.png.
	if strings.HasSuffix(opts.SrcPrefix, "/") != strings.HasSuffix(opts.DstPrefix, "/") {
		return fmt.Errorf("SRC_PREFIX and DST_PREFIX must both end with \"/\" or both not")
	}
	if opts.Limit < 0 {
		return fmt.Errorf("limit must not be negative (0 means unlimited), got %d", opts.Limit)
	}
	fixTypes := resolveFixTypes(opts.FixType)
	if len(fixTypes) == 0 {
		return fmt.Errorf("unknown fix type %q, must be one of: %s, %s, %s, %s", opts.FixType, fixTypeAll, fixTypeBranding, fixTypeAvatar, fixTypePost)
	}

	db, err := data.NewDB(false, dbConf)
	if err != nil {
		return fmt.Errorf("connect database failed: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if !strings.HasSuffix(opts.SrcPrefix, "/") {
		fmt.Printf("WARN: SRC_PREFIX does not end with \"/\", so it also matches longer paths (e.g. %q matches %q)\n",
			opts.SrcPrefix, opts.SrcPrefix+"-other/x.png")
	}
	if strings.HasPrefix(opts.SrcPrefix, opts.DstPrefix) {
		fmt.Printf("WARN: SRC_PREFIX starts with DST_PREFIX, so URLs already moved under %q cannot be told apart from old ones; do not rerun after a successful run\n",
			opts.SrcPrefix)
	}
	if opts.DryRun && opts.Limit > 0 {
		fmt.Println("NOTE: the preview counts post file records as if --limit did not apply, so a real run may fix fewer of them")
	}

	results := make(map[string]*fixResult, len(fixTypes))
	var postFiles *fixResult
	var notStarted []string
	state := newFixRunState(opts.Limit)

	if !opts.DryRun {
		// Deferred so it also runs when a later fix type fails.
		defer invalidateSiteInfoCache(opts.CacheFilePath, fixTypes)
	}

	for i, fixType := range fixTypes {
		if state.exhausted() {
			notStarted = fixTypes[i:]
			break
		}

		var result fixResult
		switch fixType {
		case fixTypeBranding:
			result, err = fixBranding(db, opts, state)
		case fixTypeAvatar:
			result, err = fixAvatar(db, opts, state)
		case fixTypePost:
			var files fixResult
			result, files, err = fixPost(db, opts, state)
			postFiles = &files
		}
		if err != nil {
			return fmt.Errorf("fix %s failed: %w", fixType, err)
		}

		results[fixType] = &result
	}

	printFixSummary(fixTypes, results, postFiles)
	if state.exhausted() {
		if opts.DryRun {
			fmt.Printf("\n--limit %d reached in preview; a real run would stop here too.\n", opts.Limit)
		} else {
			fmt.Printf("\n--limit %d reached; run the command again to fix the rest.\n", opts.Limit)
		}
		if len(notStarted) > 0 {
			fmt.Printf("Not started: %s\n", strings.Join(notStarted, ", "))
		}
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

// printFixSummary prints the rows counted by --limit, then post file records.
func printFixSummary(fixTypes []string, results map[string]*fixResult, postFiles *fixResult) {
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

	if postFiles != nil {
		fmt.Println()
		fmt.Println("File records (not counted by --limit):")
		fmt.Printf("%-12s %10d %10d %10d\n", fixTypePost, postFiles.Affected, postFiles.Fixed, postFiles.Skipped)
	}
}

// URL markers anchor prefix matching to URL positions, so plain text is kept.
var (
	// textURLMarkers precede URLs in markdown [text](URL) and HTML attr="URL".
	textURLMarkers = []string{"](", `="`}
	// jsonURLMarkers are the same in JSON-encoded revision content.
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

// prefixMatch is how stored data relates to the src/dst prefixes.
type prefixMatch int

const (
	matchNone prefixMatch = iota // neither prefix
	matchFix                     // has src: rewrite
	matchSkip                    // only dst: skip
	matchBoth                    // ambiguous: manual fix
)

// replaceURLPrefix replaces marker+src with marker+dst for each marker.
func replaceURLPrefix(text, src, dst string, markers []string) string {
	for _, m := range markers {
		text = strings.ReplaceAll(text, m+src, m+dst)
	}
	return text
}

// swapPrefix replaces the leading from prefix of url with to.
func swapPrefix(url, from, to string) string {
	return to + strings.TrimPrefix(url, from)
}

// containsPrefixNotShadowed reports whether prefix occurs other than as the
// head of a longer shadow: for /uploads/ shadowed by /uploads/new/,
// /uploads/x.png counts but /uploads/new/x.png does not.
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

// classifyValue classifies a value that is a whole URL. It never returns
// matchBoth.
func classifyValue(val, src, dst string) prefixMatch {
	matchesSrc := strings.HasPrefix(val, src)
	matchesDst := strings.HasPrefix(val, dst)
	// When both match, the longer prefix wins.
	if matchesDst && (!matchesSrc || len(dst) > len(src)) {
		return matchSkip
	}
	if matchesSrc {
		return matchFix
	}
	return matchNone
}

// classifyText classifies text embedding URLs at marker positions. Text with
// both prefixes is ambiguous only if dst starts with src, as replacing src
// would then nest dst URLs (/uploads/cdn/ -> /uploads/cdn/cdn/).
func classifyText(text, src, dst string, markers []string) prefixMatch {
	hasSrc := containsPrefixNotShadowed(text, src, dst, markers)
	hasDst := containsPrefixNotShadowed(text, dst, src, markers)
	switch {
	case hasSrc && hasDst:
		if strings.HasPrefix(dst, src) {
			return matchBoth
		}
		return matchFix
	case hasDst:
		return matchSkip
	case hasSrc:
		return matchFix
	default:
		return matchNone
	}
}

// classifyFields combines per-field results: any ambiguous field wins, then
// any fixable one. Callers rewrite only matchFix fields (see fixTextField).
func classifyFields(texts []string, src, dst string, markers []string) prefixMatch {
	result := matchNone
	for _, text := range texts {
		switch classifyText(text, src, dst, markers) {
		case matchBoth:
			return matchBoth
		case matchFix:
			result = matchFix
		case matchSkip:
			if result == matchNone {
				result = matchSkip
			}
		}
	}
	return result
}

// fixRunState tracks --limit across all fix types.
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

	newValue := swapPrefix(current, opts.SrcPrefix, opts.DstPrefix)
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

// runBatchedFix pages by id rather than OFFSET. fetch returns rows with id >
// afterID in id order, and the last row's id.
func runBatchedFix[T any](state *fixRunState, fetch func(afterID int64) (items []T, lastID int64, err error), process func(item T) error) error {
	var afterID int64
	for {
		if state.exhausted() {
			return nil
		}

		items, lastID, err := fetch(afterID)
		if err != nil {
			return err
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
		afterID = lastID
	}
}

// parseID parses a BIGINT id held as a string, for numeric SQL comparison.
func parseID(id string) (int64, error) {
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse id %q failed: %w", id, err)
	}
	return n, nil
}

// lastStringID returns the last item's parsed id, or 0 if there is none.
func lastStringID[T any](items []T, id func(T) string) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}
	return parseID(id(items[len(items)-1]))
}
