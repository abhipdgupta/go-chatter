package main

import (
	"go-chatter/db"
	"go-chatter/handlers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	err := db.ConnectDb()
	if err != nil {
		log.Println("Error while connecting to database", err)
	}
}

func main() {
	app := fiber.New()

	handlers.RegisterUserRoutes(app)
	app.Get("/health-check", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message":     "Server is up and running",
			"status_code": fiber.StatusNoContent,
			"data":        nil,
		})
	})

	log.Fatal(app.Listen(":9090"))
}
