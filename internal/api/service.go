package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (e *Endpoints) DeleteCheck(c *gin.Context) {
	checkId := c.Param("id")
	if len(checkId) == 0 {
		c.JSON(http.StatusBadRequest, struct {
			Message string
		}{Message: "Id must not be empty"})
	}

	err := e.UptimeProvider.DeleteCheck(context.Background(), checkId)

	if err == nil {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	c.JSON(http.StatusInternalServerError, err)
	return
}
