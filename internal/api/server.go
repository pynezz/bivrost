package api

import (
	"fmt"
	"log"
	"time"

	"github.com/pynezz/bivrost/internal/config"
	"github.com/pynezz/bivrost/internal/middleware"
	"github.com/pynezz/bivrost/internal/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

// JWTSecretKey is the secret key used to sign JWT tokens
// This will not be hardcoded in the build version

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
	argon2Instance := middleware.NewArgon2().InitArgonWithSalt(middleware.GetSecretKey(), "saltsalt")

	// Configure the fiber server with values from the config file
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(cfg.Network.ReadTimeout) * time.Second,  // Convert seconds to time.Duration
		WriteTimeout: time.Duration(cfg.Network.WriteTimeout) * time.Second, // TODO: Add a way to check if the config values are valid
	})

	output := fmt.Sprintf(
		"Server started with\n\tread timeout: %d\n\twrite timeout: %d\n",
		cfg.Network.ReadTimeout, cfg.Network.WriteTimeout)

	fmt.Println(output)

	// Middleware
	app.Use(logger.New()) // Log every request

	fmt.Printf("Secret key: %s%s%s\n",
		util.LightYellow, argon2Instance.GetPrintableKeyWithSalt(argon2Instance.Salt), util.Reset)

	fmt.Printf("Argon2 hash: %s%s%s\n", util.LightYellow, argon2Instance.GetEncodedHash(), util.Reset)

	// For every path except the root, check if the user is authenticated
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() != "/" && c.Path() != "/login" && c.Path() != "/register" {
			return middleware.Bouncer()(c)
		}
		return c.Next()
	})

	// Setup routes
	setupRoutes(app, cfg)

	return app
}

// The antiphishingHandler function is a function that handles the antiphishing route.
// Theoretical implementation is as follows:
// 1. The user generates a hash based on their WebGL fingerprint + url
// 2. The user sends the hash to the server along with the JWT token
// 3. The server set the hash as the session key with the JWT token as the value
// 4. The server sends the hash back to the user
// 5. The user sends the hash to the server with every elevated request and at random intervals
// 6. The server checks if the hash is in the session store
// 7. If the hash is in the session store, the server allows the request
// 8. If the hash is not in the session store, the server denies the request, alerts the user to reauthenticate, and logs the event
func antiphishingHandler(c *fiber.Ctx) error {
	h := c.Get("anti-phish") // Get the anti-phishing header
	if h == "" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Check if the hash is in the session store
	// middleware.DBInstance.Fetch(middleware.DBInstance.SelectColEq(sess))
	// If it is, allow the request
	// If it is not, deny the request

	return c.SendStatus(fiber.StatusOK)
}

// setupRoutes configures all the routes for the API server.
func setupRoutes(app *fiber.App, cfg *config.Cfg) {
	app.Get("/", indexHandler)               // Root path
	app.Get("/ws", websocket.New(wsHandler)) // WebSockets

	app.Post("/config/add_source", func(c *fiber.Ctx) error {
		c.Accepts("application/yaml", "application/json")
		// Serialize the request body to a struct
		var configRequest ConfigRequest
		if err := c.BodyParser(&configRequest.Fields.Sources); err != nil {
			util.PrintError("Error parsing request body: " + err.Error())
			return err
		}

		updatedFields := c.Body()

		fmt.Println("ID" + util.ColorF(util.DarkYellow, "%s", updatedFields))

		// TODO: Check if this works
		config.WriteConfig(cfg, "config.yaml")

		return c.SendString("Configuration updated")
	})

	// For the ThreatIntel module
	app.Post("/api/v1/intel/", func(c *fiber.Ctx) error {
		c.AcceptsEncodings("application/json")
		payload := new(IntelRequest)
		if err := c.BodyParser(payload); err != nil {
			return err
		}
		fmt.Println(payload)
		return c.JSON(payload)
	})

	// Auth route
	// For this to work, the client must send a GET request to /auth/<id>
	// With body: Bearer Token <key>

	app.Get("/auth/:id", func(c *fiber.Ctx) error {
		q := c.Queries()
		fmt.Println("[i] Query parameters")
		for k, v := range q {
			fmt.Printf("[i] Key: %s, Value: %s\n", k, v)
		}

		key := q["key"]
		fmt.Println("Key: ", key)

		return c.SendStatus(fiber.StatusUnauthorized)
	})

	app.Post("/login", middleware.BeginLogin)
	app.Post("/register", middleware.BeginRegistration)
	app.Get("/auth", handleAuth)
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
