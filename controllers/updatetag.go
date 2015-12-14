package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/techjanitor/pram-libs/audit"
	"github.com/techjanitor/pram-libs/auth"
	"github.com/techjanitor/pram-libs/config"
	e "github.com/techjanitor/pram-libs/errors"
	"github.com/techjanitor/pram-libs/perms"
	"github.com/techjanitor/pram-libs/redis"

	"github.com/techjanitor/pram-admin/models"
)

// update tag input
type updateTagForm struct {
	Id       uint   `json:"id" binding:"required"`
	Ib       uint   `json:"ib" binding:"required"`
	Tag      string `json:"name" binding:"required"`
	Type     uint   `json:"type" binding:"required"`
	Antispam string `json:"askey" binding:"required"`
}

// UpdateTagController will delete a tag
func UpdateTagController(c *gin.Context) {
	var err error
	var utf updateTagForm

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from session middleware
	userdata := c.MustGet("userdata").(auth.User)

	err = c.Bind(&utf)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err)
		return
	}

	// Set parameters to UpdateTagModel
	m := models.UpdateTagModel{
		Id:      utf.Id,
		Ib:      utf.Ib,
		Tag:     utf.Tag,
		TagType: utf.Type,
	}

	// Test for antispam key from Prim
	antispam := utf.Antispam
	if antispam != config.Settings.Antispam.AntispamKey {
		c.JSON(http.StatusBadRequest, gin.H{"error_message": e.ErrInvalidKey.Error()})
		c.Error(e.ErrInvalidKey)
		return
	}

	// check to see if user is allowed to perform action
	allowed, err := perms.Check(userdata.Id, m.Ib)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// if not allowed reject request
	if !allowed {
		c.JSON(e.ErrorMessage(e.ErrForbidden))
		c.Error(e.ErrForbidden)
		return
	}

	// Validate input parameters
	err = m.ValidateInput()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error_message": err.Error()})
		c.Error(err)
		return
	}

	// Check tag for duplicate
	err = m.Status()
	if err == e.ErrDuplicateTag {
		c.JSON(http.StatusBadRequest, gin.H{"error_message": err.Error()})
		c.Error(err)
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// Update data
	err = m.Update()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// Initialize cache handle
	cache := redis.RedisCache

	// Delete redis stuff
	tags_key := fmt.Sprintf("%s:%d", "tags", m.Ib)
	tag_key := fmt.Sprintf("%s:%d", "tag", m.Ib, m.Id)
	image_key := fmt.Sprintf("%s:%d", "image", m.Ib)

	err = cache.Delete(tags_key, tag_key, image_key)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err)
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditUpdateTag})

	// audit log
	audit := audit.Audit{
		User:   userdata.Id,
		Ib:     m.Ib,
		Ip:     c.ClientIP(),
		Action: audit.AuditUpdateTag,
		Info:   fmt.Sprintf("%s", m.Tag),
	}

	err = audit.Submit()
	if err != nil {
		c.Error(err)
	}

	return

}
