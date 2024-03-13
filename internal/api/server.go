package api

import (
	"fmt"
	"log"
	"time"

	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/internal/middleware"
	"github.com/pynezz/bivrost/internal/util"
	"github.com/pynezz/bivrost/internal/util/crypto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

type App struct {
	*fiber.App
}

type IntelRequest struct {
	Id string `json:"id"`
	Ip string `json:"ip"`
}

type ConfigRequest struct {
	Fields config.Cfg `text:"id" json:"id" yaml:"id"`
}

// NewServer initializes a new API server with the provided configuration.
// Renamed config.Config to config.Cfg to avoid confusion with the Fiber Config struct
func NewServer(cfg *config.Cfg) *fiber.App {

	// Configure the fiber server with values from the config file
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(cfg.Network.ReadTimeout) * time.Second,  // Convert seconds to time.Duration
		WriteTimeout: time.Duration(cfg.Network.WriteTimeout) * time.Second, // TODO: Add a way to check if the config values are valid
	})

	output := fmt.Sprintf("Server started with\n\tread timeout: %d\n\twrite timeout: %d\n", cfg.Network.ReadTimeout, cfg.Network.WriteTimeout)
	fmt.Println(output)

	// Middleware
	app.Use(logger.New()) // Log every request

	// Generate a secure secret key for JWT authentication. This shoukld be done for every login request
	secretKey, err := crypto.GenerateSecretKey() // I know this is not properly implemented, but it's just for testing purposes
	if err != nil {
		log.Fatalf("Error generating secret key: %v", err)
	}

	// Base64 encode the secret key
	// key := base64.StdEncoding.EncodeToString([]byte(secretKey))
	fmt.Println(util.ColorF(util.DarkYellow, "Secret key: %s", secretKey))

	// app.Use(middleware.AuthMiddleware(secretKey))
	app.Get("/dashboard", middleware.Bouncer(secretKey))

	// Setup routes
	setupRoutes(app, cfg)

	return app
}

// setupRoutes configures all the routes for the API server.
func setupRoutes(app *fiber.App, cfg *config.Cfg) {
	app.Get("/", indexHandler)

	// WebSocket route
	app.Get("/ws", websocket.New(wsHandler))

	app.Post("/config/add_source", func(c *fiber.Ctx) error {
		c.Accepts("application/yaml", "application/json")
		// Serialize the request body to a struct
		var configRequest ConfigRequest
		if err := c.BodyParser(&configRequest.Fields.Sources); err != nil {
			return err
		}

		updatedFields := c.Body()

		fmt.Println("ID" + util.ColorF(util.DarkYellow, "%s", updatedFields))

		// TODO: Write the updated configuration to the file

		return c.SendString("Configuration updated")
	})

	config.WriteConfig(cfg, "config.yaml")

	app.Post("/api/v1/intel/", func(c *fiber.Ctx) error {
		c.AcceptsEncodings("application/json")
		payload := new(IntelRequest)
		if err := c.BodyParser(payload); err != nil {
			return err
		}
		fmt.Println(payload)
		return c.JSON(payload)
	})

	app.Get("/auth/:id", func(c *fiber.Ctx) error {
		q := c.Queries()
		fmt.Println("[i] Query parameters")
		for k, v := range q {
			fmt.Printf("[i] Key: %s, Value: %s\n", k, v)
		}

		key := q["key"]
		fmt.Println("Key: ", key)

		// This section should be placed in a separate function or in the auth middleware
		if c.Params("id") == "test" {
			token, err := middleware.GenerateToken("test", key)
			if err != nil {
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			response := map[string]string{
				"status": "ok",
				"token":  token,
			}

			return c.SendString("Authenticated. Here's your session JWT: " + response["token"])
		}
		// ----------------------------

		return c.SendStatus(fiber.StatusUnauthorized)
	})

	// Threat Intel API routes
	// app.Get("/api/v1/threats", getThreatsHandler)

}

// isAuthenticated checks if the user is authenticated.
func isAuthenticated(c *fiber.Ctx) bool {
	// Check if the user is authenticated
	return true
}

func handleAuth(c *fiber.Ctx) error {
	// Check if the user is authenticated
	if !isAuthenticated(c) {
		// If not, redirect to the login page
		return c.Redirect("/login")
	}

	// If the user is authenticated, call the next handler
	return c.Next()
}

func updateConfigHandler(c *fiber.Ctx) error {
	fmt.Println(string(c.Body()))

	return c.SendString("Configuration updated")
}

// indexHandler handles the root path.
func indexHandler(c *fiber.Ctx) error {
	c.Accepts("html")                           // "html"
	c.Accepts("text/html")                      // "text/html"
	c.Accepts("json", "text")                   // "json"
	c.Accepts("application/json")               // "application/json"
	c.Accepts("text/plain", "application/json") // "application/json", due to quality
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
