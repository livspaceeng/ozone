package server

import (
	"github.com/gin-gonic/gin"
	docs "github.com/livspaceeng/ozone/docs"
	"github.com/livspaceeng/ozone/internal/controller"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	health := new(controller.HealthController)

	router.GET("/health", health.Status)
	//	router.Use(middlewares.AuthMiddleware())

	authResolver := router.Group("/api/v1/auth")
	{
		authController := new(controller.AuthController)
		authResolver.GET("/check", authController.Check)
		authResolver.GET("/expand", authController.Expand)
		authResolver.GET("/relation_tuples", authController.Query)
		authResolver.PUT("/relation_tuples", authController.Create)
		authResolver.DELETE("/relation_tuples", authController.Delete)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return router
}
