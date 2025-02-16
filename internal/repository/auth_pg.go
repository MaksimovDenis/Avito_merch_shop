package repository

import (
	"context"
	"strings"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/models"
	"github.com/Masterminds/squirrel"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Authorization interface {
	CreateUser(ctx context.Context, user models.AuthReq) (models.User, error)
	GetUser(ctx context.Context, username string) (models.User, error)
}

type AuthRepo struct {
	db  db.Client
	log zerolog.Logger
}

func newAuthRepository(db db.Client, log zerolog.Logger) *AuthRepo {
	return &AuthRepo{
		db:  db,
		log: log,
	}
}

func (ar *AuthRepo) CreateUser(ctx context.Context, user models.AuthReq) (models.User, error) {
	var res models.User

	builder := squirrel.Insert("users").
		PlaceholderFormat(squirrel.Dollar).
		Columns("username", "password_hash").
		Values(user.Username, user.Password).
		Suffix("RETURNING id, username, coins")

	query, args, err := builder.ToSql()
	if err != nil {
		ar.log.Error().Err(err).Msg("CreateUser: failed to build SQL query")
		return res, err
	}

	queryStruct := db.Query{
		Name:     "auth_repository.CreateUser",
		QueryRow: query,
	}

	err = ar.db.DB().QueryRowContext(ctx, queryStruct, args...).
		Scan(&res.Id, &res.Username, &res.Coins)
	if err != nil {
		ar.log.Error().Err(err).Msg("CreateUser: failed to execute query")
		return res, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return res, nil
}

func (ar *AuthRepo) GetUser(ctx context.Context, username string) (models.User, error) {
	var res models.User

	builder := squirrel.Select("*").
		PlaceholderFormat(squirrel.Dollar).
		From("users").
		Where(squirrel.Eq{"username": username})

	query, args, err := builder.ToSql()
	if err != nil {
		ar.log.Error().Err(err).Msg("GetUser: failed to build SQL query")
		return res, err
	}

	queryStruct := db.Query{
		Name:     "user_repository.GetUser",
		QueryRow: query,
	}

	err = ar.db.DB().QueryRowContext(ctx, queryStruct, args...).
		Scan(&res.Id, &res.Username, &res.Password, &res.Coins)
	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		ar.log.Warn().Str("username", username).Msg("GetUser: user not found")

		return res, status.Errorf(codes.NotFound, "User not found")
	} else if err != nil {
		ar.log.Error().Err(err).Msg("GetUser: failed to execute query")

		return res, status.Errorf(codes.Internal, "Internal server error")
	}

	return res, nil
}
