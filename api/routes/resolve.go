package routes

import (
	"fmt"

	"github.com/abhaysp95/chiisana_url/database"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func ResolveURL(ctx *fiber.Ctx) error {
	shortURL := ctx.Params("url")

	rdb := database.CreateClient(0)
	defer rdb.Close()

	value, err := rdb.Get(database.Ctx, shortURL).Result()
	if err == redis.Nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Provided short url not found"})
		// log.Println("couldn't found actual url for \"", shortURL, "\"")
	} else if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Problem with database: %v", err)})
	}

	rIncrdb := database.CreateClient(1)
	defer rIncrdb.Close()

	rIncrdb.Incr(database.Ctx, ctx.IP())

	return ctx.Redirect(value, 301)
}
