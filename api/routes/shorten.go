package routes

import (
	"net/url"
	"time"

	"github.com/abhaysp95/chiisana_url/api/helpers"
	"github.com/gofiber/fiber/v2"
)

type request struct {
	URL     string        `json:"url"`
	ShortAs string        `json:"short_as"`
	Expiry  time.Duration `json:"expiry"`
}

type response struct {
	URL            string        `json:"url"`
	ShortAs        string        `json:"short_as"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining uint          `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(ctx *fiber.Ctx) error {
	body := &request{}

	// parse the request body
	if err := ctx.BodyParser(body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// enforce rate limiting

	// check for request passed for URL shortening contains actual URL
	if _, err := url.ParseRequestURI(body.URL); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error":"Invalid URL passed"})
	}

	// check domain error
	if !helpers.ResolveDomainError(body.URL) {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error":"Domain error"})
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	return nil
}
