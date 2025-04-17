package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Benedict-Bsl/beautiful-routes" // Import your package
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	app := fiber.New()

	// Basic routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API Server")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Route that returns all routes
	// app.Get("/routes", func(c *fiber.Ctx) error {
	// 	routes := app.GetRoutes()
	// 	return c.JSON(routes)
	// })
	
	// API v1 routes
	v1 := app.Group("/api/v1")
	
	// Email endpoints
	email := v1.Group("/email")
	email.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"emails": []string{}})
	})
	
	email.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{"id": id, "status": "delivered"})
	})
	
	email.Post("/send", func(c *fiber.Ctx) error {
		var req EmailRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}
		return c.Status(201).JSON(EmailResponse{
			ID:      "email123",
			Status:  "queued",
			Message: "Email queued for delivery",
		})
	})
	
	email.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{"id": id, "deleted": true})
	})
	
	// SMS endpoints
	sms := v1.Group("/sms")
	sms.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"sms": []string{}})
	})
	
	sms.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{"id": id, "status": "delivered"})
	})
	
	// Set up fiberdocs
	docsConfig := fiberdocs.Config{
		Title:       "Notification API",
		Description: "API for sending emails, SMS, and managing alerts and reminders",
		Version:     "2.0.0",
		BasePath:    "https://api.example.com",
		UIPath:      "/redocs",
	}
	
	// Add extra documentation info for specific routes
	// docsConfig.AddRouteInfo("/api/v1/email/send", map[string]interface{}{
	// 	"description": "Send a new email",
	// 	"requestBody": EmailRequest{
	// 		To:      "user@example.com",
	// 		Subject: "Hello",
	// 		Body:    "This is a test email",
	// 	},
	// 	"responses": map[string]interface{}{
	// 		"201": map[string]interface{}{
	// 			"description": "Email queued successfully",
	// 			"content": EmailResponse{
	// 				ID:      "email123",
	// 				Status:  "queued",
	// 				Message: "Email queued for delivery",
	// 			},
	// 		},
	// 		"400": map[string]interface{}{
	// 			"description": "Invalid request",
	// 		},
	// 	},
	// })
	
	// Add the docs middleware
	app.Use(fiberdocs.New(docsConfig))
	
	// Start the server
	app.Listen(":3000")
}