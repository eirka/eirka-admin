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

// CloseThreadController will toggle a threads Close bool
func CloseThreadController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(user.User)

	// Initialize model struct
	m := &models.CloseModel{
		Id: params[0],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err)
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// check if the user is authorized to perform this functions
	if !userdata.IsAuthorized(m.Ib) {
		c.JSON(e.ErrorMessage(e.ErrForbidden))
		c.Error(e.ErrForbidden)
		return
	}

	// toggle status
	err = m.Toggle()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
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
		c.Error(err)
		return
	}

	var success_message string

	if m.Closed {
		success_message = audit.AuditOpenThread
	} else {
		success_message = audit.AuditCloseThread
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

	err = audit.Submit()
	if err != nil {
		c.Error(err)
	}

	return

}
