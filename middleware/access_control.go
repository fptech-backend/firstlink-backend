package middleware

import (
	"certification/constant"
	handler_auth "certification/handler/auth"
	"certification/logger"
	"certification/response"

	"github.com/gofiber/fiber/v2"
)

func ValidatePermission(moduleID string, permission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {

		err := ValidateModulePermission(ctx, moduleID, permission)
		if err.Message != "" {
			logger.Log.Error(err)
			return ctx.Status(fiber.StatusForbidden).JSON(err)
		}

		return ctx.Next()
	}
}

func ValidateModulePermission(ctx *fiber.Ctx, moduleID string, permission string) response.MessageResponse {
	access, ok := ctx.Locals(moduleID).(handler_auth.Access)
	if !ok {
		return response.AccessDeniedResponseBody(ctx.Locals("id").(string))
	}

	isAccessPass := access.ReadAccess // Read access only

	switch permission {
	case constant.WRITE:
		isAccessPass = access.ReadAccess && access.WriteAccess
	case constant.DELETE:
		isAccessPass = access.ReadAccess && access.DeleteAccess
	}

	if !isAccessPass {
		return response.AccessDeniedResponseBody(ctx.Locals("id").(string))
	}

	return response.MessageResponse{}
}
