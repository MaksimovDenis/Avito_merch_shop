package handler

import (
	"net/http"

	"github.com/MaksimovDenis/Avito_merch_shop/internal/models"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/protocol/oapi"
	"github.com/gin-gonic/gin"
)

func (hdl *Handler) PostApiAuth(ctx *gin.Context) {
	var authReq oapi.AuthRequest

	if err := ctx.BindJSON(&authReq); err != nil {
		hdl.log.Error().Err(err).Msg("failed to parse request body")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})

		return
	}

	modelReq := &models.AuthReq{
		Username: authReq.Username,
		Password: authReq.Password,
	}

	token, err := hdl.appService.Authorization.Auth(ctx, modelReq)
	if err != nil {
		hdl.log.Error().Err(err).Msg("failed to auth user")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})

		return
	}

	res := &oapi.AuthResponse{
		Token: &token,
	}

	ctx.JSON(http.StatusOK, res)
}
