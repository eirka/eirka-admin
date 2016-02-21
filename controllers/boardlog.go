package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"

	e "github.com/eirka/eirka-libs/errors"

	"github.com/eirka/eirka-admin/models"
)

// BoardLogController will get the the board audit log
func BoardLogController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("BoardLogController.protected")
		return
	}

	// Initialize model struct
	m := &models.BoardLogModel{
		Ib:   params[0],
		Page: params[1],
	}

	// Get the model which outputs JSON
	err := m.Get()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("BoardLogController.Get")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("BoardLogController.Get")
		return
	}

	// Marshal the structs into JSON
	output, err := json.Marshal(m.Result)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("BoardLogController.json.Marshal")
		return
	}

	c.Data(200, "application/json", output)

	return

}
