package handler_auth

import (
	"certification/constant"
	"certification/database"
	"certification/logger"
	"certification/mailer"
	model_account "certification/model/account"
	model_token "certification/model/token"
	model_user "certification/model/user"
	"certification/response"
	"certification/template"
	"certification/utils"
	"context"
	"mime/multipart"
	"time"

	"github.com/Boostport/mjml-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type IncomingSignUpUser struct {
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"first_name" validate:"-"`
	LastName  string `json:"last_name" validate:"-"`
}

type ResponseSignUp struct {
	AccountID uuid.UUID `json:"account_id"`
}

// @Summary Sign Up User
// @Tags Auth
// @Accept json
// @Produce json
// @Param IncomingSignUpUser body IncomingSignUpUser true "Sign up data"
// @Success 200 {object} response.DataResponse{data=ResponseSignUp} "Successful sign up"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/signup/user [post]
func SignUpUser(ctx *fiber.Ctx, initializer *database.Initializer) error {
	var body IncomingSignUpUser
	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	// Check if email already exists
	_, err := model_account.GetAccountByEmail(initializer.DB, body.Email)
	if err == nil {
		logger.Log.Error("Email already exists: ", body.Email)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Email already exists"))
	}

	accountID := uuid.New()

	hashPassword, err := HashPassword(body.Password)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Unable to hash password"))
	}

	tokenStr, err := utils.GenerateToken()
	if err != nil {
		logger.Log.Error("Error in generating token for ", accountID)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody("Error in generating token"))
	}

	account := model_account.Account{
		ID:       accountID,
		Email:    body.Email,
		Password: hashPassword,
		Role:     constant.ROLE_USER,
		Status:   constant.PENDING,
	}

	user := model_user.User{
		AccountID: &accountID,
		FirstName: body.FirstName,
		LastName:  body.LastName,
	}

	token := model_token.Token{
		AccountID: accountID,
		Token:     tokenStr,
		ExpireAt:  time.Now().Add(time.Hour * 24 * 7), // 7 days expiry
		Type:      constant.VALIDATION_TOKEN,
		Status:    constant.PENDING,
	}

	tx := initializer.DB.Begin()

	err = tx.Create(&account).Error
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(err.Error()))
	}

	err = tx.Create(&token).Error
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(err.Error()))
	}

	err = tx.Create(&user).Error
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(err.Error()))
	}

	// Commit the changes so far
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(err.Error()))
	}

	// Send email with token
	html, err := mjml.ToHTML(context.Background(), template.TemplateEmailInvitation(initializer, body.FirstName+body.LastName, tokenStr), mjml.WithMinify(true))
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Unable to convert MJML to HTML"))
	}

	subject := "Welcome to CertFirst!"
	var emptyFile *multipart.FileHeader
	err = mailer.SendEmail(html, subject, []string{body.Email}, emptyFile)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(response.ErrorResponseBody(err.Error()))
	}

	responeSignUp := ResponseSignUp{
		AccountID: accountID,
	}

	logger.Log.Info(constant.SuccessSignUp)
	return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(responeSignUp, "Sign up successful"))
}
