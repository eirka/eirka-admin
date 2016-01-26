package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/eirka/eirka-libs/audit"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/redis"
	"github.com/eirka/eirka-libs/user"

	"github.com/eirka/eirka-admin/models"
)

// StickyController will toggle a threads sticky bool
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
		Id: params[1],
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

	// Initialize cache handle
	cache := redis.RedisCache

	// Delete redis stuff
	index_key := fmt.Sprintf("%s:%d", "index", m.Ib)
	directory_key := fmt.Sprintf("%s:%d", "directory", m.Ib)
	thread_key := fmt.Sprintf("%s:%d:%d", "thread", m.Ib, m.Id)

	err = cache.Delete(index_key, directory_key, thread_key)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("StickyThreadController.cache.Delete")
		return
	}

	var success_message string

	// change the response message depending on the action
	if m.Sticky {
		success_message = audit.AuditUnstickyThread
	} else {
		success_message = audit.AuditStickyThread
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": success_message})

	// audit log
	audit := audit.Audit{
		User:   userdata.Id,
		Ib:     m.Ib,
		Ip:     c.ClientIP(),
		Action: success_message,
		Info:   fmt.Sprintf("%s", m.Name),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("StickyThreadController.audit.Submit")
	}

	return

}
