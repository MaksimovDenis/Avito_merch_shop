package transaction

import (
	"context"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/client/db/pg"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type manager struct {
	db db.Transactor
}

func NewTransactionsManager(db db.Transactor) db.TxManager {
	return &manager{
		db: db,
	}
}

func (mgr *manager) transaction(ctx context.Context, opts pgx.TxOptions, fnc db.Handler) (err error) {
	tx, ok := ctx.Value(pg.TxKey).(pgx.Tx)
	if ok {
		return fnc(ctx)
	}

	tx, err = mgr.db.BeginTx(ctx, opts)

	defer func() {
		if rec := recover(); rec != nil {
			err = errors.Errorf("panica recovered: %v", rec)
		}

		if err != nil {
			if errRollBack := tx.Rollback(ctx); errRollBack != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollBack)
			}
			return
		}

		if nil == err {
			err = tx.Commit(ctx)
			if err != nil {
				err = errors.Wrapf(err, "tx commit failed")
			}
		}
	}()

	if err = fnc(ctx); err != nil {
		err = errors.Wrapf(err, "failed executing code inside transaction")
	}

	return err
}

func (mgr *manager) ReadCommitted(ctx context.Context, fnc db.Handler) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return mgr.transaction(ctx, txOpts, fnc)
}
