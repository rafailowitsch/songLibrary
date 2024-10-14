package app

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"songLibrary/internal/config"
	deliveryHttp "songLibrary/internal/delivery/http"
	musicapi "songLibrary/internal/delivery/music_info"
	"songLibrary/internal/repository"
	"songLibrary/internal/repository/postgres"
	redi "songLibrary/internal/repository/redis"
	"songLibrary/internal/service"
	"songLibrary/pkg/logger/handlers/slogpretty"
	"songLibrary/pkg/logger/sl"
	"songLibrary/pkg/migrator"
	"syscall"
	"time"

	_ "songLibrary/docs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

const migrationsDir = "migrations"

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// Run starts the application
func Run() {
	// load configuration
	cfg := config.MustLoad()

	// setup logger
	log := setupLogger(cfg.Env)
	log.Info("starting song library", slog.String("env", cfg.Env))

	// setup context and handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setup signal handler for graceful shutdown
	gracefulShutdown(ctx, cancel, log)

	// connect to PostgreSQL
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Address,
		cfg.Postgres.DBName,
	)
	log.Info("connecting to PostgreSQL", slog.String("connection_string", connString))

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Error("unable to parse PostgreSQL connection config", sl.Err(err))
		os.Exit(1)
	}

	conn, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error("unable to establish connection to PostgreSQL", sl.Err(err))
		os.Exit(1)
	}
	defer conn.Close()

	log.Info("PostgreSQL connection established")

	// apply database migrations
	applyMigrations(log, connString)

	// connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	pong, err := client.Ping(ctx).Result()
	if err != nil {
		log.Error("unable to connect to Redis", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Redis connection established", slog.String("ping", pong))

	// create new music service API client
	musicServiceAPI := musicapi.NewMusicInfo(cfg.MusicInfo.Address, log)
	log.Info("music service address", slog.String("address", cfg.MusicInfo.Address))

	// create repositories, services, and handlers
	db := postgres.NewPostgres(conn)
	cache := redi.NewRedis(client)
	repo := repository.NewRepository(db, cache, log)
	service := service.NewService(repo, musicServiceAPI, log)
	handler := deliveryHttp.NewHandler(service, log)

	// start HTTP server
	startServer(handler, cfg, log)

	// wait for graceful shutdown
	<-ctx.Done()
	log.Info("shutting down gracefully")
}

// applyMigrations applies database migrations
func applyMigrations(log *slog.Logger, connString string) {
	sqlDB, err := sql.Open("postgres", connString)
	if err != nil {
		log.Error("unable to open SQL connection", sl.Err(err))
		os.Exit(1)
	}
	defer sqlDB.Close()

	migr := migrator.MustGetNewMigrator(MigrationsFS, migrationsDir)
	if err := migr.ApplyMigrations(sqlDB); err != nil {
		log.Error("failed to apply migrations", sl.Err(err))
		os.Exit(1)
	}

	log.Info("migrations applied successfully")
}

// startServer starts the HTTP server and swagger
func startServer(handler *deliveryHttp.Handler, cfg *config.Config, log *slog.Logger) {
	routes := handler.InitRoutes()
	routes.Get("/swagger/*", httpSwagger.WrapHandler)
	log.Info("swagger documentation available")

	srv := &http.Server{
		Addr:    cfg.HTTP.Address,
		Handler: routes,
	}

	go func() {
		log.Info("server started on address", slog.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", sl.Err(err))
			os.Exit(1)
		}
	}()
}

// gracefulShutdown handles the graceful shutdown process
func gracefulShutdown(ctx context.Context, cancel context.CancelFunc, log *slog.Logger) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Info("received shutdown signal, shutting down gracefully")
		cancel()

		timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 5*time.Second)
		defer timeoutCancel()

		// Wait for all processes to finish gracefully
		<-timeoutCtx.Done()
		log.Info("shutdown complete")
	}()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
