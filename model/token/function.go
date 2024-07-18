package model_token

import (
	"certification/constant"

	"gorm.io/gorm"
)

// Update the token status to used
func UpdateTokenStatus(token string, tx *gorm.DB) error {
	err := tx.Model(&Token{}).Where("token = ?", token).Update("status", constant.USED).Error
	if err != nil {
		return err
	}

	return nil
}
