package handler

import (
	"time"

	"github.com/MaksimovDenis/Avito_merch_shop/internal/service"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/protocol/oapi"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const (
	FileUploadBufferSize       = 512e+6 // 512MB for now
	ServerShutdownDefaultDelay = 5 * time.Second
)

type Handler struct {
	appService service.Service
	tokenMaker *token.JWTMaker
	log        zerolog.Logger
}

func NewHandler(appService service.Service, tokenMaker token.JWTMaker, log zerolog.Logger) *Handler {
	return &Handler{
		appService: appService,
		tokenMaker: &tokenMaker,
		log:        log,
	}
}

func (hdl *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.MaxMultipartMemory = FileUploadBufferSize

	tokenMaker := hdl.tokenMaker

	oapi.RegisterHandlersWithOptions(router, hdl, oapi.GinServerOptions{
		BaseURL: "/",
		Middlewares: []oapi.MiddlewareFunc{
			GetAuthMiddlewareFunc(tokenMaker),
		},
	})

	return router
}
