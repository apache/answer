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

	"github.com/apache/answer/internal/base/pager"
	"github.com/apache/answer/internal/service/file_record"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/errors"
)

// fileRecordRepo fileRecord repository
type fileRecordRepo struct {
	data *data.Data
}

// NewFileRecordRepo new repository
func NewFileRecordRepo(data *data.Data) file_record.FileRecordRepo {
	return &fileRecordRepo{
		data: data,
	}
}

// AddFileRecord add file record
func (fr *fileRecordRepo) AddFileRecord(ctx context.Context, fileRecord *entity.FileRecord) (err error) {
	_, err = fr.data.DB.Context(ctx).Insert(fileRecord)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetFileRecordPage get fileRecord page
func (fr *fileRecordRepo) GetFileRecordPage(ctx context.Context, page, pageSize int, cond *entity.FileRecord) (
	fileRecordList []*entity.FileRecord, total int64, err error) {
	fileRecordList = make([]*entity.FileRecord, 0)

	session := fr.data.DB.Context(ctx)
	total, err = pager.Help(page, pageSize, &fileRecordList, cond, session)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// DeleteFileRecord delete file record
func (fr *fileRecordRepo) DeleteFileRecord(ctx context.Context, id int) (err error) {
	_, err = fr.data.DB.Context(ctx).ID(id).Cols("status").Update(&entity.FileRecord{Status: entity.FileRecordStatusDeleted})
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// UpdateFileRecord update file record
func (fr *fileRecordRepo) UpdateFileRecord(ctx context.Context, fileRecord *entity.FileRecord) (err error) {
	_, err = fr.data.DB.Context(ctx).ID(fileRecord.ID).Update(fileRecord)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}

// GetFileRecordByURL gets a file record by its url
func (fr *fileRecordRepo) GetFileRecordByURL(ctx context.Context, fileURL string) (record *entity.FileRecord, err error) {
	record = &entity.FileRecord{}
	session := fr.data.DB.Context(ctx)
	exists, err := session.Where("file_url = ? AND status = ?", fileURL, entity.FileRecordStatusAvailable).Get(record)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
		return
	}
	if !exists {
		return
	}
	return record, nil
}
