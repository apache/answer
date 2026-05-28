package user_anonymity_config

import (
	"context"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/schema"
)

type UserAnonymityConfigRepo interface {
	Add(ctx context.Context, userIDs []string, enabled bool) (err error)
	Save(ctx context.Context, uc *entity.UserAnonymityConfig) (err error)
	GetByUserID(ctx context.Context, userID string) (uc *entity.UserAnonymityConfig, exists bool, err error)
}

type UserAnonymityConfigService struct {
	userAnonymityConfigRepo UserAnonymityConfigRepo
}

func NewUserAnonymityConfigService(userAnonymityConfigRepo UserAnonymityConfigRepo) *UserAnonymityConfigService {
	return &UserAnonymityConfigService{
		userAnonymityConfigRepo: userAnonymityConfigRepo,
	}
}

func (us *UserAnonymityConfigService) GetUserAnonymityConfig(ctx context.Context, userID string) (
	resp *schema.GetUserAnonymityConfigResp, err error,
) {
	anonymityConfig, exists, err := us.userAnonymityConfigRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		anonymityConfig = &entity.UserAnonymityConfig{}
	}
	resp = &schema.GetUserAnonymityConfigResp{}
	resp.UserAnonymityConfig = schema.NewUserAnonymityConfig(*anonymityConfig)
	return resp, nil
}

func (us *UserAnonymityConfigService) UpdateUserAnonymityConfig(
	ctx context.Context, req *schema.UpdateUserAnonymityConfigReq) (err error) {
	// req.Format()

	err = us.userAnonymityConfigRepo.Save(ctx, us.convertToEntity(ctx, req.UserID, req.Enabled))
	if err != nil {
		return err
	}
	return nil
}

func (us *UserAnonymityConfigService) SetDefaultUserAnonymityConfig(
	ctx context.Context, userIDs []string,
) (err error) {
	return us.userAnonymityConfigRepo.Add(ctx, userIDs, false)
}

func (us *UserAnonymityConfigService) convertToEntity(
	_ context.Context, userID string, enabled bool,
) *entity.UserAnonymityConfig {
	return &entity.UserAnonymityConfig{
		UserID:  userID,
		Enabled: enabled,
	}
}
