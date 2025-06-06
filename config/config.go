package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func init() {
	file, err := os.Open("/etc/pram/pram.conf")
	if err != nil {
		// file was not found so use default settings
		Settings = &Config{
			Admin: Admin{
				Host: "127.0.0.1",
				Port: 5020,
			},
			Directories: Directories{
				ImageDir:     "/tmp/eirka/src/",
				ThumbnailDir: "/tmp/eirka/thumb/",
				AvatarDir:    "/tmp/eirka/avatars/",
			},
		}
		return
	}

	// if the file is found fill settings with json
	Settings = &Config{}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

// Settings holds the current config options
var Settings *Config

// Config represents the possible configurable parameters
// for the local daemon
type Config struct {
	Admin       Admin
	Directories Directories
	CORS        CORS
	Database    Database
	Redis       Redis
}

// Admin sets what the daemon listens on
type Admin struct {
	Host                   string
	Port                   uint
	DatabaseMaxIdle        int
	DatabaseMaxConnections int
	RedisMaxIdle           int
	RedisMaxConnections    int
}

// Database holds the connection settings for MySQL
type Database struct {
	Host     string
	Protocol string
	User     string
	Password string
	Database string
}

// Redis holds the connection settings for the redis cache
type Redis struct {
	Host     string
	Protocol string
}

// Directories sets where files will be stored locally
type Directories struct {
	ImageDir     string
	ThumbnailDir string
	AvatarDir    string
}

// CORS is a list of allowed remote addresses
type CORS struct {
	Sites []string
}
