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

// BuyItem обрабатывает покупку товара пользователем.
// 1. Начинаем транзакцию в БД.
// 2. Обновляем баланс пользователя при покупке товара.
// 3. Записываем информацию о покупке в базу данных.
// 4. Фиксируем транзакцию или откатываем при ошибке.
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

// SendCoins выполняет перевод монет между пользователями.
// 1. Проверяем корректность суммы и что отправитель и получатель не совпадают.
// 2. Начинаем транзакцию в БД.
// 3. Проверяем баланс отправителя.
// 4. Обновляем баланс отправителя и получателя.
// 5. Добавлям запись о транзакции в базу данных.
// 6. Фиксируем транзакцию или откатывает при ошибке.
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

// Info предоставляет информацию о пользователе.
// 1. Получаем баланс пользователя.
// 2. Извлекаем список его покупок.
// 3. Получаем историю отправленных и полученных монет.
// 4. Возвращаем данные о пользователе.
func (svc *ShopService) Info(ctx context.Context, username string) (
	coins int,
	items []models.Items,
	sentCoins []models.SentCoins,
	receivedCoins []models.ReceivedCoins,
	err error,
) {
	userId, coins, err := svc.appRepository.Shop.UserBalanceByName(ctx, username)
	if err != nil {
		return 0, nil, nil, nil, err
	}

	items, err = svc.appRepository.Shop.GetItemsByUserId(ctx, userId)
	if err != nil {
		return 0, nil, nil, nil, err
	}

	sentCoins, err = svc.appRepository.Shop.SentCoinsByUserId(ctx, userId)
	if err != nil {
		return 0, nil, nil, nil, err
	}

	receivedCoins, err = svc.appRepository.Shop.ReceivedCoinsByUserId(ctx, userId)
	if err != nil {
		return 0, nil, nil, nil, err
	}

	return coins, items, sentCoins, receivedCoins, nil
}
