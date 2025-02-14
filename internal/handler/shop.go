package handler

import (
	"net/http"

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
