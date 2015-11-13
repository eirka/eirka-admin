package main

import (
	"fmt"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"

	"github.com/techjanitor/pram-libs/auth"
	"github.com/techjanitor/pram-libs/config"
	"github.com/techjanitor/pram-libs/cors"
	"github.com/techjanitor/pram-libs/db"
	"github.com/techjanitor/pram-libs/redis"
	"github.com/techjanitor/pram-libs/validate"

	local "github.com/techjanitor/pram-admin/config"
	c "github.com/techjanitor/pram-admin/controllers"
	u "github.com/techjanitor/pram-admin/utils"
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
	auth.Secret = local.Settings.Session.Secret

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
	cors.SetDomains(local.Settings.CORS.Sites, strings.Split("POST,DELETE", ","))

}

func main() {
	r := gin.Default()

	r.Use(cors.CORS())

	r.NoRoute(c.ErrorController)

	// requires mod perms
	mod := r.Group("/mod")
	mod.Use(validate.ValidateParams())
	mod.Use(auth.Auth(auth.Moderators))

	mod.DELETE("/tag/:id", c.DeleteTagController)
	mod.DELETE("/imagetag/:image/:tag", c.DeleteImageTagController)
	mod.DELETE("/thread/:id", c.DeleteThreadController)
	mod.DELETE("/post/:thread/:id", c.DeletePostController)
	mod.POST("/sticky/:thread", c.StickyThreadController)
	mod.POST("/close/:thread", c.CloseThreadController)

	// requires admin perms
	admin := r.Group("/admin")
	admin.Use(validate.ValidateParams())
	admin.Use(auth.Auth(auth.Admins))

	admin.DELETE("/thread/:id", c.PurgeThreadController)
	admin.DELETE("/post/:thread/:id", c.PurgePostController)
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
