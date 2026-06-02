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

// configure performs production runtime setup: the pid file, database and redis
// connection pools, runtime settings loaded from the database, and CORS domains.
// It runs from main (not an init func) so that test binaries — which import this
// package — do not try to write a pid file or open database/redis sockets.
func configure() {

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

// setupRouter builds the gin engine with every route and its middleware.
//
// Every functional endpoint is registered on the `admin` group, which enforces
// authentication (user.Auth) and moderator authorization (user.Protect). The only
// route intentionally left public is GET /status, a server-wide health check with no
// board context that the board-scoped Protect middleware cannot gate. That exception
// is the sole entry in the allowlist enforced by TestRouteSecurity in main_test.go;
// any new route added outside the `admin` group will fail that test.
func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.CORS())
	// verified the csrf token from the request
	r.Use(csrf.Verify())

	// public health check: the one intentional exception to the auth+mod rule.
	// Keep it in sync with the allowlist in main_test.go.
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

	return r
}

func main() {
	configure()

	r := setupRouter()

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
