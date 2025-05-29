package entity

import (
	"time"
)

type File struct {
	ID        string    `xorm:"pk varchar(36)"`
	FileName  string    `xorm:"varchar(255) not null"`
	MimeType  string    `xorm:"varchar(100)"`
	Size      int64     `xorm:"bigint"`
	Content   []byte    `xorm:"blob"`
	CreatedAt time.Time `xorm:"created"`
}

func (File) TableName() string {
	return "file"
}
