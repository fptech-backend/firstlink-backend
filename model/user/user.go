package model_user

import "github.com/google/uuid"

type User struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;unique;default: gen_random_uuid();"`

	AccountID *uuid.UUID `json:"account_id" gorm:"type:uuid"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
}
