package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"pr-review/internal/config"
	"pr-review/internal/http/handlers"
	"pr-review/internal/repo/postgres"
	"pr-review/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := setupDatabase(ctx)
	defer closeDatabase(db)

	services := setupServices(db)
	handlers := setupHandlers(services)
	router := setupRouter(handlers)

	server := startServer(router)
	defer shutdownServer(server, db)
}

func setupDatabase(ctx context.Context) *pgxpool.Pool {
	dbConfig := config.LoadDatabaseConfig()

	db, err := config.ConnectDatabase(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established")
	return db
}

func closeDatabase(db *pgxpool.Pool) {
	log.Println("Closing database connection...")
	db.Close()
	log.Println("Database connection closed")
}

type Services struct {
	prService   *service.PullRequestService
	teamService *service.TeamService
	userService *service.UserService
}

func setupServices(db *pgxpool.Pool) *Services {
	teamRepo := postgres.NewTeamRepository(db)
	userRepo := postgres.NewUserRepository(db)
	prRepo := postgres.NewPullRequestRepository(db)

	prService := service.NewPullRequestService(prRepo, userRepo, teamRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo, prService)

	return &Services{
		prService:   prService,
		teamService: teamService,
		userService: userService,
	}
}

type Handlers struct {
	teamHandler *handlers.TeamHandler
	userHandler *handlers.UserHandler
	prHandler   *handlers.PullRequestHandler
}

func setupHandlers(services *Services) *Handlers {
	return &Handlers{
		teamHandler: handlers.NewTeamHandler(services.teamService),
		userHandler: handlers.NewUserHandler(services.userService),
		prHandler:   handlers.NewPullRequestHandler(services.prService),
	}
}

func setupRouter(handlers *Handlers) *gin.Engine {
	if getEnv("GIN_MODE", "release") == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	setupTeamRoutes(router, handlers.teamHandler)
	setupUserRoutes(router, handlers.userHandler)
	setupPullRequestRoutes(router, handlers.prHandler)

	return router
}

func setupTeamRoutes(router *gin.Engine, teamHandler *handlers.TeamHandler) {
	teamRoutes := router.Group("/team")
	{
		teamRoutes.POST("/add", teamHandler.Add)
		teamRoutes.GET("/get", teamHandler.Get)
	}
}

func setupUserRoutes(router *gin.Engine, userHandler *handlers.UserHandler) {
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("/setIsActive", userHandler.SetIsActive)
		userRoutes.GET("/getReview", userHandler.GetReview)
	}
}

func setupPullRequestRoutes(router *gin.Engine, prHandler *handlers.PullRequestHandler) {
	prRoutes := router.Group("/pullRequest")
	{
		prRoutes.POST("/create", prHandler.Create)
		prRoutes.POST("/merge", prHandler.Merge)
		prRoutes.POST("/reassign", prHandler.Reassign)
	}
}

func startServer(router *gin.Engine) *http.Server {
	host := getEnv("HOST", config.DefaultHTTPAddr)
	port := getEnv("PORT", "8080")
	serverAddr := host + ":" + port

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return server
}

func shutdownServer(server *http.Server, _ *pgxpool.Pool) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
