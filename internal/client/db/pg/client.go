package pg

import (
	"context"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type pgClient struct {
	masterDBC db.DB
}

func New(ctx context.Context, dsn string) (db.Client, error) {
	dbc, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, errors.Errorf("failed to connect db: %v", err)
	}

	if err := migration(dsn); err != nil {
		return nil, errors.Errorf("failed to init migrations: %v", err)
	}

	return &pgClient{
		masterDBC: &pg{dbc: dbc},
	}, nil
}

func (clt *pgClient) DB() db.DB {
	return clt.masterDBC
}

func (clt *pgClient) Close() error {
	if clt.masterDBC != nil {
		clt.masterDBC.Close()
	}

	return nil
}
