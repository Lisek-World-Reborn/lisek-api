package main

import (
	"flag"
	"strconv"

	"github.com/Lisek-World-Reborn/lisek-api/channels"
	"github.com/Lisek-World-Reborn/lisek-api/config"
	"github.com/Lisek-World-Reborn/lisek-api/db"
	"github.com/Lisek-World-Reborn/lisek-api/logger"
	"github.com/Lisek-World-Reborn/lisek-api/routes"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	configurationPath := flag.String("config", "config.yml", "path to configuration file")

	flag.Parse()

	if !config.IsConfigurationExists(*configurationPath) {
		logger.Warning("Configuration file not found. Generating default configuration...")
		config.GenerateDefaultConfiguration(*configurationPath)
	}

	cfg, err := config.GetConfiguration(*configurationPath)

	if err != nil {
		logger.Fatal(err.Error())
		return
	}

	config.LoadedConfiguration = *cfg

	logger.Info("Configuration loaded")

	dsn := cfg.Dsn

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logger.Fatal(err.Error())
		return
	}

	db.OpenedConnection = database

	logger.Info("Database connection opened")

	logger.Info("Connecting to redis.")

	channels.Init()

	logger.Info("Redis connection established.")

	//	go channels.StartAcceptingRequests()

	logger.Warning("Started accepting server requests. - NOT IMPLEMENTED")

	logger.Info("Starting http server")
	r := gin.Default()

	r.GET("/", routes.Index)
	r.GET("/load", routes.GetSystemLoad)
	r.GET("/servers", routes.GetServers)
	r.GET("/servers/:id", routes.GetServer)
	r.POST("/servers", routes.CreateServer)
	r.PUT("/servers/:id", routes.UpdateServer)

	r.DELETE("/servers/:id", routes.DeleteServer)
	r.POST("/servers/:id/data", routes.PostServerMessage)

	r.Run(":" + strconv.Itoa(int(cfg.Port)))

	logger.Info("Server started")
}
