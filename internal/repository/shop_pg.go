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
	GetItemsByUserId(ctx context.Context, userId int) (*[]models.Items, error)
	SentCoinsByUserId(ctx context.Context, userId int) (*[]models.SentCoins, error)
	ReceivedCoinsByUserId(ctx context.Context, userId int) (*[]models.ReceivedCoins, error)
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

		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UpdateBalanceForPurchase",
		QueryRow: query,
	}

	var productId int

	err = sr.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&productId)
	if err != nil {
		sr.log.Error().Err(err).Msg("UpdateBalanceForPurchase: failed to update user data")
		return nil, errresponse.ErrResponse(err, productName)
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
		return errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.InsertPurchaseRecord",
		QueryRow: query,
	}

	_, err = sr.db.DB().ExecContext(ctx, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("InsertPurchaseRecord: failed to insert purchase")
		return errresponse.ErrResponse(err)
	}

	return nil
}

func (sr *ShopRepo) UserBalanceByName(ctx context.Context, username string) (userId int, coins int, err error) {
	selectQueryBalance := squirrel.Select("id", "coins").
		PlaceholderFormat(squirrel.Dollar).
		From("users").
		Where(squirrel.Eq{"username": username})

	query, args, err := selectQueryBalance.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("UserBalanceByName: failed to build SQL query")
		return 0, 0, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UserBalanceByName",
		QueryRow: query,
	}

	err = sr.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&userId, &coins)
	if err != nil {
		sr.log.Error().Err(err).Msg("UserBalanceByName: failed to get sender balance")

		return 0, 0, errresponse.ErrResponse(err, username)
	}

	return userId, coins, nil
}

func (sr *ShopRepo) UpdateSenderBalance(ctx context.Context, sender string, amount int) (senderId int, err error) {
	updateQuerySender := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("coins - ?", amount)).
		Where(squirrel.Eq{"username": sender}).
		Suffix("RETURNING id")

	query, args, err := updateQuerySender.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("UpdateSenderBalance: failed to build update SQL query")

		return 0, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UpdateSenderBalance",
		QueryRow: query,
	}

	err = sr.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&senderId)
	if err != nil {
		sr.log.Error().Err(err).Msg("UpdateSenderBalance: failed to update sender balance")

		return 0, errresponse.ErrResponse(err, sender)
	}

	return senderId, nil
}

func (sr *ShopRepo) UpdateReceiverBalance(ctx context.Context, receiver string, amount int) (receiverId int, err error) {
	updateQuerySender := squirrel.Update("users").
		PlaceholderFormat(squirrel.Dollar).
		Set("coins", squirrel.Expr("coins + ?", amount)).
		Where(squirrel.Eq{"username": receiver}).
		Suffix("RETURNING id")

	query, args, err := updateQuerySender.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("UpdateReceiverBalance: failed to build update SQL query")

		return 0, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.UpdateReceiverBalance",
		QueryRow: query,
	}

	err = sr.db.DB().QueryRowContext(ctx, queryStruct, args...).Scan(&receiverId)
	if err != nil {
		sr.log.Error().Err(err).Msg("UpdateReceiverBalance: failed to update receiver balance")

		return 0, errresponse.ErrResponse(err, receiver)
	}

	return receiverId, nil
}

func (sr *ShopRepo) AddTransaction(ctx context.Context, senderId int, receiverId int, amount int) error {
	insertQueryTransact := squirrel.Insert("transactions").
		PlaceholderFormat(squirrel.Dollar).
		Columns("sender_id", "receiver_id", "amount").
		Values(senderId, receiverId, amount)

	query, args, err := insertQueryTransact.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("AddTransaction: failed to build update SQL query")
		return errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.AddTransaction",
		QueryRow: query,
	}

	_, err = sr.db.DB().ExecContext(ctx, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("AddTransaction: failed to add transaction")
		return errresponse.ErrResponse(err)
	}

	return nil
}

func (sr *ShopRepo) GetItemsByUserId(ctx context.Context, userId int) (*[]models.Items, error) {
	builder := squirrel.Select("p.name as name", "sum(pu.quantity) as quantity").
		PlaceholderFormat(squirrel.Dollar).
		From("purchases as pu").
		Join("products as p ON pu.products_id = p.id").
		Where(squirrel.Eq{"pu.user_id": userId}).
		GroupBy("name")

	query, args, err := builder.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("GetItemsByUserId: failed to build update SQL query")
		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.GetItemsByUserId",
		QueryRow: query,
	}

	var items []models.Items

	err = sr.db.DB().ScanAllContext(ctx, &items, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("GetItemsByUserId: failed to scan rows")
		return nil, errresponse.ErrResponse(err)
	}

	return &items, nil
}

func (sr *ShopRepo) SentCoinsByUserId(ctx context.Context, userId int) (*[]models.SentCoins, error) {
	builder := squirrel.Select("us.username as to_user", "sum(amount) as amount").
		PlaceholderFormat(squirrel.Dollar).
		From("users as us").
		Join("transactions t on us.id = t.sender_id").
		Where(squirrel.Eq{"us.id": userId}).
		GroupBy("to_user")

	query, args, err := builder.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("SentCoinsByUserId: failed to build update SQL query")
		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.SentCoinsByUserId",
		QueryRow: query,
	}

	var sendCoins []models.SentCoins

	err = sr.db.DB().ScanAllContext(ctx, &sendCoins, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("SentCoinsByUserId: failed to scan rows")
		return nil, errresponse.ErrResponse(err)
	}

	return &sendCoins, nil
}

func (sr *ShopRepo) ReceivedCoinsByUserId(ctx context.Context, userId int) (*[]models.ReceivedCoins, error) {
	builder := squirrel.Select("us.username as from_user", "sum(amount) as amount").
		PlaceholderFormat(squirrel.Dollar).
		From("users as us").
		Join("transactions t on us.id = t.receiver_id").
		Where(squirrel.Eq{"us.id": userId}).
		GroupBy("from_user")

	query, args, err := builder.ToSql()
	if err != nil {
		sr.log.Error().Err(err).Msg("ReceivedCoinsByUserId: failed to build update SQL query")
		return nil, errresponse.ErrResponse(err)
	}

	queryStruct := db.Query{
		Name:     "user_repository.ReceivedCoinsByUserId",
		QueryRow: query,
	}

	var receivedCoins []models.ReceivedCoins

	err = sr.db.DB().ScanAllContext(ctx, &receivedCoins, queryStruct, args...)
	if err != nil {
		sr.log.Error().Err(err).Msg("ReceivedCoinsByUserId: failed to scan rows")
		return nil, errresponse.ErrResponse(err)
	}

	return &receivedCoins, nil
}
