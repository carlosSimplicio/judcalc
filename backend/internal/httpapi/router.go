package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
	"github.com/carlosSimplicio/judcalc/backend/internal/httpapi/handlers"
)

func NewRouter(areas domain.AreaRepository, services domain.ServiceRepository, fixedCosts domain.FixedCostsRepository, authentication AuthService, logger *slog.Logger) *gin.Engine {
	if logger == nil {
		logger = slog.Default()
	}
	router := gin.New()
	router.Use(requestLogger(logger))
	router.Use(gin.CustomRecovery(func(ctx *gin.Context, recovered any) {
		logger.ErrorContext(ctx.Request.Context(), "pânico durante requisição", "recovered", recovered)
		handlers.WriteError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
	}))

	areaHandler := handlers.NewArea(areas, logger)
	serviceHandler := handlers.NewService(services, logger)
	fixedCostsHandler := handlers.NewFixedCosts(fixedCosts, logger)
	authenticationHandler := handlers.NewAuthentication(authentication, logger)
	api := router.Group("/api/v1")
	api.POST("/auth/sign-up", authenticationHandler.SignUp)
	api.POST("/auth/sign-in", authenticationHandler.SignIn)
	protected := api.Group("")
	protected.Use(requireAuthentication(authentication, logger))
	protected.GET("/areas", areaHandler.List)
	protected.GET("/services", serviceHandler.List)
	protected.GET("/fixed-costs", fixedCostsHandler.Get)
	protected.PATCH("/fixed-costs", fixedCostsHandler.Patch)

	router.NoRoute(func(ctx *gin.Context) {
		handlers.WriteError(ctx, http.StatusNotFound, "not_found", "Rota não encontrada.")
	})
	return router
}
