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

// StickyThreadController will toggle a threads sticky bool
func StickyThreadController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("StickyThreadController.protected")
		return
	}

	// Initialize model struct
	m := &models.StickyModel{
		Ib: params[0],
		ID: params[1],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("StickyThreadController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("StickyThreadController.Status")
		return
	}

	// toggle status
	err = m.Toggle()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("StickyThreadController.Toggle")
		return
	}

	// Delete redis stuff
	indexKey := fmt.Sprintf("%s:%d", "index", m.Ib)
	directoryKey := fmt.Sprintf("%s:%d", "directory", m.Ib)
	threadKey := fmt.Sprintf("%s:%d:%d", "thread", m.Ib, m.ID)

	err = redis.Cache.Delete(indexKey, directoryKey, threadKey)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("StickyThreadController.redis.Cache.Delete")
		return
	}

	var successMessage string

	// change the response message depending on the action
	if m.Sticky {
		successMessage = audit.AuditUnstickyThread
	} else {
		successMessage = audit.AuditStickyThread
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
		c.Error(err).SetMeta("StickyThreadController.audit.Submit")
	}

	return

}
