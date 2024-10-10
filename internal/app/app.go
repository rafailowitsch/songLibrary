package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"songLibrary/internal/config"
	handler "songLibrary/internal/delivery/http"
	musicapi "songLibrary/internal/delivery/music_info"
	"songLibrary/internal/repository"
	"songLibrary/internal/repository/postgres"
	redi "songLibrary/internal/repository/redis"
	"songLibrary/internal/service"
	"songLibrary/pkg/logger/logger/handlers/slogpretty"
	"songLibrary/pkg/logger/logger/sl"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func app() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"starting song library",
		slog.String("env", cfg.Env),
	)
	log.Debug("debug messages are enabled")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
	)

	log.Info("Connect string: %s", connString)

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Error("unable to establish connection:", sl.Err(err))
		os.Exit(1)
	}
	defer conn.Close(ctx)
	log.Info("Connection established")

	err = postgres.CreateTables(ctx, *conn)
	if err != nil {
		log.Error("unable to create tables: ", sl.Err(err))
		os.Exit(1)
	}
	log.Info("Tables created")

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	musicServiceAPI := musicapi.NewMusicInfo(cfg.MusicInfo.Host + cfg.MusicInfo.Port)

	db := postgres.NewSongDB(conn)
	cache := redi.NewRedis(client)

	repo := repository.NewRepository(db, cache)
	service := service.NewService(repo, musicServiceAPI, log)
	handler := handler.NewHandler(service, log)

	var srv http.Server
	srv.Handler = handler.InitRoutes()
	srv.Addr = cfg.HTTP.Host + ":" + cfg.HTTP.Port
	srv.ListenAndServe()
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
