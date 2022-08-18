package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

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
func (h HealthController) Status(c *gin.Context) {
	c.String(http.StatusOK, "OK!")
}
