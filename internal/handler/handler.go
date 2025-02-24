package handler

import (
	"time"

	"github.com/MaksimovDenis/Avito_merch_shop/internal/metrics"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/service"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/protocol/oapi"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	metrics    *metrics.Metrics
}

func NewHandler(
	appService service.Service,
	tokenMaker token.JWTMaker,
	log zerolog.Logger,
	metrics *metrics.Metrics) *Handler {
	return &Handler{
		appService: appService,
		tokenMaker: &tokenMaker,
		log:        log,
		metrics:    metrics,
	}
}

func (hdl *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.MaxMultipartMemory = FileUploadBufferSize

	tokenMaker := hdl.tokenMaker

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.Use(hdl.metrics.HTTPMetrics())
	oapi.RegisterHandlersWithOptions(router, hdl, oapi.GinServerOptions{
		BaseURL: "/",
		Middlewares: []oapi.MiddlewareFunc{
			GetAuthMiddlewareFunc(tokenMaker),
		},
	})

	return router
}
