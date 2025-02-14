package repository

import (
	"context"
	"strings"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/client/db/pg"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Shop interface {
	UpdateBalanceForPurchase(ctx context.Context, userId int, productName string) (*int, error)
	InsertPurchaseRecord(ctx context.Context, userId int, productId int) error
	SendCoin(ctx context.Context, sender string, receiver string, amount int) error
}

type ShopRepo struct {
	db  db.Client
	log zerolog.Logger
}

func newShopRepository(db db.Client, log zerolog.Logger) *ShopRepo {
	return &ShopRepo{
		db:  db,
		log: log,
	}
}

func (sr *ShopRepo) UpdateBalanceForPurchase(ctx context.Context, userId int, productName string) (*int, error) {
	updateQuery := squirrel.Update("users").PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("users.coins - products.price")).
		From("products").
		Where(
			squirrel.Eq{"users.id": userId, "products.name": productName},
			squirrel.Expr("users.coins >= products.price"),
		).
		Suffix("RETURNING products.id")

	query, args, err := updateQuery.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("UpdateBalanceForPurchase: failed to build update SQL query")

		return nil, errors.New("недостаточно монет для покупки")
	}

	queryStruct := db.Query{
		Name:     "user_repository.UpdateBalanceForPurchase",
		QueryRow: query,
	}

	var productId int

	err = sr.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&productId)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "violates check constraint"):
			sr.log.Error().Err(err).Msg("UpdateBalanceForPurchase: not enough coins")
			return nil, errors.New("недостаточно средств для покупки")
		case strings.Contains(err.Error(), "no rows in result set"):
			sr.log.Error().Err(err).Msgf("UpdateBalanceForPurchase: item %v not found", productName)
			return nil, errors.New("товар " + productName + " не найден")
		default:
			sr.log.Error().Err(err).Msg("UpdateBalanceForPurchase: failed to update user data")
			return nil, errors.New("ошибка при обновлении данных")
		}
	}

	return &productId, nil
}

func (sr *ShopRepo) InsertPurchaseRecord(ctx context.Context, userId int, productId int) error {
	insertQuery := squirrel.Insert("purchases").PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "products_id").
		Values(userId, productId)

	query, args, err := insertQuery.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("InsertPurchaseRecord: failed to build insert SQL query")
		return err
	}

	queryStruct := db.Query{
		Name:     "user_repository.InsertPurchaseRecord",
		QueryRow: query,
	}

	_, err = sr.db.DB().ExecContext(ctx, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("InsertPurchaseRecord: failed to insert purchase")
		return err
	}

	return nil
}

func (sr *ShopRepo) SendCoin(ctx context.Context, sender string, receiver string, amount int) error {
	tx, err := sr.db.DB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		sr.log.Error().Err(err).Msg("BuyItem: failed to start transaction")
		return err
	}

	ctxWithTx := pg.MakeContextTx(ctx, tx)

	// CHECK SENDER BALANCE
	selectQueryBalance := squirrel.Select("id", "coins").
		PlaceholderFormat(squirrel.Dollar).
		From("users").
		Where(squirrel.Eq{"username": sender})

	query, args, err := selectQueryBalance.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to build select SQL query")
		_ = tx.Rollback(ctx)

		return err
	}

	queryStruct := db.Query{
		Name:     "user_repository.SendCoin",
		QueryRow: query,
	}

	var coins int
	var senderId int

	err = sr.db.DB().QueryRowContext(ctxWithTx, queryStruct, args...).Scan(&coins, &senderId)
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to get sender balance")
		_ = tx.Rollback(ctx)

		return err
	}

	if coins < amount {
		_ = tx.Rollback(ctx)
		return errors.Errorf("недостаточно денег")
	}

	// UPDATE SENEDER BALANCE
	updateQuerySender := squirrel.Update("users").PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("coins - ?", amount)).
		From("users").
		Where(squirrel.Eq{"username": sender})

	query, args, err = updateQuerySender.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to build update SQL query")
		_ = tx.Rollback(ctx)

		return err
	}

	queryStruct = db.Query{
		Name:     "user_repository.SendCoin",
		QueryRow: query,
	}

	_, err = sr.db.DB().ExecContext(ctxWithTx, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to update sender balance")
		_ = tx.Rollback(ctx)

		return err
	}

	// UPDATE RECEIVER BALANCE
	updateQueryReceiver := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("coins + ?", amount)).
		From("users").
		Where(squirrel.Eq{"username": receiver}).
		Suffix("RETURNING id")

	query, args, err = updateQueryReceiver.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to build update SQL query")
		_ = tx.Rollback(ctx)

		return err
	}

	queryStruct = db.Query{
		Name:     "user_repository.SendCoin",
		QueryRow: query,
	}

	var receiverId int

	err = sr.db.DB().QueryRowContext(ctxWithTx, queryStruct, args...).Scan(&receiverId)
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to update reciver balance")
		_ = tx.Rollback(ctx)

		return err
	}

	// INSERT TRANSACTION
	insertQueryTransact := squirrel.Insert("transactions").
		PlaceholderFormat(squirrel.Dollar).
		Columns("sender_id", "receiver_id", "amount").
		Values(senderId, receiverId, amount)

	query, args, err = insertQueryTransact.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to build update SQL query")
		_ = tx.Rollback(ctx)

		return err
	}

	queryStruct = db.Query{
		Name:     "user_repository.SendCoin",
		QueryRow: query,
	}

	_, err = sr.db.DB().ExecContext(ctxWithTx, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("SendCoin: failed to add transaction")
		_ = tx.Rollback(ctx)

		return err
	}

	return tx.Commit(ctx)
}
