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

// DeletePostController will mark a post as deleted in the database
func DeletePostController(c *gin.Context) {

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	// Initialize model struct
	m := &models.DeletePostModel{
		Thread: params[0],
		Id:     params[1],
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

	// check if the user is authorized to perform this functions
	if !userdata.IsAuthorized(m.Ib) {
		c.JSON(e.ErrorMessage(e.ErrForbidden))
		c.Error(e.ErrForbidden).SetMeta("DeletePostController.userdata.IsAuthorized")
		return
	}

	// Delete data
	err = m.Delete()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeletePostController.Delete")
		return
	}

	// Initialize cache handle
	cache := redis.RedisCache

	// Delete redis stuff
	index_key := fmt.Sprintf("%s:%d", "index", m.Ib)
	directory_key := fmt.Sprintf("%s:%d", "directory", m.Ib)
	thread_key := fmt.Sprintf("%s:%d:%d", "thread", m.Ib, m.Thread)
	post_key := fmt.Sprintf("%s:%d:%d", "post", m.Ib, m.Thread)
	tags_key := fmt.Sprintf("%s:%d", "tags", m.Ib)
	image_key := fmt.Sprintf("%s:%d", "image", m.Ib)
	new_key := fmt.Sprintf("%s:%d", "new", m.Ib)
	popular_key := fmt.Sprintf("%s:%d", "popular", m.Ib)
	favorited_key := fmt.Sprintf("%s:%d", "favorited", m.Ib)

	err = cache.Delete(index_key, directory_key, thread_key, post_key, tags_key, image_key, new_key, popular_key, favorited_key)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("DeletePostController.cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditDeletePost})

	// audit log
	audit := audit.Audit{
		User:   userdata.Id,
		Ib:     m.Ib,
		Ip:     c.ClientIP(),
		Action: audit.AuditDeletePost,
		Info:   fmt.Sprintf("%s/%d", m.Name, m.Id),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("DeletePostController.audit.Submit")
	}

	return

}
