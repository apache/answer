package fake_username

import (
	"context"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/base/reason"
	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/fake_username"
	"github.com/segmentfault/pacman/errors"
)

type fakeUsernameRepo struct {
	data *data.Data
}

func NewFakeUsernameRepo(data *data.Data) fake_username.FakeUsernameRepo {
	return &fakeUsernameRepo{
		data: data,
	}
}

func (ur *fakeUsernameRepo) Add(ctx context.Context, userID, questionID string, fakeName string) (err error) {
	fakeUsername := &entity.FakeUsername{
		UserID:     userID,
		QuestionID: questionID,
		FakeName:   fakeName,
	}
	_, err = ur.data.DB.Context(ctx).Insert(fakeUsername)
	if err != nil {
		return errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return nil
}

func (ur *fakeUsernameRepo) GetByUserIDAndQuestionID(ctx context.Context, userID, questionID string) (
	fu *entity.FakeUsername, exist bool, err error,
) {
	fu = &entity.FakeUsername{UserID: userID, QuestionID: questionID}
	exist, err = ur.data.DB.Context(ctx).Get(fu)
	if err != nil {
		err = errors.InternalServer(reason.DatabaseError).WithError(err).WithStack()
	}
	return
}
