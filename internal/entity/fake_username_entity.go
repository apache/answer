package entity

import "time"

// FakeUsername fake username
type FakeUsername struct {
	ID         string    `xorm:"pk autoincr BIGINT(20) id"`
	CreatedAt  time.Time `xorm:"created TIMESTAMP created_at"`
	UpdatedAt  time.Time `xorm:"updated TIMESTAMP updated_at"`
	UserID     string    `xorm:"not null default 0 BIGINT(20) user_id"`
	QuestionID string    `xorm:"not null default 0 BIGINT(20) question_id"`
	FakeName   string    `xorm:"not null VARCHAR(50) fake_name"`
}

// TableName fake username table name
func (FakeUsername) TableName() string {
	return "fake_username"
}
