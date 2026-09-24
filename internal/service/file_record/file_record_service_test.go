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

package file_record

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/revision"
	"github.com/apache/answer/internal/service/service_config"
)

type testFileRecordRepo struct {
	records []*entity.FileRecord
}

func (r *testFileRecordRepo) AddFileRecord(context.Context, *entity.FileRecord) error {
	return nil
}

func (r *testFileRecordRepo) UpdateFileRecord(_ context.Context, fileRecord *entity.FileRecord) error {
	for _, record := range r.records {
		if record.ID == fileRecord.ID {
			record.ObjectID = fileRecord.ObjectID
		}
	}
	return nil
}

func (r *testFileRecordRepo) GetFileRecordPage(_ context.Context, page, pageSize int, _ *entity.FileRecord) ([]*entity.FileRecord, int64, error) {
	available := make([]*entity.FileRecord, 0, len(r.records))
	for _, record := range r.records {
		if record.Status == entity.FileRecordStatusAvailable {
			available = append(available, record)
		}
	}
	start := (page - 1) * pageSize
	if start >= len(available) {
		return []*entity.FileRecord{}, int64(len(available)), nil
	}
	end := start + pageSize
	if end > len(available) {
		end = len(available)
	}
	return available[start:end], int64(len(available)), nil
}

func (r *testFileRecordRepo) DeleteFileRecord(_ context.Context, id int) error {
	for _, record := range r.records {
		if record.ID == id {
			record.Status = entity.FileRecordStatusDeleted
		}
	}
	return nil
}

func (r *testFileRecordRepo) GetFileRecordByURL(context.Context, string) (*entity.FileRecord, error) {
	return nil, nil
}

type testRevisionRepo struct {
	revision.RevisionRepo
}

func (testRevisionRepo) GetLastRevisionByObjectID(context.Context, string) (*entity.Revision, bool, error) {
	return nil, false, nil
}

func (testRevisionRepo) GetLastRevisionByFileURL(context.Context, string) (*entity.Revision, bool, error) {
	return nil, false, nil
}

func TestCleanOrphanUploadFilesDoesNotSkipRecordsWhenStatusChanges(t *testing.T) {
	const recordCount = 1001
	uploadPath := t.TempDir()
	if err := os.MkdirAll(filepath.Join(uploadPath, "tmp"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(uploadPath, constant.DeletedSubPath), 0o755); err != nil {
		t.Fatal(err)
	}

	records := make([]*entity.FileRecord, 0, recordCount)
	createdAt := time.Now().AddDate(0, 0, -3)
	for id := 1; id <= recordCount; id++ {
		filePath := filepath.Join("tmp", "orphan-"+strconv.Itoa(id))
		if err := os.WriteFile(filepath.Join(uploadPath, filePath), []byte("orphan"), 0o600); err != nil {
			t.Fatal(err)
		}
		records = append(records, &entity.FileRecord{
			ID:        id,
			CreatedAt: createdAt,
			FilePath:  filePath,
			FileURL:   "/" + filePath,
			ObjectID:  "1",
			Status:    entity.FileRecordStatusAvailable,
		})
	}

	repo := &testFileRecordRepo{records: records}
	service := NewFileRecordService(
		repo,
		testRevisionRepo{},
		&service_config.ServiceConfig{UploadPath: uploadPath},
		nil,
		nil,
	)
	service.CleanOrphanUploadFiles(context.Background())

	for _, record := range records {
		if record.Status != entity.FileRecordStatusDeleted {
			t.Fatalf("orphan file record %d was not deleted", record.ID)
		}
	}
}
