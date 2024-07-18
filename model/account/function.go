package model_account

import (
	"certification/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
