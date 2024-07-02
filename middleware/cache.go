package middleware

import (
	"certification/cache"
	"certification/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetCache(key string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		cache, err := cache.Redis.GetCache(key)
		if err != nil {
			return ctx.Next()
		}

		return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(cache, "Cache retrieved successfully"))
	}
}

func GetCacheById(key, id string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		_id := ctx.Params(id)
		cache, err := cache.Redis.GetCacheById(key, _id)
		if err != nil {
			return ctx.Next()
		}
		return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(cache, "Cache retrieved successfully"))
	}
}

func GetCacheByIdForMe(key, id string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId := ctx.Locals("id").(uuid.UUID)

		cache, err := cache.Redis.GetCacheByIdForId(key, id, userId.String())
		if err != nil {
			return ctx.Next()
		}

		return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(cache, "Cache retrieved successfully"))
	}
}

func GetCacheByIdForId(key, id, user_id string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		cache, err := cache.Redis.GetCacheByIdForId(key, id, user_id)
		if err != nil {
			return ctx.Next()
		}

		return ctx.Status(fiber.StatusOK).JSON(response.DataResponseBody(cache, "Cache retrieved successfully"))
	}
}
