package model_account

import (
	"certification/constant"
	model_company "certification/model/company"
	model_user "certification/model/user"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;unique;default: gen_random_uuid();"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Email    string                   `json:"email"`
	Password string                   `json:"password"`
	Role     constant.AccountRoleType `json:"role"`
	Status   constant.Status          `json:"status"`

	Company model_company.Company `json:"company" gorm:"foreignKey:AccountID"`
	User    model_user.User       `json:"user" gorm:"foreignKey:AccountID"`
}

type ResponseProfile struct {
	ID     uuid.UUID                `json:"id"`
	Email  string                   `json:"email"`
	Role   constant.AccountRoleType `json:"role"`
	Status constant.Status          `json:"status"`

	Company model_company.Company `json:"company" gorm:"foreignKey:AccountID"`
	User    model_user.User       `json:"user" gorm:"foreignKey:AccountID"`
}
