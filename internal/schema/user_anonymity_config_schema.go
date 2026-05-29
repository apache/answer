package schema

import "github.com/apache/answer/internal/entity"

type UserAnonymityConfig struct {
	Enabled bool `json:"enabled"`
}

func NewUserAnonymityConfig(uc entity.UserAnonymityConfig) UserAnonymityConfig {
	return UserAnonymityConfig{
		Enabled: uc.Enabled,
	}
}

type UpdateUserAnonymityConfigReq struct {
	UserAnonymityConfig
	UserID string `json:"-"`
}

type GetUserAnonymityConfigResp struct {
	UserAnonymityConfig
}
