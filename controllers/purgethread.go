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

// PurgeThreadController will remove deleted files and rows
func PurgeThreadController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("PurgeThreadController.protected")
		return
	}

	// Initialize model struct
	m := &models.PurgeThreadModel{
		Ib: params[0],
		ID: params[1],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("PurgeThreadController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgeThreadController.Status")
		return
	}

	// Delete data
	err = m.Delete()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgeThreadController.Delete")
		return
	}

	// Delete redis stuff
	indexKey := fmt.Sprintf("%s:%d", "index", m.Ib)
	directoryKey := fmt.Sprintf("%s:%d", "directory", m.Ib)
	threadKey := fmt.Sprintf("%s:%d:%d", "thread", m.Ib, m.ID)
	postKey := fmt.Sprintf("%s:%d:%d", "post", m.Ib, m.ID)
	tagsKey := fmt.Sprintf("%s:%d", "tags", m.Ib)
	imageKey := fmt.Sprintf("%s:%d", "image", m.Ib)
	newKey := fmt.Sprintf("%s:%d", "new", m.Ib)
	popularKey := fmt.Sprintf("%s:%d", "popular", m.Ib)
	favoritedKey := fmt.Sprintf("%s:%d", "favorited", m.Ib)

	err = redis.Cache.Delete(indexKey, directoryKey, threadKey, postKey, tagsKey, imageKey, newKey, popularKey, favoritedKey)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgeThreadController.redis.Cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditPurgeThread})

	// audit log
	audit := audit.Audit{
		User:   userdata.ID,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		IP:     c.ClientIP(),
		Action: audit.AuditPurgeThread,
		Info:   m.Name,
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("PurgeThreadController.audit.Submit")
	}

}
