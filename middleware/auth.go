package middleware

import (
	"certification/cache"
	"certification/config"
	"certification/constant"
	"certification/database"
	handler_auth "certification/handler/auth"
	"certification/logger"
	"certification/response"
	"certification/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func ValidateToken(initializer *database.Initializer) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Retrieve token from authorization header
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			errMsg := "Missing authorization header"
			logger.Log.Error(errMsg)
			return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(errMsg))
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errMsg := "Invalid authorization header"
			logger.Log.Error(errMsg)
			return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(errMsg))
		}

		// Retrieve token from JWT
		token, err := jwt.Parse(parts[1], JWTKeyFunc)
		if err != nil || !token.Valid {
			errMsg := "Invalid token"
			logger.Log.Error(errMsg, token)
			return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(errMsg))
		}

		// Retrieve token from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			errMsg := "Failed to extract claims from token"
			logger.Log.Error(errMsg, token)
			return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(errMsg))
		}

		// Retrieve user ID & email from token
		id, ok := claims["id"]
		if !ok {
			errMsg := "Failed to extract user ID from token"
			logger.Log.Error(errMsg, token)
			return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(errMsg))
		}

		profile_id, ok := claims["profile_id"]
		if !ok {
			errMsg := "Failed to extract user profile ID from token"
			logger.Log.Error(errMsg, token)
			return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(errMsg))
		}

		email, ok := claims["email"]
		if !ok {
			errMsg := "Failed to extract user's email from token"
			logger.Log.Error(errMsg, token)
			return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(errMsg))
		}

		// Retrieve permissions from Redis
		results, err := cache.Redis.RDB.Get(context.Background(), id.(string)).Result()
		if err != nil {
			errMsg := "Failed to extract permission from redis"
			logger.Log.Error(errMsg, id)
			return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody(errMsg))
		}

		var data utils.RedisValue
		err = json.Unmarshal([]byte(results), &data)
		if err != nil {
			logger.Log.Error("Error in decoding string to json for ID %s during log in", id, results)
			return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(
				"You are not logged in. Token is not valid",
			))
		}

		// Validate the redis status, if it's updated, return an error
		if data.Status == constant.UPDATED {
			logger.Log.Error("The status is updated")
			return ctx.Status(fiber.StatusUnauthorized).JSON(response.ErrorResponseBody(
				"Alterations have been made to your data. Kindly proceed to log in once more.",
			))
		}

		ParseModulePermission(ctx, data.Module)

		id, err = uuid.Parse(id.(string))
		if err != nil {
			logger.Log.Error("Failed to parse ID from token")
			return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody("Failed to parse ID from token"))
		}

		profile_id, err = uuid.Parse(profile_id.(string))
		if err != nil {
			logger.Log.Error("Failed to parse profile ID from token")
			return ctx.Status(fiber.StatusInternalServerError).JSON(response.ErrorResponseBody("Failed to parse profile ID from token"))
		}

		ctx.Locals("id", id)
		ctx.Locals("profile_id", profile_id)
		ctx.Locals("email", email)

		return ctx.Next()
	}
}

func JWTKeyFunc(token *jwt.Token) (interface{}, error) {
	_, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(config.SECRET), nil
}

func ParseModulePermission(ctx *fiber.Ctx, accessesJSON []map[string]interface{}) {
	accesses := make([]handler_auth.Access, len(accessesJSON))

	// Store module access and permission in Locals
	for i, accessJSON := range accessesJSON {
		moduleIDStr, _ := accessJSON["module_id"].(string)
		moduleAccess, _ := accessJSON["module_access"].(bool)
		readAccess, _ := accessJSON["read_access"].(bool)
		writeAccess, _ := accessJSON["write_access"].(bool)
		deleteAccess, _ := accessJSON["delete_access"].(bool)

		accesses[i] = handler_auth.Access{
			ModuleID:     moduleIDStr,
			ModuleAccess: moduleAccess,
			ReadAccess:   readAccess,
			WriteAccess:  writeAccess,
			DeleteAccess: deleteAccess,
		}

		ctx.Locals(moduleIDStr, accesses[i])
	}
}
