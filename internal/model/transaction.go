package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID            string `gorm:"type:uuid;primary_key"`
	WalletID      string `gorm:"type:uuid;not null"`
	Wallet        Wallet
	OperationType string `gorm:"type:varchar(10);not null"`
	Amount        int64  `gorm:"not null"`
	CreatedAt     time.Time
}

func (w *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	w.ID = uuid.New().String()
	return
}
