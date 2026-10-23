package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"ems/app/httperr"
	"ems/config"
	"ems/database"
	"ems/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config.Load()

	database.Connect(config.AppConfig)
	defer database.Close()

	database.RunMigrations()
	database.SeedDefaultRBAC()
	database.SeedAdminUser()

	app := fiber.New(fiber.Config{
		AppName:      "Recruitment Management System",
		ErrorHandler: httperr.Handler,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: config.AppConfig.AllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Security headers
	app.Use(func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		return c.Next()
	})

	app.Use(logger.New())

	app.Static("/static", config.AppConfig.FrontendPath+"/static")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile(config.AppConfig.FrontendPath + "/static/index.html")
	})

	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "image/x-icon")
		c.Set("Cache-Control", "public, max-age=86400")
		return c.SendFile(config.AppConfig.FrontendPath + "/static/favicon.ico")
	})

	routes.SetupRoutes(app)

	go func() {
		port := ":" + config.AppConfig.ServerPort
		if err := app.Listen(port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := app.Shutdown(); err != nil {
		log.Fatal("Server shutdown error:", err)
	}
}
