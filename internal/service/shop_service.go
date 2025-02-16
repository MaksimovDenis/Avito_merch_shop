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
	SendCoins(ctx context.Context, sender string, receiver string, amount int) error
	Info(ctx context.Context, username string) (
		coins int,
		items []models.Items,
		sentCoins []models.SentCoins,
		receivedCoins []models.ReceivedCoins,
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

func (svc *ShopService) BuyItem(ctx context.Context, userId int, productName string) error {
	tx, err := svc.client.DB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		svc.log.Error().Err(err).Msg("failed to start transaction")
		return err
	}

	ctx = pg.MakeContextTx(ctx, tx)

	productId, err := svc.appRepository.Shop.UpdateBalanceForPurchase(ctx, userId, productName)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := svc.appRepository.Shop.InsertPurchaseRecord(ctx, userId, *productId); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (svc *ShopService) SendCoins(ctx context.Context, sender string, receiver string, amount int) error {
	if amount <= 0 {
		return errors.New("сумма перевода должна быть положительным числом")
	}

	if sender == receiver {
		return errors.New("имя отправителя совпадает с именем получателя")
	}

	tx, err := svc.client.DB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		svc.log.Error().Err(err).Msg("failed to start transaction")
		return err
	}

	ctx = pg.MakeContextTx(ctx, tx)

	_, senderBalance, err := svc.appRepository.Shop.UserBalanceByName(ctx, sender)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if senderBalance < amount {
		svc.log.Error().Err(err).Msg("not enough coins for transaction")

		_ = tx.Rollback(ctx)

		return errors.New("недостаточно монет для перевода")
	}

	senderId, err := svc.appRepository.Shop.UpdateSenderBalance(ctx, sender, amount)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	receiverId, err := svc.appRepository.Shop.UpdateReceiverBalance(ctx, receiver, amount)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err = svc.appRepository.Shop.AddTransaction(ctx, senderId, receiverId, amount); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (svc *ShopService) Info(ctx context.Context, username string) (
	coins int,
	items []models.Items,
	sentCoins []models.SentCoins,
	receivedCoins []models.ReceivedCoins,
	err error,
) {
	tx, err := svc.client.DB().BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		svc.log.Error().Err(err).Msg("failed to start transaction")
		return 0, nil, nil, nil, err
	}

	ctx = pg.MakeContextTx(ctx, tx)

	userId, coins, err := svc.appRepository.Shop.UserBalanceByName(ctx, username)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	items, err = svc.appRepository.Shop.GetItemsByUserId(ctx, userId)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	sentCoins, err = svc.appRepository.Shop.SentCoinsByUserId(ctx, userId)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	receivedCoins, err = svc.appRepository.Shop.ReceivedCoinsByUserId(ctx, userId)
	if err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return 0, nil, nil, nil, err
	}

	return coins, items, sentCoins, receivedCoins, nil
}
