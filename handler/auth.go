package handler

import (
	"certification/cache"
	"certification/constant"
	"certification/database"
	"certification/logger"
	"certification/mailer"
	"certification/model"
	"certification/response"
	"certification/template"
	"certification/utils"
	"context"
	"errors"
	"mime/multipart"
	"time"

	"github.com/Boostport/mjml-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type IncomingLogin struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type IncomingSignUpCompany struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name" validate:"-"`
}

type IncomingSignUpUser struct {
	Email     string `json:"email" validate:"required"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"first_name" validate:"-"`
	LastName  string `json:"last_name" validate:"-"`
}

type IncomingActivate struct {
	Token string `json:"token" validate:"required"`
}

type IncomingValidation struct {
	Token         string `json:"token" validate:"required"`
	CertificateID string `json:"certificate_id" validate:"required"`
}

type ResponseSignUp struct {
	AccountID uuid.UUID `json:"account_id"`
}

type ResponseNewSignUp struct {
	AccountID      uuid.UUID `json:"account_id"`
	CertifiicateID string    `json:"certificate_id"`
}

type ResponseActivate struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type Access struct {
	ModuleID     string `json:"module_id"`
	ModuleAccess bool   `json:"module_access"`
	ReadAccess   bool   `json:"read_access"`
	WriteAccess  bool   `json:"write_access"`
	DeleteAccess bool   `json:"delete_access"`
}

type IncomingForgotPassword struct {
	Email string `json:"email" validate:"required"`
}

type IncomingResetPassword struct {
	Password string `json:"password" validate:"required"`
	Token    string `json:"token" validate:"required"`
}

func UpdateAccountStatus(id uuid.UUID, status constant.Status, tx *gorm.DB) error {

	err := tx.Model(&model.Account{}).
		Where("id = ?", id).
		Update("status", status).
		Error

	if err != nil {
		return err
	}

	return nil
}

func CheckLogin(account *model.Account, body IncomingLogin, role constant.AccountRoleType, db *gorm.DB, ctx *fiber.Ctx) error {
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

// Login
// @Summary User login
// @Description Authenticate user and generate token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body IncomingLogin true "User credentials"
// @Success 200 {object} response.MessageDataResponse{data=response.LoginSuccessResponse} "Successful login"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 401 {object} response.MessageResponse "Unauthorized"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/login/user [post]
func LoginUser(ctx *fiber.Ctx, initializer *database.Initializer, db *gorm.DB) error {
	var body IncomingLogin
	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	account, err := model.GetAccountByEmail(initializer.DB, body.Email)
	if err != nil {
		logger.Log.Error(err.Error())
		return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(err.Error()))
	}

	err = CheckLogin(account, body, constant.ROLE_USER, db, ctx)
	if err != nil {
		logger.Log.Error(err.Error())
		return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(err.Error()))
	}

	expiry := time.Now().Add(time.Hour * 24 * 7) // 7 days expiry
	token, err := utils.GenerateJWT(&account.ID, &account.User.ID, &account.Email, &expiry, &account.Role)
	if err != nil {
		logger.Log.Error("Error in generating token for ", account.ID)
		return ctx.Status(fiber.StatusUnauthorized).JSON(response.LoginFailResponseBody())
	}

	modulesArr := make([]map[string]interface{}, 0)

	// var access []model.Permission
	// db.Where("role_id = ? AND department_id = ?", account.RoleID, account.DeptID).
	// 	Find(&access)

	// for i := 0; i < len(access); i++ {
	// 	moduleMap := map[string]interface{}{
	// 		"module_id":     strconv.FormatUint(uint64(access[i].ModuleID), 10),
	// 		"module_access": access[i].ModuleAccess,
	// 		"read_access":   access[i].ReadAccess,
	// 		"write_access":  access[i].WriteAccess,
	// 		"delete_access": access[i].DeleteAccess,
	// 	}
	// 	modulesArr = append(modulesArr, moduleMap)
	// }

	data := utils.RedisValue{
		Token:  token,
		Module: modulesArr,
		Status: constant.CREATED,
	}

	jsonData, err := jsoniter.Marshal(data)
	if err != nil {
		logger.Log.Error("Error in encoding JSON: ", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.LoginFailResponseBody())
	}

	// Store in Redis
	err = cache.Redis.RDB.Set(context.Background(), account.ID.String(), jsonData, 0).Err()
	if err != nil {
		logger.Log.Errorf("unable to set ID %s and JSON %s to Redis", account.ID, data)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.LoginFailResponseBody())
	}

	logger.Log.Info(constant.SuccessLogIn, account.ID)
	return ctx.Status(fiber.StatusOK).JSON(response.LoginSuccessResponseBody(
		account.ID,
		token,
		account.Email,
	))
}

// Login
// @Summary Company login
// @Description Authenticate company and generate token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body IncomingLogin true "Company credentials"
// @Success 200 {object} response.MessageDataResponse{data=response.LoginSuccessResponse} "Successful login"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 401 {object} response.MessageResponse "Unauthorized"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/login/company [post]
func LoginCompany(ctx *fiber.Ctx, initializer *database.Initializer, db *gorm.DB) error {
	var body IncomingLogin
	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	account, err := model.GetAccountByEmail(initializer.DB, body.Email)
	if err != nil {
		logger.Log.Error(err.Error())
		return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(err.Error()))
	}

	err = CheckLogin(account, body, constant.ROLE_COMPANY, db, ctx)
	if err != nil {
		logger.Log.Error(err.Error())
		return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(err.Error()))
	}

	expiry := time.Now().Add(time.Hour * 24 * 7) // 7 days expiry
	token, err := utils.GenerateJWT(&account.ID, &account.Company.ID, &account.Email, &expiry, &account.Role)
	if err != nil {
		logger.Log.Error("Error in generating token for ", account.ID)
		return ctx.Status(fiber.StatusUnauthorized).JSON(response.LoginFailResponseBody())
	}

	modulesArr := make([]map[string]interface{}, 0)

	// var access []model.Permission
	// db.Where("role_id = ? AND department_id = ?", account.RoleID, account.DeptID).
	// 	Find(&access)

	// for i := 0; i < len(access); i++ {
	// 	moduleMap := map[string]interface{}{
	// 		"module_id":     strconv.FormatUint(uint64(access[i].ModuleID), 10),
	// 		"module_access": access[i].ModuleAccess,
	// 		"read_access":   access[i].ReadAccess,
	// 		"write_access":  access[i].WriteAccess,
	// 		"delete_access": access[i].DeleteAccess,
	// 	}
	// 	modulesArr = append(modulesArr, moduleMap)
	// }

	data := utils.RedisValue{
		Token:  token,
		Module: modulesArr,
		Status: constant.CREATED,
	}

	jsonData, err := jsoniter.Marshal(data)
	if err != nil {
		logger.Log.Error("Error in encoding JSON: ", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.LoginFailResponseBody())
	}

	// Store in Redis
	err = cache.Redis.RDB.Set(context.Background(), account.ID.String(), jsonData, 0).Err()
	if err != nil {
		logger.Log.Errorf("unable to set ID %s and JSON %s to Redis", account.ID, data)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.LoginFailResponseBody())
	}

	logger.Log.Info(constant.SuccessLogIn, account.ID)
	return ctx.Status(fiber.StatusOK).JSON(response.LoginSuccessResponseBody(
		account.ID,
		token,
		account.Email,
	))
}

// Logout
// @Summary User logout
// @Description Authenticate user and delete token
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.MessageResponse "Successful login"
// @Failure 401 {object} response.MessageResponse "Unauthorized"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/logout [post]
func Logout(ctx *fiber.Ctx, initializer *database.Initializer) error {
	id, ok := ctx.Locals("id").(uuid.UUID)
	if !ok {
		errMsg := "Failed to extract account ID from Locals"
		logger.Log.Error(errMsg)
		return ctx.Status(fiber.StatusUnauthorized).JSON(response.LogoutFailResponseBody(errMsg))
	}

	// Delete the value in Redis
	err := cache.Redis.RDB.Del(context.Background(), id.String()).Err()
	if err != nil {
		logger.Log.Errorf("unable to delete value for %s in Redis", id)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.LogoutFailResponseBody(err.Error()))
	}

	logger.Log.Info(constant.SuccessLogOut, id)
	return ctx.Status(fiber.StatusOK).JSON(response.LogoutSuccessResponseBody())
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
	_, err := model.GetAccountByEmail(initializer.DB, body.Email)
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

	account := model.Account{
		ID:       accountID,
		Email:    body.Email,
		Password: hashPassword,
		Role:     constant.ROLE_USER,
		Status:   constant.PENDING,
	}

	user := model.User{
		AccountID: &accountID,
		FirstName: body.FirstName,
		LastName:  body.LastName,
	}

	token := model.Token{
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

// @Summary Sign Up Company
// @Tags Auth
// @Accept json
// @Produce json
// @Param IncomingSignUpCompany body IncomingSignUpCompany true "Sign up data"
// @Success 200 {object} response.DataResponse{data=ResponseSignUp} "Successful sign up"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/signup/company [post]
func SignUpCompany(ctx *fiber.Ctx, initializer *database.Initializer) error {
	var body IncomingSignUpCompany
	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	// Check if email already exists
	_, err := model.GetAccountByEmail(initializer.DB, body.Email)
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

	account := model.Account{
		ID:       accountID,
		Email:    body.Email,
		Password: hashPassword,
		Role:     constant.ROLE_COMPANY,
		Status:   constant.PENDING,
	}

	token := model.Token{
		AccountID: accountID,
		Token:     tokenStr,
		ExpireAt:  time.Now().Add(time.Hour * 24 * 7), // 7 days expiry
		Type:      constant.VALIDATION_TOKEN,
		Status:    constant.PENDING,
	}

	admin := model.Company{
		AccountID: &accountID,
		Name:      body.Name,
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

	err = tx.Create(&admin).Error
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
	html, err := mjml.ToHTML(context.Background(), template.TemplateEmailInvitation(initializer, body.Name, tokenStr), mjml.WithMinify(true))
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

// Activate Account with token and password
// @Summary Validate Account with token
// @Description Validate Account with token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body IncomingActivate true "Activate Account"
// @Success 200 {object} response.MessageResponse "Successful validation"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/activate [post]
func ActivateAccount(ctx *fiber.Ctx, initializer *database.Initializer) error {

	var (
		body IncomingActivate
	)

	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	accountID, err := utils.ValidateToken(body.Token, initializer.DB)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	account, err := model.GetAccountByID(initializer.DB, accountID)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	if account.Status == constant.ACTIVE {
		logger.Log.Error("Account already active")
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Account already active"))
	}

	tx := initializer.DB.Begin()

	// update account status here
	err = UpdateAccountStatus(accountID, constant.ACTIVE, tx)
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(err.Error()))
	}

	// Remove token
	err = utils.UpdateTokenStatus(body.Token, tx)
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

	logger.Log.Info(constant.SuccessValidate)
	return ctx.Status(fiber.StatusOK).JSON(response.SuccessResponseBody(constant.SuccessValidate))
}

// @Summary Forgot Password
// @Tags Auth
// @Accept json
// @Produce json
// @Params body body IncomingForgotPassword true "Forgot password data"
// @Success 200 {object} response.MessageResponse "Successful forgot password"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/forgot [post]
func ForgotPassword(ctx *fiber.Ctx, initializer *database.Initializer) error {

	var (
		err  error
		body IncomingForgotPassword
	)

	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	account, err := model.GetAccountByEmail(initializer.DB, body.Email)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	tokenStr, err := utils.GenerateToken()
	if err != nil {
		logger.Log.Error("Error in generating token for ", account.ID)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody("Error in generating token"))
	}

	token := model.Token{
		AccountID: account.ID,
		Token:     tokenStr,
		ExpireAt:  time.Now().Add(time.Hour * 24 * 7), // 7 days expiry
		Type:      constant.RESET_PASSWORD_TOKEN,
		Status:    constant.PENDING,
	}

	tx := initializer.DB.Begin()

	err = tx.Create(&token).Error
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(err.Error()))
	}

	// Send email with token
	var name string
	if account.Role == constant.ROLE_COMPANY {
		name = account.Company.Name
	} else if account.Role == constant.ROLE_USER {
		name = account.User.FirstName + account.User.LastName
	}

	html, err := mjml.ToHTML(context.Background(), template.TemplateForgotPassword(initializer, name, tokenStr), mjml.WithMinify(true))
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Unable to convert MJML to HTML"))
	}
	subject := "Reset Password"
	var emptyFile *multipart.FileHeader
	err = mailer.SendEmail(html, subject, []string{body.Email}, emptyFile)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(response.ErrorResponseBody(err.Error()))
	}

	logger.Log.Info("Forgot password email sent to: ", body.Email)
	return ctx.Status(fiber.StatusOK).JSON(response.SuccessResponseBody("Forgot password email sent"))
}

// @Summary Reset Password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body IncomingResetPassword true "New password"
// @Success 200 {object} response.MessageResponse "Successful reset password"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/reset [post]
func ResetPassword(ctx *fiber.Ctx, initializer *database.Initializer) error {
	var (
		err  error
		body IncomingResetPassword
	)

	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	accountID, err := utils.ValidateToken(body.Token, initializer.DB)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	account, err := model.GetAccountByID(initializer.DB, accountID)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	if account.Status == constant.ACTIVE {
		logger.Log.Error("Account already active")
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Account already active"))
	}

	hashPassword, err := HashPassword(body.Password)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Unable to hash password"))
	}

	tx := initializer.DB.Begin()

	// update account password here
	err = tx.Model(&model.Account{}).
		Where("id = ?", accountID).
		Update("password", hashPassword).
		Error

	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(err.Error()))
	}

	// Remove token
	err = utils.UpdateTokenStatus(body.Token, tx)
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

	logger.Log.Info("Password reset for: ", accountID)
	return ctx.Status(fiber.StatusOK).JSON(response.SuccessResponseBody("Password reset successfully"))
}

// @Summary Get Profile
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.DataResponse{data=model.ResponseProfile} "Successful get profile"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/profile [get]
func GetProfile(ctx *fiber.Ctx, initializer *database.Initializer) error {
	accountID := ctx.Locals("id").(uuid.UUID)

	account, err := model.GetProfileByAccountID(initializer.DB, accountID)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}
	//-----------------
	cache.Redis.SetCacheByIdForId("profile", "", accountID.String(), account, time.Minute*30)

	return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(account, "Successfully get profile"))

	//---------------Need to do for User
}

type IncomingUpdateUser struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

// @Summary Update User Profile
// @Description Update User Profile
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body IncomingUpdateUser true "Profile"
// @Success 200 {object} response.DataResponse{data=model.ResponseProfile} "Successful update profile"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/user [patch]
func UpdateUserProfile(ctx *fiber.Ctx, initializer *database.Initializer) error {
	var (
		accountID = ctx.Locals("id").(uuid.UUID)
		body      IncomingUpdateUser
		tx        = initializer.DB.Begin()
	)

	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	err := tx.Model(&model.User{}).
		Where("account_id = ?", accountID).
		Updates(model.User{
			FirstName: body.FirstName,
			LastName:  body.LastName,
		}).Error
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	tx.Commit()

	profile, err := model.GetProfileByAccountID(initializer.DB, accountID)
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	cache.Redis.SetCacheByIdForId("profile", "", accountID.String(), profile, time.Minute*30)

	return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(profile, "Successfully update profile"))
}

type IncomingUpdateCompany struct {
	Name string `json:"name" validate:"required"`
}

// @Summary Update Company Profile
// @Description Update Company Profile
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body IncomingUpdateCompany true "Profile"
// @Success 200 {object} response.DataResponse{data=model.ResponseProfile} "Successful update profile"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/company [patch]
func UpdateCompanyProfile(ctx *fiber.Ctx, initializer *database.Initializer) error {
	var (
		accountID = ctx.Locals("id").(uuid.UUID)
		body      IncomingUpdateCompany
		tx        = initializer.DB.Begin()
	)

	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	err := tx.Model(&model.Company{}).
		Where("account_id = ?", accountID).
		Updates(model.Company{
			Name: body.Name,
		}).Error
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	tx.Commit()

	profile, err := model.GetProfileByAccountID(initializer.DB, accountID)
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	cache.Redis.SetCacheByIdForId("profile", "", accountID.String(), profile, time.Minute*30)

	return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(profile, "Successfully update profile"))
}

type IncomingChangePassword struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
}

// @Summary Change Password
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body IncomingChangePassword true "Change Password"
// @Success 200 {object} response.MessageResponse "Successful change password"
// @Failure 400 {object} response.MessageResponse "Bad request"
// @Failure 500 {object} response.MessageResponse "Internal server error"
// @Router /auth/change [patch]
func ChangePassword(ctx *fiber.Ctx, initializer *database.Initializer) error {
	var (
		accountID = ctx.Locals("id").(uuid.UUID)
		body      IncomingChangePassword
		account   model.Account
		tx        = initializer.DB.Begin()
	)

	if err := utils.ValidateParser(&body, ctx, constant.VALIDATE); err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	err := tx.Where(&model.Account{ID: accountID}).First(&account).Error
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	if !CheckPasswordHash(body.CurrentPassword, account.Password) {
		logger.Log.Error("Invalid password by: ", accountID)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Invalid password"))
	}

	newPasswordHash, err := HashPassword(body.NewPassword)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody("Unable to hash password"))
	}

	err = tx.Model(&model.Account{}).
		Where("id = ?", accountID).
		Update("password", newPasswordHash).
		Error

	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}

	tx.Commit()

	return ctx.Status(fiber.StatusOK).JSON(response.SuccessResponseBody("Successfully change password"))
}
