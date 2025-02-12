package repository

import (
	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/rs/zerolog"
)

type Repository struct {
	Authorization
}

func NewRepository(db db.Client, log zerolog.Logger) *Repository {
	return &Repository{Authorization: newAuthRepository(db, log)}
}
