package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/audit"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/user"

	"github.com/eirka/eirka-admin/models"
)

// ban file input
type banFileForm struct {
	Reason string `json:"reason" binding:"required"`
}

// BanFileController will ban an image file hash
func BanFileController(c *gin.Context) {
	var err error
	var bff banFileForm

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("BanFileController.protected")
		return
	}

	err = c.Bind(&bff)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err).SetMeta("BanFileController.Bind")
		return
	}

	// Initialize model struct
	m := &models.BanFileModel{
		Ib:     params[0],
		Thread: params[1],
		ID:     params[2],
		User:   userdata.ID,
		Reason: bff.Reason,
	}

	// Check the record id and get further info
	err = m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("BanFileController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("BanFileController.Status")
		return
	}

	// add ban to database
	err = m.Post()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("BanFileController.Post")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditBanFile})

	// audit log
	audit := audit.Audit{
		User:   userdata.ID,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		IP:     c.ClientIP(),
		Action: audit.AuditBanFile,
		Info:   m.Reason,
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("BanFileController.audit.Submit")
	}

}
