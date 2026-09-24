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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/constant"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/service_config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// fakeFileRecordRepo is an in-memory file record repo. It keeps the same semantics as the
// database one: GetFileRecordListAfterID only returns available records ordered by id, and
// DeleteFileRecord marks a record as deleted so that it leaves the available result set.
type fakeFileRecordRepo struct {
	records   []*entity.FileRecord
	listCalls int
}

func (f *fakeFileRecordRepo) AddFileRecord(_ context.Context, fileRecord *entity.FileRecord) error {
	f.records = append(f.records, fileRecord)
	return nil
}

func (f *fakeFileRecordRepo) UpdateFileRecord(_ context.Context, fileRecord *entity.FileRecord) error {
	return nil
}

func (f *fakeFileRecordRepo) GetFileRecordListAfterID(_ context.Context, lastID, limit int) (
	[]*entity.FileRecord, error) {
	f.listCalls++
	list := make([]*entity.FileRecord, 0)
	for _, record := range f.records {
		if record.Status != entity.FileRecordStatusAvailable || record.ID <= lastID {
			continue
		}
		list = append(list, record)
		if len(list) == limit {
			break
		}
	}
	return list, nil
}

func (f *fakeFileRecordRepo) DeleteFileRecord(_ context.Context, id int) error {
	for _, record := range f.records {
		if record.ID == id {
			record.Status = entity.FileRecordStatusDeleted
			return nil
		}
	}
	return fmt.Errorf("file record %d not found", id)
}

func (f *fakeFileRecordRepo) GetFileRecordByURL(_ context.Context, fileURL string) (
	*entity.FileRecord, error) {
	for _, record := range f.records {
		if record.FileURL == fileURL && record.Status == entity.FileRecordStatusAvailable {
			return record, nil
		}
	}
	return nil, nil
}

// fakeRevisionRepo is a revision repo stub which reports the given file urls as referenced.
type fakeRevisionRepo struct {
	usedFileURLs map[string]*entity.Revision
}

func (f *fakeRevisionRepo) AddRevision(_ context.Context, _ *entity.Revision, _ bool) error {
	return nil
}

func (f *fakeRevisionRepo) GetRevisionByID(_ context.Context, _ string) (*entity.Revision, bool, error) {
	return nil, false, nil
}

func (f *fakeRevisionRepo) GetLastRevisionByObjectID(_ context.Context, _ string) (*entity.Revision, bool, error) {
	return nil, false, nil
}

func (f *fakeRevisionRepo) GetLastRevisionByFileURL(_ context.Context, fileURL string) (*entity.Revision, bool, error) {
	if revision, ok := f.usedFileURLs[fileURL]; ok {
		return revision, true, nil
	}
	return nil, false, nil
}

func (f *fakeRevisionRepo) GetRevisionList(_ context.Context, _ *entity.Revision) ([]entity.Revision, error) {
	return nil, nil
}

func (f *fakeRevisionRepo) UpdateObjectRevisionId(_ context.Context, _ *entity.Revision, _ *xorm.Session) error {
	return nil
}

func (f *fakeRevisionRepo) ExistUnreviewedByObjectID(_ context.Context, _ string) (*entity.Revision, bool, error) {
	return nil, false, nil
}

func (f *fakeRevisionRepo) GetUnreviewedRevisionPage(_ context.Context, _, _ int, _ []int) (
	[]*entity.Revision, int64, error) {
	return nil, 0, nil
}

func (f *fakeRevisionRepo) CountUnreviewedRevision(_ context.Context, _ []int) (int64, error) {
	return 0, nil
}

func (f *fakeRevisionRepo) UpdateStatus(_ context.Context, _ string, _ int, _ string) error {
	return nil
}

// TestFileRecordService_CleanOrphanUploadFiles checks that all orphan files are cleaned even
// though the scan marks records as deleted while iterating, which used to make offset
// pagination skip records between pages.
func TestFileRecordService_CleanOrphanUploadFiles(t *testing.T) {
	const (
		recordCount        = 2500
		referencedRecordID = 1234
		recentRecordID     = 2500
	)

	uploadPath := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(uploadPath, constant.DeletedSubPath), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(uploadPath, "files", "post"), 0o755))

	repo := &fakeFileRecordRepo{}
	revisionRepo := &fakeRevisionRepo{usedFileURLs: map[string]*entity.Revision{}}
	oldCreatedAt := time.Now().Add(-time.Hour * 72)
	for id := 1; id <= recordCount; id++ {
		filePath := fmt.Sprintf("files/post/%d.txt", id)
		fileURL := fmt.Sprintf("uploads/%s", filePath)
		require.NoError(t, os.WriteFile(filepath.Join(uploadPath, filePath), []byte("x"), 0o644))

		createdAt := oldCreatedAt
		if id == recentRecordID {
			createdAt = time.Now()
		}
		repo.records = append(repo.records, &entity.FileRecord{
			ID:        id,
			CreatedAt: createdAt,
			FilePath:  filePath,
			FileURL:   fileURL,
			ObjectID:  "0",
			Status:    entity.FileRecordStatusAvailable,
		})
		if id == referencedRecordID {
			revisionRepo.usedFileURLs[fileURL] = &entity.Revision{ObjectID: "100"}
		}
	}

	serviceConfig := &service_config.ServiceConfig{UploadPath: uploadPath}
	service := NewFileRecordService(repo, revisionRepo, serviceConfig, nil, nil)

	service.CleanOrphanUploadFiles(context.Background())

	assert.GreaterOrEqual(t, repo.listCalls, 3, "all records should be scanned page by page")

	deletedCount := 0
	for _, record := range repo.records {
		filePath := filepath.Join(uploadPath, record.FilePath)
		deletedFilePath := filepath.Join(uploadPath, constant.DeletedSubPath, filepath.Base(record.FilePath))
		switch record.ID {
		case referencedRecordID, recentRecordID:
			assert.Equal(t, entity.FileRecordStatusAvailable, record.Status,
				"record %d should not be deleted", record.ID)
			assert.FileExists(t, filePath, "file of record %d should stay in place", record.ID)
			assert.NoFileExists(t, deletedFilePath, "file of record %d should not be moved", record.ID)
		default:
			deletedCount++
			assert.Equal(t, entity.FileRecordStatusDeleted, record.Status,
				"record %d should be deleted", record.ID)
			assert.NoFileExists(t, filePath, "file of record %d should be moved away", record.ID)
			assert.FileExists(t, deletedFilePath, "file of record %d should be moved to deleted dir", record.ID)
		}
	}
	assert.Equal(t, recordCount-2, deletedCount)
}
