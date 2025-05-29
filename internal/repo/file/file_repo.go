package file

import (
	"context"
	"database/sql"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/errors"
)

type FileRepo interface {
	Save(ctx context.Context, file *entity.File) error
	GetByID(ctx context.Context, id string) (*entity.File, error)
}

type fileRepo struct {
	data *data.Data
}

func NewFileRepo(data *data.Data) FileRepo {
	return &fileRepo{data: data}
}

func (r *fileRepo) Save(ctx context.Context, file *entity.File) error {
	_, err := r.data.DB.Context(ctx).Insert(file)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (r *fileRepo) GetByID(ctx context.Context, id string) (*entity.File, error) {
	var blob entity.File
	ok, err := r.data.DB.Context(ctx).ID(id).Get(&blob)
	if err != nil {
		return nil, errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	if !ok {
		return nil, sql.ErrNoRows
	}
	return &blob, nil
}
