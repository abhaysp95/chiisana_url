package routes

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/abhaysp95/chiisana_url/database"
	"github.com/abhaysp95/chiisana_url/helpers"
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
	log.Printf("body: %v", body)

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
		return ctx.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": fmt.Sprintf("Problem getting IP rate from database: %v", err)})
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
	if _, err := url.Parse(body.URL); err != nil {
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
		err = rdb.Set(database.Ctx, id, body.URL, body.Expiry*time.Second*3600).Err()
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem in storing URL to db"})
		}
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem  getting id from database"})
	} else { // can't use this custom short url
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Provided custom short URL is already in use"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24 // set default expiry
	}

	rIpdb.Decr(database.Ctx, ctx.IP()) // one more url shortend for "this" ip

	// get remaining rate for the requesting IP
	remaining_rate, err := rIpdb.Get(database.Ctx, ctx.IP()).Result()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem getting remaining rate"})
	}

	// get TTL for the url being returned (cooling period)
	rate_ttl, err := rdb.TTL(database.Ctx, ctx.IP()).Result()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Problem getting cooling period"})
	}

	rate_left, err := strconv.Atoi(remaining_rate)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Rate parsing error"})
	}

	resp := response {
		URL: body.URL,
		ShortAs: id,
		Expiry: body.Expiry,
		XRateRemaining: uint(rate_left),
		XRateLimitRest: rate_ttl / (time.Nanosecond * time.Minute),
	}

	return ctx.Status(fiber.StatusOK).JSON(resp)
}
