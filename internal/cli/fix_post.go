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

func fixPost(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState) (fixResult, error) {
	result := fixResult{}

	// Fix question and answer body text.
	for _, tableName := range []string{"question", "answer"} {
		r, err := fixPostTable(x, opts, state, tableName)
		if err != nil {
			return fixResult{}, err
		}
		result.add(r)
	}

	// Fix revisions for questions and answers.
	for _, objectType := range []int{
		constant.ObjectTypeStrMapping[constant.QuestionObjectType],
		constant.ObjectTypeStrMapping[constant.AnswerObjectType],
	} {
		r, err := fixRevisionByObjectType(x, opts, state, objectType, fixTypePost)
		if err != nil {
			return fixResult{}, err
		}
		result.add(r)
	}

	// Fix file records.
	fileResult, err := fixFileRecordBySources(x, opts, state,
		[]string{string(plugin.UserPost), string(plugin.UserPostAttachment)}, fixTypePost)
	if err != nil {
		return fixResult{}, err
	}
	result.add(fileResult)

	return result, nil
}

// postRow represents a question or answer row with its text fields.
type postRow struct {
	ID           string `xorm:"id"`
	OriginalText string `xorm:"original_text"`
	ParsedText   string `xorm:"parsed_text"`
}

func fixPostTable(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState, tableName string) (fixResult, error) {
	result := fixResult{}
	err := runBatchedFix(state,
		func(offset int) ([]*postRow, error) { return queryPostRows(x, tableName, offset) },
		func(row *postRow) error {
			return applyTextFix(&result, opts, state, fixTypePost, fmt.Sprintf("%s id=%s", tableName, row.ID),
				[]string{row.OriginalText, row.ParsedText}, textURLMarkers,
				func() error { return updatePostText(x, tableName, row, opts) })
		})
	if err != nil {
		return fixResult{}, err
	}
	return result, nil
}

func queryPostRows(x *xorm.Engine, tableName string, offset int) ([]*postRow, error) {
	rows := make([]*postRow, 0)
	err := x.Table(tableName).Select("id, original_text, parsed_text").
		OrderBy("id").
		Limit(fixBatchSize, offset).
		Find(&rows)
	if err != nil {
		return nil, fmt.Errorf("query %s failed: %w", tableName, err)
	}
	return rows, nil
}

func updatePostText(x *xorm.Engine, tableName string, row *postRow, opts *FixURLPrefixOptions) error {
	origText := replaceURLPrefix(row.OriginalText, opts.SrcPrefix, opts.DstPrefix, textURLMarkers)
	parsedText := replaceURLPrefix(row.ParsedText, opts.SrcPrefix, opts.DstPrefix, textURLMarkers)

	var bean interface{}
	switch tableName {
	case "question":
		bean = &entity.Question{OriginalText: origText, ParsedText: parsedText}
	case "answer":
		bean = &entity.Answer{OriginalText: origText, ParsedText: parsedText}
	default:
		return fmt.Errorf("unknown post table %q", tableName)
	}

	if _, err := x.ID(row.ID).Cols("original_text", "parsed_text").NoAutoTime().Update(bean); err != nil {
		return fmt.Errorf("update %s (id=%s) failed: %w", tableName, row.ID, err)
	}
	return nil
}

func fixRevisionByObjectType(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState, objectType int, label string) (fixResult, error) {
	result := fixResult{}
	err := runBatchedFix(state,
		func(offset int) ([]*entity.Revision, error) {
			return queryRevisionsByObjectType(x, objectType, label, offset)
		},
		func(revision *entity.Revision) error {
			return applyTextFix(&result, opts, state, label, fmt.Sprintf("revision id=%s", revision.ID),
				[]string{revision.Content}, jsonURLMarkers,
				func() error { return updateRevisionContent(x, revision, opts, label) })
		})
	if err != nil {
		return fixResult{}, err
	}
	return result, nil
}

func queryRevisionsByObjectType(x *xorm.Engine, objectType int, label string, offset int) ([]*entity.Revision, error) {
	revisions := make([]*entity.Revision, 0)
	err := x.Select("id, content").
		Where("object_type = ?", objectType).
		OrderBy("id").
		Limit(fixBatchSize, offset).
		Find(&revisions)
	if err != nil {
		return nil, fmt.Errorf("query %s revision failed: %w", label, err)
	}
	return revisions, nil
}

func updateRevisionContent(x *xorm.Engine, revision *entity.Revision, opts *FixURLPrefixOptions, label string) error {
	updated := &entity.Revision{
		Content: replaceURLPrefix(revision.Content, opts.SrcPrefix, opts.DstPrefix, jsonURLMarkers),
	}
	_, err := x.ID(revision.ID).Cols("content").NoAutoTime().Update(updated)
	if err != nil {
		return fmt.Errorf("update %s revision (id=%s) failed: %w", label, revision.ID, err)
	}
	return nil
}

func applyTextFix(result *fixResult, opts *FixURLPrefixOptions, state *fixRunState,
	label, target string, texts []string, markers []string, update func() error,
) error {
	switch classifyText(texts, opts.SrcPrefix, opts.DstPrefix, markers) {
	case matchBoth:
		fmt.Printf("[%s] WARN %s: has BOTH prefixes, skipping (manual fix needed)\n", label, target)
		result.Skipped++
		return nil
	case matchSkip:
		fmt.Printf("[%s] skip %s: already has DST_PREFIX\n", label, target)
		result.Skipped++
		return nil
	case matchNone:
		return nil
	}

	state.markAffected(result)
	fmt.Printf("[%s] %s: replacing %q with %q\n", label, target, opts.SrcPrefix, opts.DstPrefix)

	if opts.DryRun {
		return nil
	}
	if err := update(); err != nil {
		return err
	}
	result.Fixed++
	return nil
}
