package app

import (
	"git.mobiledep.ru/flagshtok/backend/ishd/integrator/service"
	"git.mobiledep.ru/flagshtok/backend/preprocessor/handler"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/config"
	"github.com/rs/zerolog"
)

type serviceProvider struct {
	pgConfig     config.PGConfig
	serverConfig config.ServerConfig
	tokenConfig  config.TokenConfig

	dbClient      db.Client
	txManager     db.TxManager
	appRepository *repository.Repository

	appService *service.Service

	tokenMaker *token.JWTMaker

	log zerolog.Logger

	handler *handler.Handler
}
