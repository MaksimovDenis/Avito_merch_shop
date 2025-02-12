package service

import (
	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/repository"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/rs/zerolog"
)

type Service struct {
	Authorization Authorization
}

func NewService(repos repository.Repository, txManager db.TxManager, token token.JWTMaker, log zerolog.Logger) *Service {
	return &Service{
		Authorization: newAuthService(repos, txManager, token, log),
	}
}
