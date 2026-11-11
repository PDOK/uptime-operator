package api

import (
	"net/http"

	"github.com/PDOK/uptime-operator/internal/service"
	"github.com/gin-gonic/gin"
)

func AddRoutes(router *gin.Engine, e Endpoints) {
	v1 := router.Group("/api/v1")
	{
		check := v1.Group("/check")
		{
			check.DELETE("/:id", e.DeleteCheck)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, struct {
		}{})
	})
}

type Endpoints struct {
	UptimeProvider service.UptimeProvider
}
