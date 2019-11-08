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

// update tag input
type updateTagForm struct {
	ID   uint   `json:"id" binding:"required"`
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
		ID:      utf.ID,
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

	// Delete redis stuff
	tagsKey := fmt.Sprintf("%s:%d", "tags", m.Ib)
	tagKey := fmt.Sprintf("%s:%d:%d", "tag", m.Ib, m.ID)
	imageKey := fmt.Sprintf("%s:%d", "image", m.Ib)

	err = redis.Cache.Delete(tagsKey, tagKey, imageKey)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("UpdateTagController.redis.Cache.Delete")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditUpdateTag})

	// audit log
	audit := audit.Audit{
		User:   userdata.ID,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		IP:     c.ClientIP(),
		Action: audit.AuditUpdateTag,
		Info:   m.Tag,
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("UpdateTagController.audit.Submit")
	}

}
