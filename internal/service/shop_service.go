package service

import (
	"context"
	"errors"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/client/db/pg"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/models"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/repository"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"
)

type Shop interface {
	BuyItem(ctx context.Context, userId int, productName string) error
	SendCoin(ctx context.Context, sender string, receiver string, amount int) error
	Info(ctx context.Context, username string) (
		coins int,
		items *[]models.Items,
		sentCoins *[]models.SentCoins,
		receivedCoins *[]models.ReceivedCoins,
		err error,
	)
}

type ShopService struct {
	appRepository repository.Repository
	client        db.Client
	log           zerolog.Logger
}

func newShopService(
	appRepository repository.Repository,
	client db.Client,
	log zerolog.Logger,
) *ShopService {
	return &ShopService{
		appRepository: appRepository,
		client:        client,
		log:           log,
	}
}

func (ss *ShopService) BuyItem(ctx context.Context, userId int, productName string) error {
	tx, err := ss.client.DB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		ss.log.Error().Err(err).Msg("failed to start transaction")
		return err
	}

	ctx = pg.MakeContextTx(ctx, tx)

	productId, err := ss.appRepository.Shop.UpdateBalanceForPurchase(ctx, userId, productName)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := ss.appRepository.Shop.InsertPurchaseRecord(ctx, userId, *productId); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (ss *ShopService) SendCoin(ctx context.Context, sender string, receiver string, amount int) error {
	if amount <= 0 {
		return errors.New("сумма перевода должна быть положительным числом")
	}

	if sender == receiver {
		return errors.New("имя отправителя совпадает с именем получателя")
	}

	tx, err := ss.client.DB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		ss.log.Error().Err(err).Msg("failed to start transaction")
		return err
	}

	ctx = pg.MakeContextTx(ctx, tx)

	_, senderBalance, err := ss.appRepository.Shop.UserBalanceByName(ctx, sender)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if senderBalance < amount {
		ss.log.Error().Err(err).Msg("not enough coins for transaction")
		_ = tx.Rollback(ctx)

		return errors.New("недостаточно монет для перевода")
	}

	senderId, err := ss.appRepository.Shop.UpdateSenderBalance(ctx, sender, amount)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	receiverId, err := ss.appRepository.Shop.UpdateReceiverBalance(ctx, receiver, amount)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err = ss.appRepository.Shop.AddTransaction(ctx, senderId, receiverId, amount); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (ss *ShopService) Info(ctx context.Context, username string) (
	coins int,
	items *[]models.Items,
	sentCoins *[]models.SentCoins,
	receivedCoins *[]models.ReceivedCoins,
	err error,
) {
	tx, err := ss.client.DB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		ss.log.Error().Err(err).Msg("failed to start transaction")
		return 0, nil, nil, nil, err
	}

	ctx = pg.MakeContextTx(ctx, tx)

	userId, coins, err := ss.appRepository.Shop.UserBalanceByName(ctx, username)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	items, err = ss.appRepository.Shop.GetItemsByUserId(ctx, userId)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	sentCoins, err = ss.appRepository.Shop.SentCoinsByUserId(ctx, userId)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	receivedCoins, err = ss.appRepository.Shop.ReceivedCoinsByUserId(ctx, userId)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	tx.Commit(ctx)

	return coins, items, sentCoins, receivedCoins, nil
}
