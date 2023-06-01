package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var (
	redisClient *redis.Client
)

func main() {
	// Kết nối tới Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Thay đổi địa chỉ Redis tùy thuộc vào cấu hình của bạn
		Password: "",               // Mật khẩu Redis (nếu có)
		DB:       0,                // Sử dụng database số 0
	})

	_, err := redisClient.Ping(redisClient.Context()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Đăng ký các đường dẫn (routes)
	router.POST("/login", handleLogin)
	router.Run(":8080")
}

func handlePing(c *gin.Context) {
	sessionId := c.GetHeader("sessionId")
	exists, err := redisClient.Exists(redisClient.Context(), sessionId).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check session existence"})
		return
	}
	if exists == 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "Another user is already pinging"})
		return
	}

}

func handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Kiểm tra thông tin đăng nhập (trong ví dụ này chỉ kiểm tra mật khẩu là "password")
	if password != "password" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Kiểm tra xem key đã tồn tại trong Redis chưa
	exists, err := redisClient.Exists(redisClient.Context(), username).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check key existence"})
		return
	}

	if exists == 1 {
		c.JSON(http.StatusConflict, gin.H{"error": "User already logged in"})
		return
	}

	// Tạo session ID ngẫu nhiên
	sessionID := generateSessionID()

	// Lưu thông tin session vào Redis
	err = redisClient.Set(redisClient.Context(), sessionID, username, time.Minute*5).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Lưu key vào Redis để đánh dấu người dùng đang đăng nhập
	err = redisClient.Set(redisClient.Context(), username, sessionID, 0).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user login status"})
		return
	}

	// Trả về thông tin session ID cho client
	c.JSON(http.StatusOK, gin.H{"sessionID": sessionID})
}
func generateSessionID() string {
	rand.Seed(time.Now().UnixNano())
	sessionID := strconv.Itoa(rand.Int())

	return sessionID
}
