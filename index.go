package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
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

// Initialize the Gin router and define routes
func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", pingHandler)
	router.GET("/user/:name", getUserHandler)
	router.POST("/login", loginHandler)

	authorized := router.Group("/admin")
	authorized.Use(authMiddleware())
	authorized.POST("/user", createUserHandler)

	return router
}

// Ping handler
func pingHandler(ctx *gin.Context) {
	ctx.String(http.StatusOK, "pong")
}

// Get user by name handler
func getUserHandler(ctx *gin.Context) {
	username := ctx.Param("name")
	value, ok := db[username]
	if ok {
		ctx.JSON(http.StatusOK, gin.H{"username": username, "value": value})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"username": username, "status": "no value"})
	}
}

// Login handler
func loginHandler(ctx *gin.Context) {
	var user User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	storedUser, ok := db[user.UserName]
	if ok && storedUser.Password == user.Password {
		tokenString := generateToken(user.UserName)
		db[user.UserName] = user
		ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	}
}

// Generate JWT token
func generateToken(username string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expiration time

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		panic("error generating token")
	}

	return tokenString
}

// Create user handler
func createUserHandler(ctx *gin.Context) {
	var user User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, exists := db[user.UserName]; exists {
		ctx.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	db[user.UserName] = user
	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "user": user})
}

// JWT authentication middleware
func authMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			ctx.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func main() {
	db["amin"] = User{UserName: "amin", Password: "123"}
	router := setupRouter()
	router.Run(":4000")
}
