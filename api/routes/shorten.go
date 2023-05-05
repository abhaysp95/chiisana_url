package routes

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/abhaysp95/chiisana_url/api/database"
	"github.com/abhaysp95/chiisana_url/api/helpers"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	rIpdb := database.CreateClient(1)
	defer rIpdb.Close()

	ipVal, err := rIpdb.Get(database.Ctx, ctx.IP()).Result()
	if err == redis.Nil {
		_ = rIpdb.Set(database.Ctx, ctx.IP(), os.Getenv("API_QUOTA"), time.Minute*30)
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": "Problem with database"})
	} else {
		valInt, err := strconv.Atoi(ipVal)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem parsing rate"})
		}
		if valInt <= 0 {
			limit, _ := rIpdb.TTL(database.Ctx, ctx.IP()).Result()
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":           fmt.Sprintf("Rate limit exceeded: %v", ctx.IP()),
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

	// dealing with custom short and thus making the short url
	var id string
	if body.ShortAs != "" {
		id = body.ShortAs
	} else {
		id = uuid.NewString()[:7]
	}

	rdb := database.CreateClient(0)
	defer rdb.Close()

	_, err = rdb.Get(database.Ctx, id).Result()
	if err == redis.Nil {
		// store the new url
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem with database"})
	} else { // can't use this custom short url
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Provided custom short URL is already in use"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24 // set default expiry
	}

	err = rdb.Set(database.Ctx, id, body.URL, body.Expiry*time.Second*3600).Err()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem in storing URL to db"})
	}

	rIpdb.Decr(database.Ctx, ctx.IP()) // one more url shortend for "this" ip

	return nil
}
