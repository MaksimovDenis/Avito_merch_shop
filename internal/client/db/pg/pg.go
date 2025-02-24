package pg

import (
	"context"
	"fmt"
	"log"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/client/db/pg/prettier"

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
	LogQuery(ctx, quer, args...)

	row, err := pg.QueryContext(ctx, quer, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanOne(dest, row)
}

func (pg *pg) ScanAllContext(ctx context.Context, dest interface{}, quer db.Query, args ...interface{}) error {
	LogQuery(ctx, quer, args...)

	rows, err := pg.QueryContext(ctx, quer, args...)
	if err != nil {
		return err
	}

	return pgxscan.ScanAll(dest, rows)
}

func (pg *pg) ExecContext(ctx context.Context, quer db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	LogQuery(ctx, quer, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, quer.QueryRow, args...)
	}

	return pg.dbc.Exec(ctx, quer.QueryRow, args...)
}

func (pg *pg) QueryContext(ctx context.Context, quer db.Query, args ...interface{}) (pgx.Rows, error) {
	LogQuery(ctx, quer, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, quer.QueryRow, args...)
	}

	return pg.dbc.Query(ctx, quer.QueryRow, args...)
}

func (pg *pg) QueryRowContext(ctx context.Context, quer db.Query, args ...interface{}) pgx.Row {
	LogQuery(ctx, quer, args...)

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

func LogQuery(ctx context.Context, q db.Query, args ...any) {
	prettyQuery := prettier.Pretty(q.QueryRow, prettier.PlaceholderDollar, args...)
	log.Println(
		ctx,
		fmt.Sprintf("sql: %s", q.Name),
		fmt.Sprintf("query: %s", prettyQuery),
	)
}
