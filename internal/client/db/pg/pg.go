package pg

import (
	"context"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type key string

const (
	TxKey key = "tx"
)

type pg struct {
	dbc *pgxpool.Pool
}

func NewDB(dbc *pgxpool.Pool) db.DB {
	return &pg{
		dbc: dbc,
	}
}

func (pg *pg) ScanOneContext(ctx context.Context, dest interface{}, quer db.Query, args ...interface{}) error {
	row, err := pg.QueryContext(ctx, quer, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, row)
}

func (pg *pg) ScanAllContext(ctx context.Context, dest interface{}, quer db.Query, args ...interface{}) error {
	rows, err := pg.QueryContext(ctx, quer, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}

func (pg *pg) ExecContext(ctx context.Context, quer db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, quer.QueryRow, args...)
	}

	return pg.dbc.Exec(ctx, quer.QueryRow, args...)
}

func (pg *pg) QueryContext(ctx context.Context, quer db.Query, args ...interface{}) (pgx.Rows, error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, quer.QueryRow, args...)
	}

	return pg.dbc.Query(ctx, quer.QueryRow, args...)
}

func (pg *pg) QueryRowContext(ctx context.Context, quer db.Query, args ...interface{}) pgx.Row {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, quer.QueryRow, args...)
	}

	return pg.dbc.QueryRow(ctx, quer.QueryRow, args...)
}

func (p *pg) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.dbc.BeginTx(ctx, txOptions)
}

func (p *pg) Ping(ctx context.Context) error {
	return p.dbc.Ping(ctx)
}

func (pg *pg) Close() {
	pg.dbc.Close()
}

func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}
