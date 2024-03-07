package api

import (
	"fmt"
	"log"
	"time"

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
		ReadTimeout:  time.Duration(cfg.Network[1].ReadTimeout) * time.Second, // Convert seconds to time.Duration
		WriteTimeout: time.Duration(cfg.Network[1].WriteTimeout) * time.Second,

		// Allow methods
		RequestMethods: []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
	})

	output := fmt.Sprintf("Server started with\n\tread timeout: %d\n\twrite timeout: %d\n", cfg.Network[0].ReadTimeout, cfg.Network[1].WriteTimeout)
	fmt.Println(output)

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

	app.Post("/config/:id", updateConfigHandler)

	// Threat Intel API routes
	// app.Get("/api/v1/threats", getThreatsHandler)

}

func updateConfigHandler(c *fiber.Ctx) error {
	// Update the configuration here
	id := c.Params("id")
	fmt.Println("Updating configuration for ID:", id)
	return c.SendString("Configuration updated")
}

// indexHandler handles the root path.
func indexHandler(c *fiber.Ctx) error {
	c.Accepts("html")                           // "html"
	c.Accepts("text/html")                      // "text/html"
	c.Accepts("json", "text")                   // "json"
	c.Accepts("application/json")               // "application/json"
	c.Accepts("text/plain", "application/json") // "application/json", due to quality
	// c.Accepts("POST")                           // ""
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

// func newRoute(method string, path string, handler func(*fiber.Ctx) error) *fiber.Route {
// 	return &fiber.Route{
// 		Method:  method,
// 		Path:    path,
// 		Handler: handler,
// 	}
// }
