package handler

import (
	"net/http"

	"github.com/MaksimovDenis/Avito_merch_shop/pkg/protocol/oapi"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/gin-gonic/gin"
)

func (hdl *Handler) GetApiBuyItem(ctx *gin.Context, productName string) {
	claims, ok := ctx.Get("user")
	if !ok {
		hdl.log.Error().Msg("user claims not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})

		return
	}

	userId := claims.(*token.UserClaims).ID

	if err := hdl.appService.Shop.BuyItem(ctx, int(userId), productName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Товар приобретён"})
}

func (hdl *Handler) PostApiSendCoin(ctx *gin.Context) {
	var sendCoinsReq oapi.SendCoinRequest

	if err := ctx.BindJSON(&sendCoinsReq); err != nil {
		hdl.log.Error().Err(err).Msg("failed to parse request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})

		return
	}

	claims, ok := ctx.Get("user")
	if !ok {
		hdl.log.Error().Msg("user claims not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})

		return
	}

	sender := claims.(*token.UserClaims).UserName

	if err := hdl.appService.Shop.SendCoin(ctx, sender, sendCoinsReq.ToUser, sendCoinsReq.Amount); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Перевод выполнен"})
}

func (hdl *Handler) GetApiInfo(ctx *gin.Context) {
	claims, ok := ctx.Get("user")
	if !ok {
		hdl.log.Error().Msg("user claims not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	username := claims.(*token.UserClaims).UserName

	coins, items, sentCoins, receivedCoins, err := hdl.appService.Shop.Info(ctx, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Преобразуем Items в Inventory
	var inventory []struct {
		Quantity *int    `json:"quantity,omitempty"`
		Type     *string `json:"type,omitempty"`
	}

	if items != nil {
		for _, item := range *items {
			itemType := item.Name // В `InfoResponse` используется `Type` вместо `Name`
			quantity := item.Quantity
			inventory = append(inventory, struct {
				Quantity *int    `json:"quantity,omitempty"`
				Type     *string `json:"type,omitempty"`
			}{
				Quantity: &quantity,
				Type:     &itemType,
			})
		}
	}

	// Преобразуем ReceivedCoins
	var received []struct {
		Amount   *int    `json:"amount,omitempty"`
		FromUser *string `json:"fromUser,omitempty"`
	}

	if receivedCoins != nil {
		for _, rc := range *receivedCoins {
			amount := rc.Amount
			fromUser := rc.FromUser
			received = append(received, struct {
				Amount   *int    `json:"amount,omitempty"`
				FromUser *string `json:"fromUser,omitempty"`
			}{
				Amount:   &amount,
				FromUser: &fromUser,
			})
		}
	}

	// Преобразуем SentCoins
	var sent []struct {
		Amount *int    `json:"amount,omitempty"`
		ToUser *string `json:"toUser,omitempty"`
	}

	if sentCoins != nil {
		for _, sc := range *sentCoins {
			amount := sc.Amount
			toUser := sc.ToUser
			sent = append(sent, struct {
				Amount *int    `json:"amount,omitempty"`
				ToUser *string `json:"toUser,omitempty"`
			}{
				Amount: &amount,
				ToUser: &toUser,
			})
		}
	}

	// Формируем ответ
	res := oapi.InfoResponse{
		Coins:     &coins,
		Inventory: &inventory,
		CoinHistory: &struct {
			Received *[]struct {
				Amount   *int    `json:"amount,omitempty"`
				FromUser *string `json:"fromUser,omitempty"`
			} `json:"received,omitempty"`
			Sent *[]struct {
				Amount *int    `json:"amount,omitempty"`
				ToUser *string `json:"toUser,omitempty"`
			} `json:"sent,omitempty"`
		}{
			Received: &received,
			Sent:     &sent,
		},
	}

	ctx.JSON(http.StatusOK, res)
}
