package model_company

import (
	"github.com/google/uuid"
)

type Company struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;unique;default: gen_random_uuid();"`

	AccountID *uuid.UUID `json:"account_id" gorm:"type:uuid"`
	Name      string     `json:"name"`
}
