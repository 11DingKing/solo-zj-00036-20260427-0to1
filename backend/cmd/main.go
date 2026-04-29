package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"survey-platform/internal/cache"
	"survey-platform/internal/config"
	"survey-platform/internal/database"
	"survey-platform/internal/handler"
	"survey-platform/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	cfg := config.Load()

	gin.SetMode(cfg.Server.GinMode)

	if err := database.InitDB(&cfg.DB); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB()
	log.Println("Database connected successfully")

	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed")

	if err := cache.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer cache.CloseRedis()
	log.Println("Redis connected successfully")

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authHandler := handler.NewAuthHandler(&cfg.JWT)
	surveyHandler := handler.NewSurveyHandler()
	questionHandler := handler.NewQuestionHandler()
	logicHandler := handler.NewLogicHandler()
	responseHandler := handler.NewResponseHandler()
	statsHandler := handler.NewStatsHandler()

	authMiddleware := middleware.JWTAuthMiddleware(&cfg.JWT)
	optionalAuth := middleware.OptionalJWTAuth(&cfg.JWT)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authMiddleware, authHandler.GetCurrentUser)
		}

		surveys := api.Group("/surveys")
		{
			surveys.GET("", authMiddleware, surveyHandler.List)
			surveys.POST("", authMiddleware, surveyHandler.Create)
			surveys.GET("/:id", authMiddleware, surveyHandler.Get)
			surveys.PUT("/:id", authMiddleware, surveyHandler.Update)
			surveys.DELETE("/:id", authMiddleware, surveyHandler.Delete)
			surveys.GET("/:id/fill", optionalAuth, surveyHandler.GetForFill)

			surveys.POST("/:id/questions", authMiddleware, questionHandler.Create)
			surveys.PUT("/questions/:question_id", authMiddleware, questionHandler.Update)
			surveys.DELETE("/questions/:question_id", authMiddleware, questionHandler.Delete)
			surveys.PUT("/:id/questions/order", authMiddleware, questionHandler.UpdateOrder)

			surveys.GET("/:id/logic", authMiddleware, logicHandler.List)
			surveys.POST("/:id/logic", authMiddleware, logicHandler.Create)
			surveys.DELETE("/logic/:logic_id", authMiddleware, logicHandler.Delete)
			surveys.PUT("/:id/logic/order", authMiddleware, logicHandler.UpdateOrder)

			surveys.GET("/:id/stats", authMiddleware, statsHandler.GetSurveyStats)
			surveys.GET("/:id/questions-stats", authMiddleware, statsHandler.GetQuestionStats)
			surveys.GET("/:id/crosstab", authMiddleware, statsHandler.CrosstabAnalysis)
			surveys.GET("/:id/export/csv", authMiddleware, statsHandler.ExportCSV)
			surveys.GET("/:id/export/excel", authMiddleware, statsHandler.ExportExcel)
		}

		responses := api.Group("/responses")
		{
			responses.POST("/start", optionalAuth, responseHandler.Start)
			responses.POST("/submit", responseHandler.Submit)
		}
	}

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", serverAddr)

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
