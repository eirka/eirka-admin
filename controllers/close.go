package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/audit"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/redis"
	"github.com/eirka/eirka-libs/user"

	"github.com/eirka/eirka-admin/models"
)

// CloseThreadController will toggle a threads close bool
func CloseThreadController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("CloseThreadController.protected")
		return
	}

	// Initialize model struct
	m := &models.CloseModel{
		Ib: params[0],
		ID: params[1],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("CloseThreadController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("CloseThreadController.Status")
		return
	}

	// toggle status
	err = m.Toggle()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("CloseThreadController.Toggle")
		return
	}

	// Initialize cache handle
	cache := redis.Cache

	// Delete redis stuff
	indexKey := fmt.Sprintf("%s:%d", "index", m.Ib)
	directoryKey := fmt.Sprintf("%s:%d", "directory", m.Ib)
	threadKey := fmt.Sprintf("%s:%d:%d", "thread", m.Ib, m.ID)

	err = cache.Delete(indexKey, directoryKey, threadKey)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("CloseThreadController.cache.Delete")
		return
	}

	var successMessage string

	// change response message depending on bool state
	if m.Closed {
		successMessage = audit.AuditOpenThread
	} else {
		successMessage = audit.AuditCloseThread
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": successMessage})

	// audit log
	audit := audit.Audit{
		User:   userdata.ID,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		IP:     c.ClientIP(),
		Action: successMessage,
		Info:   fmt.Sprintf("%s", m.Name),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("CloseThreadController.audit.Submit")
	}

	return

}
