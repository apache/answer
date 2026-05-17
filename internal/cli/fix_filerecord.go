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

	"github.com/apache/answer/internal/entity"
	"xorm.io/xorm"
)

func fixFileRecordBySources(x *xorm.Engine, opts *FixURLPrefixOptions, state *fixRunState, sources []string, label string) (fixResult, error) {
	result := fixResult{}

	err := runBatchedFix(state, func(offset int) ([]*entity.FileRecord, error) {
		records := make([]*entity.FileRecord, 0)
		err := x.Select("id, file_url").
			In("source", sources).
			OrderBy("id").
			Limit(fixBatchSize, offset).
			Find(&records)
		if err != nil {
			return nil, fmt.Errorf("query %s file_record failed: %w", label, err)
		}
		return records, nil
	}, func(record *entity.FileRecord) error {
		return applyValueFix(&result, opts, state, label, fmt.Sprintf("file_record id=%d", record.ID), record.FileURL, func(newURL string) error {
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
