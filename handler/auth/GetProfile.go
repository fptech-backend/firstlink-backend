package handler_auth

import (
	"certification/cache"
	"certification/database"
	"certification/logger"
	model_account "certification/model/account"
	"certification/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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

	account, err := model_account.GetProfileByAccountID(initializer.DB, accountID)
	if err != nil {
		logger.Log.Error(err)
		return ctx.Status(fiber.StatusBadRequest).JSON(response.ErrorResponseBody(err.Error()))
	}
	//-----------------
	cache.Redis.SetCacheByIdForId("profile", "", accountID.String(), account, time.Minute*30)

	return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(account, "Successfully get profile"))

	//---------------Need to do for User
}
