package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MaksimovDenis/Avito_merch_shop/internal/closer"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/config"

	"github.com/rs/zerolog/log"
)

type App struct {
	serviceProvider *serviceProvider
	httpServer      *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	app := &App{}

	err := app.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (app *App) Run() {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	app.runHTTPServer()
}

func (app *App) initDeps(ctx context.Context) error {
	inits := []func(ctx context.Context) error{
		app.initConfig,
		app.initServiceProvider,
		app.initHTTPServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initConfig(_ context.Context) error {
	err := config.Load(".env")
	if err != nil {
		return err
	}

	return nil
}

func (app *App) initServiceProvider(_ context.Context) error {
	app.serviceProvider = newServiceProvider()
	return nil
}

func (app *App) initHTTPServer(ctx context.Context) error {
	router := app.serviceProvider.AppHandler(ctx).InitRoutes()

	app.httpServer = &http.Server{
		Addr:    app.serviceProvider.ServerConfig().Address(),
		Handler: router,
	}

	return nil
}

func (app *App) runHTTPServer() {
	log.Printf("HTTP server is running on %s", app.httpServer.Addr)

	go func() {
		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("Could not listen on %s\n", app.httpServer.Addr)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx := context.Background()

	if err := app.httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("HTTP server Shutdown")
	}

	log.Logger.Println("HTTP server existing")
}
