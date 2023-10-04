package models

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	//ID        uuid.UUID `gorm:"primarykey;type:uuid;default:uuid_generate_v4()"`
	ID        uint `gorm:"primarykey,autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
