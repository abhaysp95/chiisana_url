package routes

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/abhaysp95/chiisana_url/api/database"
	"github.com/abhaysp95/chiisana_url/api/helpers"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
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
	rdb := database.CreateClient(1)
	defer rdb.Close()

	ipVal, err := rdb.Get(database.Ctx, ctx.IP()).Result()
	if err == redis.Nil {
		_ = rdb.Set(database.Ctx, ctx.IP(), os.Getenv("API_QUOTA"), time.Minute*30)
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": "Problem with database"})
	} else {
		valInt, err := strconv.Atoi(ipVal)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem parsing rate"})
		}
		if valInt <= 0 {
			limit, _ := rdb.TTL(database.Ctx, ctx.IP()).Result()
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fmt.Sprintf("Rate limit exceeded: %v", ctx.IP()),
				"rate_limit_rest": limit / (time.Nanosecond * time.Minute),
			})
		}
	}

	// check for request passed for URL shortening contains actual URL
	if _, err := url.ParseRequestURI(body.URL); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL passed"})
	}

	// check domain error
	if !helpers.ResolveDomainError(body.URL) {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Domain error"})
	}

	body.URL = helpers.EnforceHTTP(body.URL)

	rdb.Decr(database.Ctx, ctx.IP())  // one more url shortend for "this" ip

	return nil
}
