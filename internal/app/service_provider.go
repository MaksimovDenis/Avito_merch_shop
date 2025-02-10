package app

import (
	"os"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/config"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type serviceProvider struct {
	pgConfig     config.PGConfig
	serverConfig config.ServerConfig
	tokenConfig  config.TokenConfig

	dbClient  db.Client
	txManager db.TxManager
	//appRepository *repository.Repository

	//appService *service.Service

	tokenMaker *token.JWTMaker

	log zerolog.Logger

	//handler *handler.Handler
}

func newServiceProvider() *serviceProvider {
	srv := &serviceProvider{}
	srv.log = srv.initLogger()

	return srv
}

func (srv *serviceProvider) initLogger() zerolog.Logger {
	logFile, err := os.OpenFile("./internal/logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open log file")
	}

	logLevel, err := zerolog.ParseLevel("debug")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse log level")
	}

	// Записываем логи и в файл, и в консоль
	multiWriter := zerolog.MultiLevelWriter(os.Stdout, logFile)

	logger := zerolog.New(multiWriter).Level(logLevel).With().Timestamp().Logger()
	return logger
}
