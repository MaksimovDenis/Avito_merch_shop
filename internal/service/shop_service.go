package service

import (
	"context"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/client/db/pg"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/repository"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"
)

type Shop interface {
	BuyItem(ctx context.Context, userId int, productName string) error
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
		ss.log.Error().Err(err).Msg("BuyItem: failed to start transaction")
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

	return nil
}

func (ss *ShopService) SendCoin(ctx context.Context, sender string, receiver string, amount int) error {
	if err := ss.appRepository.Shop.SendCoin(ctx, sender, receiver, amount); err != nil {
		return err
	}

	return nil
}
