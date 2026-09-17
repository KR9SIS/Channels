package handlers

import (
	"github.com/gofiber/fiber/v3"
)

func Welcome(ctx fiber.Ctx) error {
	return ctx.Render("Welcome", nil, "layouts/main")
}
