package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"

	"github.com/eirka/eirka-admin/models"
)

// StatisticsController will get the visitor stats for a board
func StatisticsController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("StatisticsController.protected")
		return
	}

	// Initialize model struct
	m := &models.StatisticsModel{
		Ib: params[0],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("StatisticsController.Get")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("StatisticsController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("StatisticsController.json.Marshal")
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Write(output)

	return

}
