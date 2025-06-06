package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/facebookgo/pidfile"
	"github.com/gin-gonic/gin"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/cors"
	"github.com/eirka/eirka-libs/csrf"
	"github.com/eirka/eirka-libs/db"
	"github.com/eirka/eirka-libs/redis"
	"github.com/eirka/eirka-libs/status"
	"github.com/eirka/eirka-libs/user"
	"github.com/eirka/eirka-libs/validate"

	local "github.com/eirka/eirka-admin/config"
	c "github.com/eirka/eirka-admin/controllers"
)

func init() {

	// create pid file
	pidfile.SetPidfilePath("/run/eirka/eirka-admin.pid")

	err := pidfile.Write()
	if err != nil {
		panic("Could not write pid file")
	}

	// Database connection settings
	dbase := db.Database{
		User:           local.Settings.Database.User,
		Password:       local.Settings.Database.Password,
		Proto:          local.Settings.Database.Protocol,
		Host:           local.Settings.Database.Host,
		Database:       local.Settings.Database.Database,
		MaxIdle:        local.Settings.Admin.DatabaseMaxIdle,
		MaxConnections: local.Settings.Admin.DatabaseMaxConnections,
	}

	// Set up DB connection
	dbase.NewDb()

	// Get limits and stuff from database
	config.GetDatabaseSettings()

	// redis settings
	r := redis.Redis{
		// Redis address and max pool connections
		Protocol:       local.Settings.Redis.Protocol,
		Address:        local.Settings.Redis.Host,
		MaxIdle:        local.Settings.Admin.RedisMaxIdle,
		MaxConnections: local.Settings.Admin.RedisMaxConnections,
	}

	// Set up Redis connection
	r.NewRedisCache()

	// set cors domains
	cors.SetDomains(local.Settings.CORS.Sites, strings.Split("GET,POST,DELETE", ","))

}

func main() {
	r := gin.Default()

	r.Use(cors.CORS())
	// verified the csrf token from the request
	r.Use(csrf.Verify())

	r.GET("/status", status.StatusController)
	r.NoRoute(c.ErrorController)

	// requires mod perms
	admin := r.Group("/")

	admin.Use(validate.ValidateParams())
	admin.Use(user.Auth(true))
	admin.Use(user.Protect())

	admin.GET("/statistics/:ib", c.StatisticsController)
	admin.GET("/log/board/:ib/:page", c.BoardLogController)
	admin.GET("/log/mod/:ib/:page", c.ModLogController)

	admin.DELETE("/tag/:ib/:id", c.DeleteTagController)
	admin.DELETE("/imagetag/:ib/:image/:tag", c.DeleteImageTagController)
	admin.DELETE("/thread/:ib/:id", c.DeleteThreadController)
	admin.DELETE("/post/:ib/:thread/:id", c.DeletePostController)

	admin.POST("/tag/:ib", c.UpdateTagController)
	admin.POST("/sticky/:ib/:thread", c.StickyThreadController)
	admin.POST("/close/:ib/:thread", c.CloseThreadController)
	admin.POST("/ban/ip/:ib/:thread/:post", c.BanIPController)
	admin.POST("/ban/file/:ib/:thread/:post", c.BanFileController)
	admin.POST("/user/resetpassword/:ib", c.ResetPasswordController)

	s := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", local.Settings.Admin.Host, local.Settings.Admin.Port),
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           r,
	}

	err := gracehttp.Serve(s)
	if err != nil {
		panic("Could not start server")
	}

}
