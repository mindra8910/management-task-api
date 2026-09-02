package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	handler "task-api/internal/delivery/http"
	"task-api/internal/delivery/http/midlleware"
	"task-api/internal/pkg/response"
	"task-api/internal/repository/postgres"
	"task-api/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type dailyLogger struct {
	dir string
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (l *dailyLogger) Write(p []byte) (n int, err error) {
	os.MkdirAll(l.dir, 0755)
	logFileName := filepath.Join(l.dir, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return file.Write(p)
}

func main() {
	gin.SetMode(gin.ReleaseMode)

	// Setup logging to daily files in "logs" folder
	logger := &dailyLogger{dir: "logs"}
	log.SetOutput(io.MultiWriter(logger, os.Stdout))
	gin.DefaultWriter = io.MultiWriter(logger, os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(logger, os.Stderr)

	router := gin.New()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "postgres://root:root@localhost:5432/taskdb?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Setup middlewares
	router.Use(StructuredLogger())
	router.Use(PanicRecovery())

	// Setup repositories
	userRepo := postgres.NewPostgresUserRepository(db)
	taskRepo := postgres.NewPostgresTaskRepository(db)

	// Setup usecases
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretkey"
	}
	userUsecase := usecase.NewUserUsecase(userRepo, jwtSecret)
	taskUsecase := usecase.NewTaskUsecase(taskRepo)

	authMiddleware := midlleware.AuthMiddleware([]byte(jwtSecret))

	// Setup handlers (routes)
	handler.NewUserHandler(router, userUsecase)
	handler.NewTaskHandler(router, taskUsecase, authMiddleware)

	// API Routes Placeholder
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("Server running on port 8080")
	router.Run(":8080")
}

func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.InternalServerError(c, response.CodeInternalServer, "An unexpected error occurred")
				log.Printf("PANIC: %v", err)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-Id", requestID)
		c.Next()
	}
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // salin response body
	return w.ResponseWriter.Write(b)
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		requestID := c.GetString("request_id")

		// Baca request body (untuk log)
		var reqBody []byte
		if c.Request.Body != nil {
			reqBody, _ = io.ReadAll(c.Request.Body)
			// Restore body agar bisa dibaca handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// Bungkus ResponseWriter
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// Ambil response body
		responseBody := blw.body.String()

		// Struktur log JSON
		logEntry := map[string]interface{}{
			"request_id":   requestID,
			"method":       method,
			"path":         path,
			"status":       statusCode,
			"latency_ms":   latency.Milliseconds(),
			"client_ip":    clientIP,
			"request_body": string(reqBody),
			"response":     responseBody, // bisa di-truncate jika terlalu panjang
		}

		// Log sebagai JSON
		jsonLog, _ := json.Marshal(logEntry)
		log.Println(string(jsonLog))
	}
}
