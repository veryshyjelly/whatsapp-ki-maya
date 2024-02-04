package api

import (
	"github.com/gofiber/fiber/v2"
)

func Login() func(c *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		return ctx.SendFile("scan.png")
	}
}