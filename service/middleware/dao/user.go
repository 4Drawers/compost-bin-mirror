package dao

import (
	"compost-bin/service/middleware/identity"
	"time"

	"gorm.io/gorm"
)

type User struct {
	Id        int64  `gorm:"primarykey;autoIncrement:false"`
	Username  string `gorm:"unique;not null;size:20"`
	Email     string `gorm:"unique;size:50;default:null"`
	Avatar    string `gorm:"size:200"`
	Sign      string `gorm:"not null;size:200;default:'这位画师比较高冷，还没有留下签名(′゜ω。‵)。。。'"`
	Password  string `gorm:"not null;size:200"`
	Salt      string `gorm:"not null;size:100"`
	Character string `gorm:"not null;size:10;default:'sensitive'"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.Id = identity.GenerateId()
	return nil
}
