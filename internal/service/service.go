package service

import (
	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/repository"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/rs/zerolog"
)

type Service struct {
	Authorization
	Shop
}

func NewService(repos repository.Repository, client db.Client, token token.JWTMaker, log zerolog.Logger) *Service {
	return &Service{
		Authorization: newAuthService(repos, token, log),
		Shop:          newShopService(repos, client, log),
	}
}
