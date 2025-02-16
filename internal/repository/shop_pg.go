package repository

import (
	"context"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	errresponse "github.com/MaksimovDenis/Avito_merch_shop/internal/err_response"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/models"
	"github.com/Masterminds/squirrel"
	"github.com/rs/zerolog"
)

type Shop interface {
	UpdateBalanceForPurchase(ctx context.Context, userId int, productName string) (*int, error)
	InsertPurchaseRecord(ctx context.Context, userId int, productId int) error
	UserBalanceByName(ctx context.Context, username string) (userId int, coins int, err error)
	UpdateSenderBalance(ctx context.Context, sender string, amount int) (senderId int, err error)
	UpdateReceiverBalance(ctx context.Context, receiver string, amount int) (receiverId int, err error)
	AddTransaction(ctx context.Context, senderId int, receiverId int, amount int) error
	GetItemsByUserId(ctx context.Context, userId int) ([]models.Items, error)
	SentCoinsByUserId(ctx context.Context, userId int) ([]models.SentCoins, error)
	ReceivedCoinsByUserId(ctx context.Context, userId int) ([]models.ReceivedCoins, error)
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

func (srp *ShopRepo) UpdateBalanceForPurchase(ctx context.Context, userId int, productName string) (*int, error) {
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
		srp.log.Error().Err(err).Msg("UpdateBalanceForPurchase: failed to build update SQL query")

		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UpdateBalanceForPurchase",
		QueryRow: query,
	}

	var productId int

	err = srp.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&productId)
	if err != nil {
		srp.log.Error().Err(err).Msg("UpdateBalanceForPurchase: failed to update user data")
		return nil, errresponse.ErrResponse(err, productName)
	}

	return &productId, nil
}

func (srp *ShopRepo) InsertPurchaseRecord(ctx context.Context, userId int, productId int) error {
	insertQuery := squirrel.Insert("purchases").PlaceholderFormat(squirrel.Dollar).
		Columns("user_id", "products_id").
		Values(userId, productId)

	query, args, err := insertQuery.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("InsertPurchaseRecord: failed to build insert SQL query")
		return errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.InsertPurchaseRecord",
		QueryRow: query,
	}

	_, err = srp.db.DB().ExecContext(ctx, queryStruct, args...)
	if err != nil {
		srp.log.Error().Err(err).Msg("InsertPurchaseRecord: failed to insert purchase")
		return errresponse.ErrResponse(err)
	}

	return nil
}

func (srp *ShopRepo) UserBalanceByName(ctx context.Context, username string) (userId int, coins int, err error) {
	selectQueryBalance := squirrel.Select("id", "coins").
		PlaceholderFormat(squirrel.Dollar).
		From("users").
		Where(squirrel.Eq{"username": username})

	query, args, err := selectQueryBalance.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("UserBalanceByName: failed to build SQL query")
		return 0, 0, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UserBalanceByName",
		QueryRow: query,
	}

	err = srp.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&userId, &coins)
	if err != nil {
		srp.log.Error().Err(err).Msg("UserBalanceByName: failed to get sender balance")

		return 0, 0, errresponse.ErrResponse(err, username)
	}

	return userId, coins, nil
}

func (srp *ShopRepo) UpdateSenderBalance(ctx context.Context, sender string, amount int) (senderId int, err error) {
	updateQuerySender := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("coins - ?", amount)).
		Where(squirrel.Eq{"username": sender}).
		Suffix("RETURNING id")

	query, args, err := updateQuerySender.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("UpdateSenderBalance: failed to build update SQL query")

		return 0, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UpdateSenderBalance",
		QueryRow: query,
	}

	err = srp.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&senderId)
	if err != nil {
		srp.log.Error().Err(err).Msg("UpdateSenderBalance: failed to update sender balance")

		return 0, errresponse.ErrResponse(err, sender)
	}

	return senderId, nil
}

func (srp *ShopRepo) UpdateReceiverBalance(ctx context.Context, receiver string, amount int) (
	receiverId int, err error) {
	updateQuerySender := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("coins + ?", amount)).
		Where(squirrel.Eq{"username": receiver}).
		Suffix("RETURNING id")

	query, args, err := updateQuerySender.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("UpdateReceiverBalance: failed to build update SQL query")

		return 0, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UpdateReceiverBalance",
		QueryRow: query,
	}

	err = srp.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&receiverId)
	if err != nil {
		srp.log.Error().Err(err).Msg("UpdateReceiverBalance: failed to update receiver balance")

		return 0, errresponse.ErrResponse(err, receiver)
	}

	return receiverId, nil
}

func (srp *ShopRepo) AddTransaction(ctx context.Context, senderId int, receiverId int, amount int) error {
	insertQueryTransact := squirrel.Insert("transactions").
		PlaceholderFormat(squirrel.Dollar).
		Columns("sender_id", "receiver_id", "amount").
		Values(senderId, receiverId, amount)

	query, args, err := insertQueryTransact.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("AddTransaction: failed to build update SQL query")
		return errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.AddTransaction",
		QueryRow: query,
	}

	_, err = srp.db.DB().ExecContext(ctx, queryStruct, args...)
	if err != nil {
		srp.log.Error().Err(err).Msg("AddTransaction: failed to add transaction")
		return errresponse.ErrResponse(err)
	}

	return nil
}

func (srp *ShopRepo) GetItemsByUserId(ctx context.Context, userId int) ([]models.Items, error) {
	var items []models.Items

	builder := squirrel.Select("p.name as name", "sum(pu.quantity) as quantity").
		PlaceholderFormat(squirrel.Dollar).
		From("purchases as pu").
		Join("products as p ON pu.products_id = p.id").
		Where(squirrel.Eq{"pu.user_id": userId}).
		GroupBy("name")

	query, args, err := builder.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("GetItemsByUserId: failed to build update SQL query")
		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.GetItemsByUserId",
		QueryRow: query,
	}

	err = srp.db.DB().ScanAllContext(ctx, &items, queryStruct, args...)
	if err != nil {
		srp.log.Error().Err(err).Msg("GetItemsByUserId: failed to scan rows")
		return nil, errresponse.ErrResponse(err)
	}

	return items, nil
}

func (srp *ShopRepo) SentCoinsByUserId(ctx context.Context, userId int) ([]models.SentCoins, error) {
	var sentCoins []models.SentCoins

	builder := squirrel.Select("recipient.username as to_user", "SUM(t.amount) AS amount").
		PlaceholderFormat(squirrel.Dollar).
		From("transactions t").
		Join("users recipient ON recipient.id = t.receiver_id").
		Where(squirrel.Eq{"t.sender_id": userId}).
		GroupBy("recipient.username")

	query, args, err := builder.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("SentCoinsByUserId: failed to build update SQL query")
		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.SentCoinsByUserId",
		QueryRow: query,
	}

	err = srp.db.DB().ScanAllContext(ctx, &sentCoins, queryStruct, args...)
	if err != nil {
		srp.log.Error().Err(err).Msg("SentCoinsByUserId: failed to scan rows")
		return nil, errresponse.ErrResponse(err)
	}

	return sentCoins, nil
}

func (srp *ShopRepo) ReceivedCoinsByUserId(ctx context.Context, userId int) ([]models.ReceivedCoins, error) {
	var receivedCoins []models.ReceivedCoins

	builder := squirrel.Select("sender.username AS from_user", "SUM(t.amount) AS amount").
		PlaceholderFormat(squirrel.Dollar).
		From("transactions t").
		Join("users sender ON sender.id = t.sender_id").
		Where(squirrel.Eq{"t.receiver_id": userId}).
		GroupBy("sender.username")

	query, args, err := builder.ToSql()
	if err != nil {
		srp.log.Error().Err(err).Msg("ReceivedCoinsByUserId: failed to build update SQL query")
		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.ReceivedCoinsByUserId",
		QueryRow: query,
	}

	err = srp.db.DB().ScanAllContext(ctx, &receivedCoins, queryStruct, args...)
	if err != nil {
		srp.log.Error().Err(err).Msg("ReceivedCoinsByUserId: failed to scan rows")
		return nil, errresponse.ErrResponse(err)
	}

	return receivedCoins, nil
}
