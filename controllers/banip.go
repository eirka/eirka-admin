package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/audit"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/user"

	"github.com/eirka/eirka-admin/models"
	u "github.com/eirka/eirka-admin/utils"
)

// ban Ip input
type banIPForm struct {
	Reason string `json:"reason" binding:"required"`
}

// BanIPController will ban an ip
func BanIPController(c *gin.Context) {
	var err error
	var bif banIPForm

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("BanIpController.protected")
		return
	}

	err = c.Bind(&bif)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err).SetMeta("BanIpController.Bind")
		return
	}

	// Initialize model struct
	m := &models.BanIPModel{
		Ib:     params[0],
		Thread: params[1],
		ID:     params[2],
		User:   userdata.ID,
		Reason: bif.Reason,
	}

	// Check the record id and get further info
	err = m.Status()
	if err == e.ErrNotFound {
		c.JSON(e.ErrorMessage(e.ErrNotFound))
		c.Error(err).SetMeta("BanIpController.Status")
		return
	} else if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("BanIpController.Status")
		return
	}

	// add ban to database
	err = m.Post()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("BanIpController.Post")
		return
	}

	// ban the ip in cloudflare
	go u.CloudFlareBanIP(m.IP, m.Reason)

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditBanIP})

	// audit log
	audit := audit.Audit{
		User:   userdata.ID,
		Ib:     m.Ib,
		Type:   audit.ModLog,
		IP:     c.ClientIP(),
		Action: audit.AuditBanIP,
		Info:   m.Reason,
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("BanIpController.audit.Submit")
	}

}
