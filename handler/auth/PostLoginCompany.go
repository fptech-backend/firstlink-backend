package handler_auth

import (
	"certification/cache"
	"certification/constant"
	"certification/database"
	"certification/logger"
	model_account "certification/model/account"
	"certification/response"
	"certification/utils"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

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

	account, err := model_account.GetAccountByEmail(initializer.DB, body.Email)
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
