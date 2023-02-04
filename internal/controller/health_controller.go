package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController interface {
	Status(c *gin.Context)
}

type healthController struct{}

func NewHealthController() HealthController {
	return &healthController{}
}

// HealthController godoc
// @Summary      health check
// @Schemes      http
// @Description  check health
// @Tags         health
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string   true  "Bearer <Bouncer_access_token>" 
// @Success      200  {string}  OK!
// @Router       /health [get]
func (h healthController) Status(c *gin.Context) {
	c.String(http.StatusOK, "OK!")
}
