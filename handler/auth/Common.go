package handler_auth

import (
	"certification/constant"
	"certification/logger"
	model_account "certification/model/account"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Access struct {
	ModuleID     string `json:"module_id"`
	ModuleAccess bool   `json:"module_access"`
	ReadAccess   bool   `json:"read_access"`
	WriteAccess  bool   `json:"write_access"`
	DeleteAccess bool   `json:"delete_access"`
}

type IncomingLogin struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func CheckLogin(account *model_account.Account, body IncomingLogin, role constant.AccountRoleType, db *gorm.DB, ctx *fiber.Ctx) error {
	if account.Role != role {
		logger.Log.Error("Unauthorized account for email: ", account.Email)
		return errors.New("Unauthorized account")
	}
	if account.Status != "active" {
		logger.Log.Error("Inactive account for email: ", account.Email)
		return errors.New("Inactive account")
	}

	if !CheckPasswordHash(body.Password, account.Password) {
		logger.Log.Error("Invalid password by: ", body.Email)
		return errors.New("Invalid password")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)

	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
