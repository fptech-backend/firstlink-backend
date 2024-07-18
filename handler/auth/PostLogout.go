package handler_auth

import (
	"certification/cache"
	"certification/constant"
	"certification/database"
	"certification/logger"
	"certification/response"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
