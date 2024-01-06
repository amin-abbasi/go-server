package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

var (
	db        = make(map[string]User)
	secretKey = []byte("your_secret_key") // Replace with your secret key
)

// User struct to represent a user
type User struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func getUser(ctx *fiber.Ctx) error {
	username := ctx.Params("name")
	value, ok := db[username]
	if ok {
		return ctx.JSON(map[string]interface{}{"username": username, "value": value})
	}
	return ctx.JSON(map[string]interface{}{"username": username, "status": "no value"})
}

func login(ctx *fiber.Ctx) error {
	var user User
	if err := ctx.BodyParser(&user); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{"error": err.Error()})
	}

	// Get user from DB
	storedUser, ok := db[user.UserName]

	// Validate credentials (this is a simple example, use your authentication logic here)
	if ok && storedUser.Password == user.Password {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["username"] = user.UserName
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expiration time

		tokenString, err := token.SignedString(secretKey)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(map[string]interface{}{"error": "error generating token"})
		}

		// adds user in db
		db[user.UserName] = user

		return ctx.JSON(map[string]interface{}{"token": tokenString})
	}

	return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{"error": "invalid credentials"})
}

func createUser(ctx *fiber.Ctx) error {
	var user User
	if err := ctx.BodyParser(&user); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(map[string]interface{}{"error": err.Error()})
	}

	// Check if the username already exists in the db
	if _, exists := db[user.UserName]; exists {
		return ctx.Status(http.StatusConflict).JSON(map[string]interface{}{"error": "username already exists"})
	}

	// adds user in db
	db[user.UserName] = user
	return ctx.JSON(map[string]interface{}{"status": "ok", "user": user})
}

func authMiddleware(ctx *fiber.Ctx) error {
	tokenString := ctx.Get("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if tokenString == "" {
		return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{"error": "missing token"})
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return ctx.Status(http.StatusUnauthorized).JSON(map[string]interface{}{"error": "invalid token"})
	}

	claims := token.Claims.(jwt.MapClaims)
	ctx.Locals("claims", claims)
	return ctx.Next()
}

func main() {
	db["amin"] = User{UserName: "amin", Password: "123"}

	app := fiber.New()

	// Routes
	app.Get("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendString("pong")
	})
	app.Get("/user/:name", getUser)
	app.Post("/login", login)

	authorized := app.Group("/admin")
	authorized.Use(authMiddleware)
	authorized.Post("/user", createUser)

	// Start server
	app.Listen(":4000")
}
