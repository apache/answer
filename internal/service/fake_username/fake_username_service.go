package fake_username

import (
	"context"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/service/user_anonymity_config"
	"github.com/segmentfault/pacman/log"
)

type FakeUsernameRepo interface {
	Add(ctx context.Context, userID, questionID string, fakeName string) (err error)
	GetByUserIDAndQuestionID(ctx context.Context, userID, questionID string) (fu *entity.FakeUsername, exist bool, err error)
	BatchGetByUserIDs(ctx context.Context, userIDs []string, questionID string) ([]*entity.FakeUsername, error)
}

type FakeUsernameService struct {
	fakeUsernameGenerator   *FakeUsernameGenerator
	fakeUsernameRepo        FakeUsernameRepo
	userAnonymityConfigRepo user_anonymity_config.UserAnonymityConfigRepo
}

func NewFakeUsernameService(
	fakeUsernameGenerator *FakeUsernameGenerator,
	fakeUsernameRepo FakeUsernameRepo,
	userAnonymityConfigRepo user_anonymity_config.UserAnonymityConfigRepo,
) *FakeUsernameService {
	return &FakeUsernameService{
		fakeUsernameGenerator:   fakeUsernameGenerator,
		fakeUsernameRepo:        fakeUsernameRepo,
		userAnonymityConfigRepo: userAnonymityConfigRepo,
	}
}

// func (fs *FakeUsernameService) AddFakeUsernameFor(ctx context.Context, userID, questionID string) (err error) {
// 	return fs.fakeUsernameRepo.Add(ctx, userID, questionID, fs.fakeUsernameGenerator.GenerateFakeName())
// }

func (fs *FakeUsernameService) AddFakeUsernameIfNeeded(ctx context.Context, userID, questionID string) (err error) {
	userAnonymityConfig, _, err := fs.userAnonymityConfigRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to get user anonymity config: %w", err)
	}

	if userAnonymityConfig.Enabled {
		return fs.fakeUsernameRepo.Add(ctx, userID, questionID, fs.fakeUsernameGenerator.GenerateFakeName())
	}

	return
}
