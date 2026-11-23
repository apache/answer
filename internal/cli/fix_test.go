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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

const (
	testSrc = "https://old.example.com/uploads/"
	testDst = "https://new.example.com/uploads/"
)

func TestClassifyText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		src, dst string
		want     prefixMatch
	}{
		{"none", "plain text https://old.example.com/uploads/x.png", testSrc, testDst, matchNone},
		{"markdown src", "![](https://old.example.com/uploads/a.png)", testSrc, testDst, matchFix},
		{"html dst", `<img src="https://new.example.com/uploads/a.png">`, testSrc, testDst, matchSkip},
		{"both, disjoint prefixes", "[a](https://old.example.com/uploads/a) [b](https://new.example.com/uploads/b)", testSrc, testDst, matchFix},
		{"dst nested under src is not src", "![](/uploads/new/a.png)", "/uploads/", "/uploads/new/", matchSkip},
		{"src nested under dst is not dst", "![](/uploads/new/a.png)", "/uploads/new/", "/uploads/", matchFix},
		{"both, dst starts with src", "![](/uploads/a.png) ![](/uploads/new/b.png)", "/uploads/", "/uploads/new/", matchBoth},
		{"both, src starts with dst", "![](/uploads/new/a.png) ![](/uploads/b.png)", "/uploads/new/", "/uploads/", matchFix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, classifyText(tt.text, tt.src, tt.dst, textURLMarkers))
		})
	}
}

func TestClassifyFields(t *testing.T) {
	src := "![](" + testSrc + "a.png)"
	dst := "![](" + testDst + "a.png)"
	const nestedSrc, nestedDst = "/uploads/", "/uploads/cdn/"
	nsrc := "![](" + nestedSrc + "a.png)"
	ndst := "![](" + nestedDst + "a.png)"
	tests := []struct {
		name     string
		fields   []string
		src, dst string
		want     prefixMatch
	}{
		{"none", []string{"x", "y"}, testSrc, testDst, matchNone},
		{"all dst", []string{dst, dst}, testSrc, testDst, matchSkip},
		{"src and dst in different fields", []string{src, dst}, testSrc, testDst, matchFix},
		{"dst then src", []string{dst, src}, testSrc, testDst, matchFix},
		{"both in one field, disjoint prefixes", []string{src + dst, dst}, testSrc, testDst, matchFix},
		{"both in one field, nested prefixes", []string{nsrc + ndst, nsrc}, nestedSrc, nestedDst, matchBoth},
		{"both in a later field, nested prefixes", []string{nsrc, ndst + nsrc}, nestedSrc, nestedDst, matchBoth},
		{"nested prefixes across fields", []string{nsrc, ndst}, nestedSrc, nestedDst, matchFix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, classifyFields(tt.fields, tt.src, tt.dst, textURLMarkers))
		})
	}
}

func TestClassifyValue(t *testing.T) {
	assert.Equal(t, matchFix, classifyValue(testSrc+"a.png", testSrc, testDst))
	assert.Equal(t, matchSkip, classifyValue(testDst+"a.png", testSrc, testDst))
	assert.Equal(t, matchNone, classifyValue("https://other/a.png", testSrc, testDst))
	assert.Equal(t, matchSkip, classifyValue("/uploads/new/a.png", "/uploads/", "/uploads/new/"))
	assert.Equal(t, matchFix, classifyValue("/uploads/new/a.png", "/uploads/new/", "/uploads/"))
}

func newFixTestDB(t *testing.T) *xorm.Engine {
	t.Helper()
	x, err := xorm.NewEngine("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = x.Close() })
	require.NoError(t, x.Sync(new(entity.Question), new(entity.Answer), new(entity.Tag), new(entity.Comment),
		new(entity.Revision), new(entity.FileRecord), new(entity.User), new(entity.SiteInfo), new(entity.Notification)))
	return x
}

func TestFixPost(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}

	// Prefixes mixed across fields are fixed field by field.
	_, err := x.Insert(&entity.Question{ID: "1",
		OriginalText: "![](" + testSrc + "q.png)",
		ParsedText:   `<img src="` + testDst + `q.png">`})
	require.NoError(t, err)
	// Both prefixes in one field are fixed, since the prefixes do not overlap.
	_, err = x.Insert(&entity.Answer{ID: "5",
		OriginalText: "![](" + testSrc + "a.png) ![](" + testDst + "b.png)",
		ParsedText:   `<img src="` + testSrc + `a.png">`})
	require.NoError(t, err)
	_, err = x.Insert(&entity.Tag{ID: "2", SlugName: "go",
		OriginalText: "![](" + testSrc + "t.png)",
		ParsedText:   `<img src="` + testSrc + `t.png">`})
	require.NoError(t, err)

	tagObjectType := constant.ObjectTypeStrMapping[constant.TagObjectType]
	migratable, _ := json.Marshal(map[string]any{
		"ID": "2", "OriginalText": "![](" + testSrc + "t.png)", "ParsedText": `<img src="` + testSrc + `t.png">`})
	mixed, _ := json.Marshal(map[string]any{
		"ID": "3", "OriginalText": "![](" + testSrc + "keep.png) ![](" + testDst + "other.png)"})
	_, err = x.Insert(&entity.Revision{ID: "11", ObjectType: tagObjectType, ObjectID: "2", Content: string(migratable)})
	require.NoError(t, err)
	_, err = x.Insert(&entity.Revision{ID: "12", ObjectType: tagObjectType, ObjectID: "3", Content: string(mixed)})
	require.NoError(t, err)

	_, err = x.Insert(&entity.FileRecord{ID: 1, FileURL: testSrc + "t.png", ObjectID: "0", Status: entity.FileRecordStatusAvailable, Source: string(plugin.UserPost)})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 2, FileURL: testSrc + "keep.png", ObjectID: "0", Status: entity.FileRecordStatusAvailable, Source: string(plugin.UserPost)})
	require.NoError(t, err)
	// Linked records are not checked against revisions.
	_, err = x.Insert(&entity.FileRecord{ID: 3, FileURL: testSrc + "linked.png", ObjectID: "100", Source: string(plugin.UserPost)})
	require.NoError(t, err)
	// Deleted records are not checked against revisions either.
	_, err = x.Insert(&entity.FileRecord{ID: 4, FileURL: testSrc + "deleted.png", ObjectID: "0",
		Status: entity.FileRecordStatusDeleted, Source: string(plugin.UserPost)})
	require.NoError(t, err)

	rows, files, err := fixPost(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, int64(4), rows.Fixed) // question 1, answer 5, tag 2 with revision 11, orphan revision 12
	assert.Equal(t, int64(0), rows.Skipped)
	assert.Equal(t, int64(4), files.Fixed) // file_record 1 to 4
	assert.Equal(t, int64(0), files.Skipped)

	q := &entity.Question{}
	_, err = x.ID("1").Get(q)
	require.NoError(t, err)
	assert.Equal(t, "![]("+testDst+"q.png)", q.OriginalText)
	assert.Equal(t, `<img src="`+testDst+`q.png">`, q.ParsedText)

	a := &entity.Answer{}
	_, err = x.ID("5").Get(a)
	require.NoError(t, err)
	assert.Equal(t, "![]("+testDst+"a.png) ![]("+testDst+"b.png)", a.OriginalText)
	assert.Equal(t, `<img src="`+testDst+`a.png">`, a.ParsedText)

	tag := &entity.Tag{}
	_, err = x.ID("2").Get(tag)
	require.NoError(t, err)
	assert.Equal(t, "![]("+testDst+"t.png)", tag.OriginalText)
	assert.Equal(t, `<img src="`+testDst+`t.png">`, tag.ParsedText)

	rev := &entity.Revision{}
	_, err = x.ID("11").Get(rev)
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, json.Unmarshal([]byte(rev.Content), &doc))
	assert.Equal(t, "2", doc["ID"])
	assert.Equal(t, "![]("+testDst+"t.png)", doc["OriginalText"])
	assert.Equal(t, `<img src="`+testDst+`t.png">`, doc["ParsedText"])

	rev = &entity.Revision{}
	_, err = x.ID("12").Get(rev)
	require.NoError(t, err)
	doc = map[string]any{}
	require.NoError(t, json.Unmarshal([]byte(rev.Content), &doc))
	assert.Equal(t, "![]("+testDst+"keep.png) ![]("+testDst+"other.png)", doc["OriginalText"])

	assertFileURL(t, x, 1, testDst+"t.png")
	// Referenced by the fixed mixed revision, so it follows.
	assertFileURL(t, x, 2, testDst+"keep.png")
	assertFileURL(t, x, 3, testDst+"linked.png")
	assertFileURL(t, x, 4, testDst+"deleted.png")
}

// assertFileURL asserts the file_url of file record id.
func assertFileURL(t *testing.T, x *xorm.Engine, id int, want string) {
	t.Helper()
	fr := &entity.FileRecord{}
	_, err := x.ID(id).Get(fr)
	require.NoError(t, err)
	assert.Equal(t, want, fr.FileURL)
}

func TestFixAvatarLimitKeepsFileRecordsInSync(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}

	for i, name := range []string{"a.png", "b.png"} {
		_, err := x.Insert(&entity.User{ID: fmt.Sprint(i + 1), Username: name,
			Avatar: `{"type":"custom","custom":"` + testSrc + name + `"}`})
		require.NoError(t, err)
		_, err = x.Insert(&entity.FileRecord{ID: i + 1, FileURL: testSrc + name, Source: string(plugin.UserAvatar)})
		require.NoError(t, err)
	}

	result, err := fixAvatar(x, opts, newFixRunState(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Fixed)

	// Only the migrated user's file record follows; the other stays on the old URL.
	assertFileURL(t, x, 1, testDst+"a.png")
	assertFileURL(t, x, 2, testSrc+"b.png")
}

func TestFixAvatarRepairsStaleFileRecord(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}

	// The avatar was already moved (e.g. by hand), but its file record was not.
	_, err := x.Insert(&entity.User{ID: "1", Username: "u1",
		Avatar: `{"type":"custom","custom":"` + testDst + `a.png"}`})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 1, FileURL: testSrc + "a.png", Source: string(plugin.UserAvatar)})
	require.NoError(t, err)

	_, err = fixAvatar(x, &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst, DryRun: true}, newFixRunState(0))
	require.NoError(t, err)
	assertFileURL(t, x, 1, testSrc+"a.png")

	result, err := fixAvatar(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Fixed)
	assert.Equal(t, int64(1), result.Skipped)
	assertFileURL(t, x, 1, testDst+"a.png")
}

func TestFixBrandingLimitKeepsFileRecordsInSync(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}

	// logo and mobile_logo share one file record.
	_, err := x.Insert(&entity.SiteInfo{Type: constant.SiteTypeBranding,
		Content: `{"logo":"` + testSrc + `logo.png","mobile_logo":"` + testSrc + `logo.png","favicon":"` + testSrc + `favicon.png"}`})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 1, FileURL: testSrc + "logo.png", Source: string(plugin.AdminBranding)})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 2, FileURL: testSrc + "favicon.png", Source: string(plugin.AdminBranding)})
	require.NoError(t, err)

	result, err := fixBranding(x, opts, newFixRunState(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Fixed)
	// mobile_logo still uses the old URL, so the shared record must keep it.
	assertFileURL(t, x, 1, testSrc+"logo.png")
	assertFileURL(t, x, 2, testSrc+"favicon.png")

	result, err = fixBranding(x, opts, newFixRunState(1))
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Fixed)
	assertFileURL(t, x, 1, testDst+"logo.png")
	assertFileURL(t, x, 2, testSrc+"favicon.png")
}

func TestFixPostDryRunMatchesRealRun(t *testing.T) {
	x := newFixTestDB(t)

	content, _ := json.Marshal(map[string]any{"ID": "1", "OriginalText": "![](" + testSrc + "a.png)"})
	_, err := x.Insert(&entity.Revision{ID: "11", ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		ObjectID: "1", Content: string(content)})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 1, FileURL: testSrc + "a.png", ObjectID: "0", Status: entity.FileRecordStatusAvailable, Source: string(plugin.UserPost)})
	require.NoError(t, err)

	// A revision mixing disjoint prefixes is fixed, so a dry run must count its file record as referenced.
	mixed, _ := json.Marshal(map[string]any{
		"ID": "2", "OriginalText": "![](" + testSrc + "b.png) ![](" + testDst + "c.png)"})
	_, err = x.Insert(&entity.Revision{ID: "12", ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		ObjectID: "2", Content: string(mixed)})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 2, FileURL: testSrc + "b.png", ObjectID: "0", Status: entity.FileRecordStatusAvailable, Source: string(plugin.UserPost)})
	require.NoError(t, err)

	// d.png is a bare URL, so fixing the revision's other link does not rewrite it.
	bare, _ := json.Marshal(map[string]any{
		"ID": "3", "OriginalText": "![](" + testSrc + "e.png) " + testSrc + "d.png"})
	_, err = x.Insert(&entity.Revision{ID: "13", ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		ObjectID: "3", Content: string(bare)})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 3, FileURL: testSrc + "d.png", ObjectID: "0", Status: entity.FileRecordStatusAvailable, Source: string(plugin.UserPost)})
	require.NoError(t, err)

	dryRows, dryFiles, err := fixPost(x, &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst, DryRun: true}, newFixRunState(0))
	require.NoError(t, err)
	assertFileURL(t, x, 1, testSrc+"a.png")

	rows, files, err := fixPost(x, &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, rows.Affected, dryRows.Affected)
	assert.Equal(t, rows.Skipped, dryRows.Skipped)
	assert.Equal(t, files.Affected, dryFiles.Affected)
	assert.Equal(t, files.Skipped, dryFiles.Skipped)
	assertFileURL(t, x, 1, testDst+"a.png")
	assertFileURL(t, x, 2, testDst+"b.png")
	assertFileURL(t, x, 3, testSrc+"d.png")
}

func TestFixPostSkipsBothWithOverlappingPrefixes(t *testing.T) {
	x := newFixTestDB(t)
	const src, dst = "/uploads/", "/uploads/cdn/"

	// Rewriting src here would also turn /uploads/cdn/ into /uploads/cdn/cdn/.
	mixedText := "![](" + src + "a.png) ![](" + dst + "b.png)"
	_, err := x.Insert(&entity.Answer{ID: "5", OriginalText: mixedText, ParsedText: `<img src="` + src + `a.png">`})
	require.NoError(t, err)
	mixed, _ := json.Marshal(map[string]any{"ID": "5", "OriginalText": mixedText})
	_, err = x.Insert(&entity.Revision{ID: "12", ObjectType: constant.ObjectTypeStrMapping[constant.AnswerObjectType],
		ObjectID: "5", Content: string(mixed)})
	require.NoError(t, err)
	// Referenced only by the skipped revision, so it must keep the old URL.
	_, err = x.Insert(&entity.FileRecord{ID: 1, FileURL: src + "a.png", ObjectID: "0", Status: entity.FileRecordStatusAvailable, Source: string(plugin.UserPost)})
	require.NoError(t, err)

	dryRows, dryFiles, err := fixPost(x, &FixURLPrefixOptions{SrcPrefix: src, DstPrefix: dst, DryRun: true}, newFixRunState(0))
	require.NoError(t, err)
	rows, files, err := fixPost(x, &FixURLPrefixOptions{SrcPrefix: src, DstPrefix: dst}, newFixRunState(0))
	require.NoError(t, err)

	assert.Equal(t, fixResult{Skipped: 1}, rows) // answer 5 with revision 12
	assert.Equal(t, fixResult{Skipped: 1}, files)
	assert.Equal(t, rows, dryRows)
	assert.Equal(t, files, dryFiles)

	a := &entity.Answer{}
	_, err = x.ID("5").Get(a)
	require.NoError(t, err)
	assert.Equal(t, mixedText, a.OriginalText)
	rev := &entity.Revision{}
	_, err = x.ID("12").Get(rev)
	require.NoError(t, err)
	assert.Equal(t, string(mixed), rev.Content)
	assertFileURL(t, x, 1, src+"a.png")
}

func TestFixAvatarSkipsInvalidJSON(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}

	// Matches the custom-avatar query but cannot be parsed.
	invalid := `{"custom":"` + testSrc + `a.png"`
	_, err := x.Insert(&entity.User{ID: "1", Username: "u1", Avatar: invalid})
	require.NoError(t, err)
	// A valid row after the invalid one must still be fixed.
	_, err = x.Insert(&entity.User{ID: "2", Username: "u2",
		Avatar: `{"type":"custom","custom":"` + testSrc + `b.png"}`})
	require.NoError(t, err)

	result, err := fixAvatar(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Fixed)
	assert.Equal(t, int64(1), result.Skipped)

	u := &entity.User{}
	_, err = x.ID("1").Get(u)
	require.NoError(t, err)
	assert.Equal(t, invalid, u.Avatar)

	u = &entity.User{}
	_, err = x.ID("2").Get(u)
	require.NoError(t, err)
	// Stored in the same format the app writes.
	assert.Equal(t, schema.CustomAvatar(testDst+"b.png").ToJsonString(), u.Avatar)
}

func TestFixPostMixedFieldsWithNestedDst(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: "/uploads/", DstPrefix: "/uploads/cdn/"}

	// parsed_text is already on DST, which starts with SRC; it must not be rewritten again.
	_, err := x.Insert(&entity.Question{ID: "1",
		OriginalText: "![](/uploads/q.png)",
		ParsedText:   `<img src="/uploads/cdn/q.png">`})
	require.NoError(t, err)

	rows, _, err := fixPost(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows.Fixed)

	q := &entity.Question{}
	_, err = x.ID("1").Get(q)
	require.NoError(t, err)
	assert.Equal(t, "![](/uploads/cdn/q.png)", q.OriginalText)
	assert.Equal(t, `<img src="/uploads/cdn/q.png">`, q.ParsedText)
}

func TestFixPostLimitStillSyncsFileRecords(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst, Limit: 1}

	content, _ := json.Marshal(map[string]any{"ID": "1", "OriginalText": "![](" + testSrc + "a.png)"})
	_, err := x.Insert(&entity.Revision{ID: "11", ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		ObjectID: "1", Content: string(content)})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 1, FileURL: testSrc + "a.png", ObjectID: "0", Status: entity.FileRecordStatusAvailable, Source: string(plugin.UserPost)})
	require.NoError(t, err)

	state := newFixRunState(opts.Limit)
	rows, files, err := fixPost(x, opts, state)
	require.NoError(t, err)
	assert.True(t, state.exhausted())
	// The file record is reported apart from the rows --limit counts.
	assert.Equal(t, int64(1), rows.Affected)
	assert.Equal(t, int64(1), files.Affected)
	assertFileURL(t, x, 1, testDst+"a.png")
}

// insertQuestionWithRevision inserts a question and one revision of it, both
// linking name under prefix.
func insertQuestionWithRevision(t *testing.T, x *xorm.Engine, id, revID, prefix, name string) {
	t.Helper()
	_, err := x.Insert(&entity.Question{ID: id,
		OriginalText: "![](" + prefix + name + ")",
		ParsedText:   `<img src="` + prefix + name + `">`})
	require.NoError(t, err)
	content, _ := json.Marshal(map[string]any{"ID": id, "OriginalText": "![](" + prefix + name + ")"})
	_, err = x.Insert(&entity.Revision{ID: revID, ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		ObjectID: id, Content: string(content)})
	require.NoError(t, err)
}

// assertRevisionHas asserts that the content of revision id contains want.
func assertRevisionHas(t *testing.T, x *xorm.Engine, id, want string) {
	t.Helper()
	rev := &entity.Revision{}
	_, err := x.ID(id).Get(rev)
	require.NoError(t, err)
	assert.Contains(t, rev.Content, want)
}

func TestFixPostLimitFixesPostWithItsRevisions(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst, Limit: 1}
	insertQuestionWithRevision(t, x, "1", "11", testSrc, "a.png")
	insertQuestionWithRevision(t, x, "2", "12", testSrc, "b.png")

	rows, _, err := fixPost(x, opts, newFixRunState(opts.Limit))
	require.NoError(t, err)
	assert.Equal(t, fixResult{Affected: 1, Fixed: 1}, rows)

	// The post within the limit is fixed with its revision; the other keeps both on SRC.
	q := &entity.Question{}
	_, err = x.ID("1").Get(q)
	require.NoError(t, err)
	assert.Equal(t, "![]("+testDst+"a.png)", q.OriginalText)
	assertRevisionHas(t, x, "11", "]("+testDst+"a.png)")

	q = &entity.Question{}
	_, err = x.ID("2").Get(q)
	require.NoError(t, err)
	assert.Equal(t, "![]("+testSrc+"b.png)", q.OriginalText)
	assertRevisionHas(t, x, "12", "]("+testSrc+"b.png)")
}

func TestFixPostFixesOldRevisionsOfPostOnDst(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}
	// The post is already on DST, but one of its revisions is not.
	insertQuestionWithRevision(t, x, "1", "11", testDst, "a.png")
	old, _ := json.Marshal(map[string]any{"ID": "1", "OriginalText": "![](" + testSrc + "old.png)"})
	_, err := x.Insert(&entity.Revision{ID: "10", ObjectType: constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		ObjectID: "1", Content: string(old)})
	require.NoError(t, err)

	rows, _, err := fixPost(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, fixResult{Affected: 1, Fixed: 1}, rows)
	assertRevisionHas(t, x, "10", "]("+testDst+"old.png)")
	assertRevisionHas(t, x, "11", "]("+testDst+"a.png)")
}

func TestFixPostPagesByID(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}
	// More than two pages; ids 1..n also order differently as strings ("10" < "9").
	n := 2*fixBatchSize + 1
	for i := 1; i <= n; i++ {
		insertQuestionWithRevision(t, x, fmt.Sprint(i), fmt.Sprint(1000+i), testSrc, fmt.Sprintf("%d.png", i))
	}

	rows, _, err := fixPost(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, fixResult{Affected: int64(n), Fixed: int64(n)}, rows)

	count, err := x.Where("original_text LIKE ?", "%"+testSrc+"%").Count(new(entity.Question))
	require.NoError(t, err)
	assert.Zero(t, count)
	count, err = x.Where("content LIKE ?", "%"+testSrc+"%").Count(new(entity.Revision))
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestFixAvatarFixesCustomURLOfOtherType(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}

	// The user switched to gravatar; the old custom URL is kept for switching back.
	_, err := x.Insert(&entity.User{ID: "1", Username: "u1",
		Avatar: `{"type":"gravatar","gravatar":"https://gravatar.example/x","custom":"` + testSrc + `a.png"}`})
	require.NoError(t, err)
	_, err = x.Insert(&entity.FileRecord{ID: 1, FileURL: testSrc + "a.png", Source: string(plugin.UserAvatar)})
	require.NoError(t, err)

	result, err := fixAvatar(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Fixed)

	u := &entity.User{}
	_, err = x.ID("1").Get(u)
	require.NoError(t, err)
	var avatar schema.AvatarInfo
	require.NoError(t, json.Unmarshal([]byte(u.Avatar), &avatar))
	assert.Equal(t, schema.AvatarInfo{Type: "gravatar", Gravatar: "https://gravatar.example/x", Custom: testDst + "a.png"}, avatar)
	assertFileURL(t, x, 1, testDst+"a.png")
}

func TestFixPostFixesComments(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}
	_, err := x.Insert(&entity.Comment{ID: "1", UserID: "1", ObjectID: "1",
		OriginalText: "see ![](" + testSrc + "c.png)",
		ParsedText:   `<p>see <img src="` + testSrc + `c.png"></p>`})
	require.NoError(t, err)

	rows, _, err := fixPost(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, fixResult{Affected: 1, Fixed: 1}, rows)

	c := &entity.Comment{}
	_, err = x.ID("1").Get(c)
	require.NoError(t, err)
	assert.Equal(t, "see ![]("+testDst+"c.png)", c.OriginalText)
	assert.Equal(t, `<p>see <img src="`+testDst+`c.png"></p>`, c.ParsedText)
}

func TestFixAvatarFixesNotificationAvatars(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}

	// Encoded the way the app stores notification content.
	notificationContent := func(avatar string) string {
		content, _ := json.Marshal(&schema.NotificationContent{
			UserInfo:           &schema.UserBasicInfo{ID: "7", Username: "u7", Avatar: avatar},
			ObjectInfo:         schema.ObjectInfo{Title: "q", ObjectID: "1", ObjectType: "question"},
			NotificationAction: "notification.action.answered_question",
		})
		return string(content)
	}
	for id, avatar := range map[string]string{
		"1": testSrc + "a.png",
		"2": testDst + "a.png",
		"3": "https://gravatar.example/x",
	} {
		_, err := x.Insert(&entity.Notification{ID: id, UserID: "8", ObjectID: "1", Content: notificationContent(avatar)})
		require.NoError(t, err)
	}

	result, err := fixAvatar(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, fixResult{Affected: 1, Fixed: 1}, result)

	for id, want := range map[string]string{
		"1": notificationContent(testDst + "a.png"),
		"2": notificationContent(testDst + "a.png"),
		"3": notificationContent("https://gravatar.example/x"),
	} {
		n := &entity.Notification{}
		_, err = x.ID(id).Get(n)
		require.NoError(t, err)
		assert.Equal(t, want, n.Content, "notification %s", id)
	}
}

func TestFixPostRollsBackUnitOnFailedWrite(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: testSrc, DstPrefix: testDst}
	insertQuestionWithRevision(t, x, "1", "11", testSrc, "a.png")
	// The post row is written first; failing its revision must undo that write.
	_, err := x.Exec("CREATE TRIGGER fail_revision BEFORE UPDATE ON revision BEGIN SELECT RAISE(ABORT, 'boom'); END")
	require.NoError(t, err)

	_, _, err = fixPost(x, opts, newFixRunState(0))
	require.ErrorContains(t, err, "boom")

	q := &entity.Question{}
	_, err = x.ID("1").Get(q)
	require.NoError(t, err)
	assert.Equal(t, "![]("+testSrc+"a.png)", q.OriginalText)
	assertRevisionHas(t, x, "11", "]("+testSrc+"a.png)")
}

func TestFixPostMixedFieldWhenSrcStartsWithDst(t *testing.T) {
	x := newFixTestDB(t)
	opts := &FixURLPrefixOptions{SrcPrefix: "/uploads/cdn/", DstPrefix: "/uploads/"}
	// Replacing /uploads/cdn/ cannot touch a URL that is only /uploads/.
	_, err := x.Insert(&entity.Answer{ID: "5",
		OriginalText: "![a](/uploads/cdn/a.png) ![b](/uploads/b.png)",
		ParsedText:   `<img src="/uploads/cdn/a.png"><img src="/uploads/b.png">`})
	require.NoError(t, err)

	rows, _, err := fixPost(x, opts, newFixRunState(0))
	require.NoError(t, err)
	assert.Equal(t, fixResult{Affected: 1, Fixed: 1}, rows)

	a := &entity.Answer{}
	_, err = x.ID("5").Get(a)
	require.NoError(t, err)
	assert.Equal(t, "![a](/uploads/a.png) ![b](/uploads/b.png)", a.OriginalText)
	assert.Equal(t, `<img src="/uploads/a.png"><img src="/uploads/b.png">`, a.ParsedText)
}

func TestFixURLPrefixRejectsMismatchedTrailingSlash(t *testing.T) {
	err := FixURLPrefix(nil, &FixURLPrefixOptions{FixType: fixTypeAll, SrcPrefix: "/uploads/", DstPrefix: "/cdn"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `end with "/"`)
}

func TestFixURLPrefixRejectsNegativeLimit(t *testing.T) {
	err := FixURLPrefix(nil, &FixURLPrefixOptions{FixType: fixTypeAll, SrcPrefix: testSrc, DstPrefix: testDst, Limit: -1})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "limit must not be negative")
}

func TestFixURLPrefixRejectsUnknownTypeBeforeConnecting(t *testing.T) {
	// A nil database config would fail to connect, so this only passes if the type is checked first.
	err := FixURLPrefix(nil, &FixURLPrefixOptions{FixType: "avtar", SrcPrefix: testSrc, DstPrefix: testDst})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown fix type")
}
