package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abhaysp95/chiisana_url/api/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	// setting up router
	app := fiber.New()
	setupRoutes(app)

	// loading up .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Printf("key: MY_SECRET_KEY, value: %v\n", os.Getenv("MY_SECRET_KEY"));

	log.Fatal(app.Listen(":8080"))
}
