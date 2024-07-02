package router

import (
	"certification/database"
	"certification/handler"
	"certification/middleware"

	"github.com/gofiber/fiber/v2"
)

func AuthenticationRoutes(app *fiber.App, initializer *database.Initializer) {
	auth := app.Group("/auth")

	auth.Post("/login/user", func(c *fiber.Ctx) error {
		return handler.LoginUser(c, initializer, initializer.DB)
	})
	auth.Post("/login/company", func(c *fiber.Ctx) error {
		return handler.LoginCompany(c, initializer, initializer.DB)
	})
	auth.Post("/logout", middleware.ValidateToken(initializer), func(c *fiber.Ctx) error {
		return handler.Logout(c, initializer)
	})
	auth.Post("/signup/user", func(c *fiber.Ctx) error {
		return handler.SignUpUser(c, initializer)
	})
	auth.Post("/signup/company", func(c *fiber.Ctx) error {
		return handler.SignUpCompany(c, initializer)
	})
	auth.Post("/activate", func(c *fiber.Ctx) error {
		return handler.ActivateAccount(c, initializer)
	})
	auth.Post("/forgot", func(c *fiber.Ctx) error {
		return handler.ForgotPassword(c, initializer)
	})
	auth.Post("/reset", func(c *fiber.Ctx) error {
		return handler.ResetPassword(c, initializer)
	})
	auth.Get("/profile", middleware.ValidateToken(initializer), middleware.GetCacheByIdForMe("profile", ""), func(c *fiber.Ctx) error {
		return handler.GetProfile(c, initializer)
	})
	auth.Patch("/company", middleware.ValidateToken(initializer), func(c *fiber.Ctx) error {
		return handler.UpdateCompanyProfile(c, initializer)
	})
	auth.Patch("/user", middleware.ValidateToken(initializer), func(c *fiber.Ctx) error {
		return handler.UpdateUserProfile(c, initializer)
	})
	auth.Patch("/change", middleware.ValidateToken(initializer), func(c *fiber.Ctx) error {
		return handler.ChangePassword(c, initializer)
	})
}
