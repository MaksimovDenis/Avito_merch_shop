package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/gin-gonic/gin"
)

func GetAuthMiddlewareFunc(tokenMaker *token.JWTMaker) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/api/auth" {
			ctx.Next()
			return
		}

		claims, err := verifyClaimsFromAuthHeader(ctx, *tokenMaker)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx.Set("user", claims)
		ctx.Next()
	}
}

func verifyClaimsFromAuthHeader(ctx *gin.Context, tokenMaker token.JWTMaker) (*token.UserClaims, error) {
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("autorizatoin header is missing")
	}

	fields := strings.Fields(authHeader)
	if len(fields) != 2 || fields[0] != "Bearer" {
		return nil, fmt.Errorf("invalid autorization header")
	}

	token := fields[1]
	claims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
