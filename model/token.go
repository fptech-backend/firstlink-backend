package model

import (
	"certification/constant"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;unique;default: gen_random_uuid();"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	AccountID uuid.UUID       `json:"account_id" gorm:"type:uuid"`
	Token     string          `json:"token"`
	ExpireAt  time.Time       `json:"expire_at"`
	Type      string          `json:"type"`
	Status    constant.Status `json:"status"`
}
