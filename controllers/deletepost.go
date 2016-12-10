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

// DeletePostController will mark a post as deleted in the database
func DeletePostController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("DeletePostController.protected")
		return
	}

	// Initialize model struct
	m := &models.DeletePostModel{
		Ib:     params[0],
		Thread: params[1],
		ID:     params[2],
	}

	// Check the record id and get further info
	err := m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("DeletePostController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeletePostController.Status")
		return
	}

	// Delete data
	err = m.Delete()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeletePostController.Delete")
		return
	}

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

	err = redis.Cache.Delete(indexKey, directoryKey, threadKey, postKey, tagsKey, imageKey, newKey, popularKey, favoritedKey)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeletePostController.redis.Cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditDeletePost})

	// audit log
	audit := audit.Audit{
		User:   userdata.ID,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		IP:     c.ClientIP(),
		Action: audit.AuditDeletePost,
		Info:   fmt.Sprintf("%s/%d", m.Name, m.ID),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("DeletePostController.audit.Submit")
	}

	return

}
