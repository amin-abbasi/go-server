package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var db = make(map[string]User)
var secretKey = []byte("your_secret_key") // Replace with your secret key

// Define a struct
type User struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		username := c.Params.ByName("name")
		value, ok := db[username]
		if ok {
			c.JSON(http.StatusOK, gin.H{"username": username, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"username": username, "status": "no value"})
		}
	})

	r.POST("/login", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": "error generating token"})
				return
			}

			// adds user in db
			db[user.UserName] = user

			c.JSON(http.StatusOK, gin.H{"token": tokenString})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		}
	})

	// Protected endpoints using JWT authentication middleware
	authorized := r.Group("/admin")
	authorized.Use(authMiddleware())

	// Protected endpoint to create a new user
	authorized.POST("/user", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Gets the authorized user [admin]
		// claims := c.MustGet("claims").(jwt.MapClaims)
		// username := claims["username"].(string)

		// Check if the username already exists in the db
		if _, exists := db[user.UserName]; exists {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}

		// adds user in db
		db[user.UserName] = user
		c.JSON(http.StatusOK, gin.H{"status": "ok", "user": user})
	})

	return r
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("claims", claims)
		c.Next()
	}
}

func main() {
	db["amin"] = User{UserName: "amin", Password: "123"}
	r := setupRouter()
	r.Run(":4000")
}
