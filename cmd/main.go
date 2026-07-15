package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GroVlAn/auth-access/internal/config"
	grpcHandler "github.com/GroVlAn/auth-access/internal/handler/grpc-handler"
	httpHandler "github.com/GroVlAn/auth-access/internal/handler/http-handler"
	"github.com/GroVlAn/auth-access/internal/infrastructure/database"
	"github.com/GroVlAn/auth-access/internal/infrastructure/preloader"
	"github.com/GroVlAn/auth-access/internal/repository"
	grpcServer "github.com/GroVlAn/auth-access/internal/server/grpc-server"
	httpServer "github.com/GroVlAn/auth-access/internal/server/http-server"
	"github.com/GroVlAn/auth-access/internal/service"
	"github.com/GroVlAn/auth-base/ew/httpx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

const (
	localConfigPath = "configs/config-local.yml"
)

func main() {
	timeStart := time.Now()

	l := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := config.LoadEnv(); err != nil {
		l.Fatal().Err(err).Msg("failed to load env variables")
	}

	configPath := flag.String("config", localConfigPath, "Path to the configuration file")
	defRolesConfigPath := flag.String("config-def-roles", "", "Path to the default roles configuration file")
	flag.Parse()

	cfg, err := config.New(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load configuration")
	}

	db, err := database.NewPostgresqlDB(database.PostgresSettings{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		Username: cfg.DB.Username,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.DBName,
		SSLMode:  cfg.DB.SSLMode,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to database")
	}

	r := repository.New(db)

	if len(*defRolesConfigPath) > 0 {
		pr := preloader.New(r, preloader.Deps{
			DefRolesConfigPath: *defRolesConfigPath,
		})

		loadDefaultRoles(ctx, l, cfg, pr)
	}

	s := service.New(r)

	h := httpHandler.New(l, s, httpHandler.Deps{
		BasePath:       cfg.HTTP.BaseHTTPPath,
		DefaultTimeout: cfg.Settings.DefaultTimeout,
	})

	gh := grpcHandler.New(l, s, cfg.Settings.DefaultTimeout)

	hServer := httpServer.New(
		h.Handler(),
		httpServer.Settings{
			Port:              cfg.HTTP.Port,
			MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
			ReadHeaderTimeout: time.Duration(cfg.HTTP.ReadHeaderTimeout) * time.Second,
			WriteTimeout:      time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		},
	)

	gServer := grpcServer.New(gh)

	go func() {
		if err := hServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	go func() {
		l.Info().Msgf("grpc server started on port: %s", cfg.GRPC.Port)

		if err := gServer.ListenAndServe(cfg.GRPC.Port); err != nil {
			l.Fatal().Err(err).Msg("failed to start grpc server")
		}
	}()

	l.Info().Msgf("server start on port: %s load time: %v", cfg.HTTP.Port, time.Since(timeStart))

	<-ctx.Done()
	err = hServer.Shutdown(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to shutdown server")
	} else {
		l.Info().Msg("server shutdown gracefully")
	}
}

func loadDefaultRoles(ctx context.Context, l zerolog.Logger, cfg *config.Config, preloader *preloader.Preloader) {
	ctxR, cancelR := context.WithTimeout(ctx, cfg.Settings.DefaultTimeout)
	defer cancelR()

	err := preloader.Preload(ctxR)

	if err != nil {
		respErr := httpx.HandleError(err)
		l.Fatal().Err(err).Msg(respErr.LogMsg)
	}
}
