package model

import (
	"certification/constant"
	"certification/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Account struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;unique;default: gen_random_uuid();"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Email    string                   `json:"email"`
	Password string                   `json:"password"`
	Role     constant.AccountRoleType `json:"role"`
	Status   constant.Status          `json:"status"`

	Company Company `json:"company" gorm:"foreignKey:AccountID"`
	User    User    `json:"user" gorm:"foreignKey:AccountID"`
}

type ResponseProfile struct {
	ID     uuid.UUID                `json:"id"`
	Email  string                   `json:"email"`
	Role   constant.AccountRoleType `json:"role"`
	Status constant.Status          `json:"status"`

	Company Company `json:"company" gorm:"foreignKey:AccountID"`
	User    User    `json:"user" gorm:"foreignKey:AccountID"`
}

// ----------------- Account Functions -----------------

// get account by id
func GetAccountByID(db *gorm.DB, id uuid.UUID) (*Account, error) {
	var a Account
	if err := db.Where("id = ?", id).First(&a).Error; err != nil {
		return nil, err
	}

	return &a, nil
}

// get account by email
func GetAccountByEmail(db *gorm.DB, email string) (*Account, error) {
	var a Account
	if err := db.Where("email = ?", email).First(&a).Error; err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	return &a, nil
}

// Get Profile by Account ID without password
func GetProfileByAccountID(db *gorm.DB, accountID uuid.UUID) (*ResponseProfile, error) {
	var a Account
	if err := db.Preload("Company").Preload("User").Where("id = ?", accountID).First(&a).Error; err != nil {
		logger.Log.Error(err)
		return nil, err
	}

	return &ResponseProfile{
		ID:     a.ID,
		Email:  a.Email,
		Role:   a.Role,
		Status: a.Status,

		Company: a.Company,
		User:    a.User,
	}, nil
}
