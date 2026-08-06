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
	"github.com/GroVlAn/auth-access/internal/infrastructure/secrets"
	vaultClient "github.com/GroVlAn/auth-access/internal/infrastructure/vault-client"
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

	configPath := flag.String("config", localConfigPath, "Path to the configuration file")
	defRolesConfigPath := flag.String("config-def-roles", "", "Path to the default roles configuration file")
	flag.Parse()

	cfg, err := config.New(*configPath)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load configuration")
	}

	vc, err := vaultClient.New(vaultClient.Conf{
		SecretToken: cfg.Vault.SecretToken,
		Address:     cfg.Vault.Address,
		Mount:       cfg.Vault.Mount,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load vault client")
	}

	provider := secrets.New(vc, secrets.Paths{
		Postgres: cfg.VaultPaths.Postgres,
	})

	scrt, err := provider.Load(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("failed load secrets")
	}

	db, err := database.NewPostgresqlDB(database.PostgresSettings{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		Username: scrt.Postgres.Username,
		Password: scrt.Postgres.Password,
		DBName:   scrt.Postgres.DBName,
		SSLMode:  cfg.DB.SSLMode,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			l.Error().Err(err).Msg("failed to close postgresql db connection")
		}
	}()

	r := repository.New(db)

	if len(*defRolesConfigPath) > 0 {
		l.Info().Msg("starting load default roles")
		pr := preloader.New(r, preloader.Deps{
			DefRolesConfigPath: *defRolesConfigPath,
		})

		loadDefaultRoles(ctx, l, cfg, pr)
	}

	s := service.New(r)

	h := httpHandler.New(
		l,
		s,
		httpHandler.Deps{
			BasePath:       cfg.HTTP.BaseHTTPPath,
			DefaultTimeout: cfg.Settings.DefaultTimeout,
		},
	)

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

	errCh := make(chan error, 2)

	go func() {
		l.Info().Msgf("starting http server on port: %s", cfg.HTTP.Port)

		err := hServer.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		l.Info().Msgf("starting grpc server on port: %s", cfg.GRPC.Port)

		if err := gServer.ListenAndServe(cfg.GRPC.Port); err != nil {
			errCh <- err
		}

	}()

	l.Info().
		Dur("startup_time", time.Since(timeStart)).
		Str("http_port", cfg.HTTP.Port).
		Str("grpc_port", cfg.GRPC.Port).
		Msg("server started")

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := hServer.Shutdown(shutdownCtx); err != nil {
			l.Error().Err(err).Msg("failed to shutdown server")
		} else {
			l.Info().Msg("server shutdown gracefully")
		}
		gServer.Stop()
	}

	select {
	case <-ctx.Done():
		shutdown()
	case err := <-errCh:
		if err != nil {
			l.Error().Err(err).Msg("server exited with error")

			shutdown()
		}
	}
}

func loadDefaultRoles(ctx context.Context, l zerolog.Logger, cfg *config.Config, preloader *preloader.Preloader) {
	ctxR, cancelR := context.WithTimeout(ctx, cfg.Settings.DefaultTimeout)
	defer cancelR()

	err := preloader.Preload(ctxR)

	if err != nil {
		respErr := httpx.HandleError(l, err)
		l.Fatal().Err(err).Msg(respErr.LogMsg)
	}
}
