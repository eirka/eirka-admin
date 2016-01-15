package main

import (
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/cors"
	"github.com/eirka/eirka-libs/csrf"
	"github.com/eirka/eirka-libs/db"
	"github.com/eirka/eirka-libs/redis"
	"github.com/eirka/eirka-libs/user"
	"github.com/eirka/eirka-libs/validate"

	local "github.com/eirka/eirka-admin/config"
	c "github.com/eirka/eirka-admin/controllers"
	u "github.com/eirka/eirka-admin/utils"
)

var (
	version = "0.0.1"
)

func init() {

	// Database connection settings
	dbase := db.Database{

		User:           local.Settings.Database.User,
		Password:       local.Settings.Database.Password,
		Proto:          local.Settings.Database.Proto,
		Host:           local.Settings.Database.Host,
		Database:       local.Settings.Database.Database,
		MaxIdle:        local.Settings.Database.MaxIdle,
		MaxConnections: local.Settings.Database.MaxConnections,
	}

	// Set up DB connection
	dbase.NewDb()

	// Get limits and stuff from database
	config.GetDatabaseSettings()

	// redis settings
	r := redis.Redis{
		// Redis address and max pool connections
		Protocol:       local.Settings.Redis.Protocol,
		Address:        local.Settings.Redis.Address,
		MaxIdle:        local.Settings.Redis.MaxIdle,
		MaxConnections: local.Settings.Redis.MaxConnections,
	}

	// Set up Redis connection
	r.NewRedisCache()

	// set auth middleware secret
	user.Secret = local.Settings.Session.Secret

	// print the starting info
	StartInfo()

	// Print out config
	config.Print()

	// Print out config
	local.Print()

	// check what services are available
	u.CheckServices()

	// Print capabilities
	u.Services.Print()

	// set cors domains
	cors.SetDomains(local.Settings.CORS.Sites, strings.Split("GET,POST,DELETE", ","))

}

func main() {
	r := gin.Default()

	r.Use(cors.CORS())
	// verified the csrf token from the request
	r.Use(csrf.Verify())

	r.GET("/uptime", c.UptimeController)
	r.NoRoute(c.ErrorController)

	// requires mod perms
	admin := r.Group("/")

	admin.Use(validate.ValidateParams())
	admin.Use(user.Auth(true))

	admin.GET("/statistics/:ib", c.StatisticsController)
	admin.DELETE("/tag/:id", c.DeleteTagController)
	admin.POST("/tag", c.UpdateTagController)
	admin.DELETE("/imagetag/:image/:tag", c.DeleteImageTagController)
	admin.DELETE("/thread/:id", c.DeleteThreadController)
	admin.DELETE("/post/:thread/:id", c.DeletePostController)
	admin.POST("/sticky/:thread", c.StickyThreadController)
	admin.POST("/close/:thread", c.CloseThreadController)

	//admin.DELETE("/thread/:id", c.PurgeThreadController)
	//admin.DELETE("/post/:thread/:id", c.PurgePostController)
	//admin.POST("/ban/:ip", c.BanIpController)
	//admin.DELETE("/flushcache", c.DeleteCacheController)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", local.Settings.Admin.Address, local.Settings.Admin.Port),
		Handler: r,
	}

	gracehttp.Serve(s)

}

func StartInfo() {

	fmt.Println(strings.Repeat("*", 60))
	fmt.Printf("%-20v\n\n", "PRAM-ADMIN")
	fmt.Printf("%-20v%40v\n", "Version", version)
	fmt.Println(strings.Repeat("*", 60))

}
