package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	docs "github.com/livspaceeng/ozone/docs"
	"github.com/livspaceeng/ozone/internal/controller"
	"github.com/livspaceeng/ozone/internal/services"
	"github.com/livspaceeng/ozone/internal/utils"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	httpClient                                      = &http.Client{}
	httpClientInterface        utils.HttpClient     = utils.NewHttpClient(httpClient)
	hydraService		services.HydraService		= services.NewHydraService(httpClient)
	ketoService			services.KetoService		= services.NewKetoService(httpClient)

	authController		controller.AuthController 	= controller.NewAuthController(hydraService, ketoService)
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	health := new(controller.HealthController)

	router.GET("/health", health.Status)
	//	router.Use(middlewares.AuthMiddleware())

	authResolver := router.Group("/api/v1/auth")
	{
		authResolver.GET("/check", authController.Check)
		authResolver.GET("/expand", authController.Expand)
		authResolver.GET("/relation_tuples", authController.Query)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return router
}
