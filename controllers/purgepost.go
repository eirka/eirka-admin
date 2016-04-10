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

// PurgePostController will remove a deleted posts files and rows
func PurgePostController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("PurgePostController.protected")
		return
	}

	// Initialize model struct
	m := &models.PurgePostModel{
		Ib:     params[0],
		Thread: params[1],
		ID:     params[2],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("PurgePostController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgePostController.Status")
		return
	}

	// Delete data
	err = m.Delete()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgePostController.Delete")
		return
	}

	// Initialize cache handle
	cache := redis.Cache

	// Delete redis stuff
	indexKey := fmt.Sprintf("%s:%d", "index", m.Ib)
	directoryKey := fmt.Sprintf("%s:%d", "directory", m.Ib)
	threadKey := fmt.Sprintf("%s:%d:%d", "thread", m.Ib, m.Thread)
	postKey := fmt.Sprintf("%s:%d:%d", "post", m.Ib, m.Thread)
	tagsKey := fmt.Sprintf("%s:%d", "tags", m.Ib)
	imageKey := fmt.Sprintf("%s:%d", "image", m.Ib)
	newKey := fmt.Sprintf("%s:%d", "new", m.Ib)
	popularKey := fmt.Sprintf("%s:%d", "popular", m.Ib)
	favoritedKey := fmt.Sprintf("%s:%d", "favorited", m.Ib)

	err = cache.Delete(indexKey, directoryKey, threadKey, postKey, tagsKey, imageKey, newKey, popularKey, favoritedKey)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("PurgePostController.cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditPurgePost})

	// audit log
	audit := audit.Audit{
		User:   userdata.ID,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		IP:     c.ClientIP(),
		Action: audit.AuditPurgePost,
		Info:   fmt.Sprintf("%s/%d", m.Name, m.ID),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("PurgePostController.audit.Submit")
	}

	return

}
