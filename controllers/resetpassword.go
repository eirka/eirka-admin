package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/eirka/eirka-libs/audit"
	e "github.com/eirka/eirka-libs/errors"
	"github.com/eirka/eirka-libs/user"

	u "github.com/eirka/eirka-admin/utils"
)

// reset password input
type resetPasswordForm struct {
	Uid uint `json:"uid" binding:"required"`
}

// ResetPasswordController will reset an ip
func ResetPasswordController(c *gin.Context) {
	var err error
	var rpf ResetPasswordForm

	// Get parameters from validate middleware
	params := c.MustGet("params").([]uint)

	// get userdata from user middleware
	userdata := c.MustGet("userdata").(user.User)

	if !c.MustGet("protected").(bool) {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(e.ErrInternalError).SetMeta("ResetPasswordController.protected")
		return
	}

	err = c.Bind(&rpf)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInvalidParam))
		c.Error(err).SetMeta("ResetPasswordController.Bind")
		return
	}

	// generate a random password
	password, hash, err := user.RandomPassword()
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ResetPasswordController.RandomPassword")
		return
	}

	// update the password in the database
	err = user.UpdatePassword(hash, rpf.Uid)
	if err != nil {
		c.JSON(e.ErrorMessage(e.ErrInternalError))
		c.Error(err).SetMeta("ResetPasswordController.UpdatePassword")
		return
	}

	// response message
	c.JSON(http.StatusOK, gin.H{"success_message": audit.AuditResetPassword, "password": password})

	// audit log
	audit := audit.Audit{
		User:   userdata.Id,
		Ib:     params[0],
		Type:   audit.UserLog,
		Ip:     c.ClientIP(),
		Action: audit.AuditResetPassword,
		Info:   fmt.Sprintf("%s", rpf.Uid),
	}

	// submit audit
	err = audit.Submit()
	if err != nil {
		c.Error(err).SetMeta("ResetPasswordController.audit.Submit")
	}

	return

}
