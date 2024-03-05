package api

import (
	"fmt"
	"log"

	"github.com/pynezz/bivrost/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

type app struct {
	*fiber.App
	fiber.Route
}

// NewServer initializes a new API server with the provided configuration.
func NewServer(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		// Fiber configuration options here
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Middleware
	app.Use(logger.New()) // Log every request

	// Setup routes
	setupRoutes(app, cfg)

	return app
}

// setupRoutes configures all the routes for the API server.
func setupRoutes(app *fiber.App, cfg *config.Config) {
	app.Get("/", indexHandler)

	// WebSocket route
	app.Get("/ws", websocket.New(wsHandler))

	// Threat Intel API routes
	// app.Get("/api/v1/threats", getThreatsHandler)

}

// indexHandler handles the root path.
func indexHandler(c *fiber.Ctx) error {
	return c.SendString("Bivrost Fiber API Server is up and running!")
}

// wsHandler handles WebSocket connections.
func wsHandler(c *websocket.Conn) {
	// WebSocket connection handling logic here
	fmt.Println("WebSocket Connected")

	// Example: Echo received message back to the client
	var (
		mt  int
		msg []byte
		err error
	)
	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", msg)
		if err = c.WriteMessage(mt, msg); err != nil {
			log.Println("write:", err)
			break
		}
	}
}
