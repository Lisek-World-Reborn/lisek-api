package main

import (
	"flag"
	"strconv"

	"github.com/Lisek-World-Reborn/lisek-api/config"
	"github.com/Lisek-World-Reborn/lisek-api/logger"
	"github.com/Lisek-World-Reborn/lisek-api/routes"
	"github.com/gin-gonic/gin"
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

	logger.Info("Starting http server")
	r := gin.Default()

	r.GET("/", routes.Index)
	r.GET("/load", routes.GetSystemLoad)

	r.Run(":" + strconv.Itoa(int(cfg.Port)))

	logger.Info("Server started")
}
