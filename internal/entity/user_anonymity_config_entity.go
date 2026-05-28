package entity

import "time"

type UserAnonymityConfig struct {
	ID        string    `xorm:"not null pk autoincr BIGINT(20) id"`
	CreatedAt time.Time `xorm:"created TIMESTAMP created_at"`
	UpdatedAt time.Time `xorm:"updated TIMESTAMP updated_at"`
	UserID    string    `xorm:"not null default 0 BIGINT(20) INDEX UNIQUE user_id"`
	Enabled   bool      `xorm:"not null default false BOOL enabled"`
}

func (UserAnonymityConfig) TableName() string {
	return "user_anonymity_config"
}
