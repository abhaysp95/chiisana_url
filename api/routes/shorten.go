package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	URL     string        `json:"url"`
	ShortAs string        `json:"short_as"`
	Expiry  time.Duration `json:"expiry"`
}

type Response struct {
	URL            string        `json:"url"`
	ShortAs        string        `json:"short_as"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining uint          `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(ctx *fiber.Ctx) error {
	return nil
}
