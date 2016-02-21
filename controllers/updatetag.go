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

// update tag input
type updateTagForm struct {
	Id   uint   `json:"id" binding:"required"`
	Tag  string `json:"name" binding:"required"`
	Type uint   `json:"type" binding:"required"`
}

// UpdateTagController will update a tags properties
func UpdateTagController(c *gin.Context) {
	var err error
	var utf updateTagForm

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("UpdateTagController.protected")
		return
	}

	err = c.Bind(&utf)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err).SetMeta("UpdateTagController.Bind")
		return
	}

	// Set parameters to UpdateTagModel
	m := models.UpdateTagModel{
		Ib:      params[0],
		Id:      utf.Id,
		Tag:     utf.Tag,
		TagType: utf.Type,
	}

	// Validate input parameters
	err = m.ValidateInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_message": err.Error()})
		c.Error(err).SetMeta("UpdateTagController.ValidateInput")
		return
	}

	// Check tag for duplicate
	err = m.Status()
	if err == e.ErrDuplicateTag {
		c.JSON(http.StatusBadRequest, gin.H{"error_message": err.Error()})
		c.Error(err).SetMeta("UpdateTagController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("UpdateTagController.Status")
		return
	}

	// Update data
	err = m.Update()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("UpdateTagController.Update")
		return
	}

	// Initialize cache handle
	cache := redis.RedisCache

	// Delete redis stuff
	tags_key := fmt.Sprintf("%s:%d", "tags", m.Ib)
	tag_key := fmt.Sprintf("%s:%d:%d", "tag", m.Ib, m.Id)
	image_key := fmt.Sprintf("%s:%d", "image", m.Ib)

	err = cache.Delete(tags_key, tag_key, image_key)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("UpdateTagController.cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditUpdateTag})

	// audit log
	audit := audit.Audit{
		User:   userdata.Id,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		Ip:     c.ClientIP(),
		Action: audit.AuditUpdateTag,
		Info:   fmt.Sprintf("%s", m.Tag),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("UpdateTagController.audit.Submit")
	}

	return

}
