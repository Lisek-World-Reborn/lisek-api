package routes

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Lisek-World-Reborn/lisek-api/channels"
	"github.com/Lisek-World-Reborn/lisek-api/config"
	"github.com/Lisek-World-Reborn/lisek-api/db"
	"github.com/Lisek-World-Reborn/lisek-api/docker"
	"github.com/Lisek-World-Reborn/lisek-api/logger"
	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
)

type ServerBody struct {
	BaseName string `json:"base_name"`
}

type ServerResponse struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	State         string    `json:"state"`
	Health        string    `json:"health"`
	IP            string    `json:"ip"`
	Region        string    `json:"region"`
	CreatedAt     time.Time `json:"created_at"`
	ContainerName string    `json:"container_name"`
	Port          int       `json:"port"`
}

func GetServer(c *gin.Context) {
	server := db.Server{}
	db.OpenedConnection.First(&server, c.Param("id"))
	c.JSON(200, server)
}

func CreateServer(c *gin.Context) {

	if c.GetHeader("Authorization") != config.LoadedConfiguration.Secret {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	server := db.Server{}
	c.BindJSON(&server)
	db.OpenedConnection.Create(&server)
	c.JSON(200, server)
}

func UpdateServer(c *gin.Context) {

	if c.GetHeader("Authorization") != config.LoadedConfiguration.Secret {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	server := db.Server{}
	db.OpenedConnection.First(&server, c.Param("id"))
	c.BindJSON(&server)
	db.OpenedConnection.Save(&server)
	c.JSON(200, server)
}

func DeleteServer(c *gin.Context) {

	if c.GetHeader("Authorization") != config.LoadedConfiguration.Secret {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	server := db.Server{}
	db.OpenedConnection.First(&server, c.Param("id"))
	db.OpenedConnection.Delete(&server)
	c.JSON(200, gin.H{"status": "ok"})
}

func PostServerMessage(c *gin.Context) {
	server := db.Server{}
	db.OpenedConnection.First(&server, c.Param("id"))

	message := channels.MinecraftRequest{
		UUID:      c.Param("uuid"),
		Target:    c.Param("target"),
		Arguments: strings.Split(c.Param("arguments"), ","),
		Hash:      c.Param("hash"),
	}

	converted := strconv.Itoa(int(server.ID))

	result, err := message.SendToServer(converted)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !result {
		c.JSON(500, gin.H{"error": "invalid request"})
		return
	}

	c.JSON(200, gin.H{"status": "ok"})
}

func GenerateServer(c *gin.Context) {

	if c.GetHeader("Authorization") != config.LoadedConfiguration.Secret {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	var body ServerBody

	if c.BindJSON(&body) != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	var latestServer db.Server

	db.OpenedConnection.Last(&latestServer)

	lastId := 0

	if latestServer.ID > 1 {
		lastId = int(latestServer.ID)
		logger.Info("Last server was not null!")
	}

	logger.Info("Generating server with port " + strconv.Itoa(25565+lastId))

	server := db.Server{
		Name:          body.BaseName,
		IP:            "0.0.0.0",
		Region:        "eu",
		CreatedAt:     time.Now(),
		ContainerName: "server-" + body.BaseName + "-" + strconv.Itoa(int(time.Now().Unix())),
		Port:          25565 + lastId,
	}

	db.OpenedConnection.Create(&server)

	go docker.CreateServer(server)

	c.JSON(200, server)
}

func ServerStatus(c *gin.Context) {
	server := db.Server{}
	db.OpenedConnection.First(&server, c.Param("id"))

	containers, err := docker.DockerClient.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	for _, container := range containers {
		if container.Names[0] == "/"+server.ContainerName {
			c.JSON(200, gin.H{"status": "online", "health": container.Status, "state": container.State})
			return
		}
	}

	c.JSON(200, gin.H{"status": "offline"})
}

func GetServers(c *gin.Context) {
	servers := []db.Server{}
	db.OpenedConnection.Find(&servers)

	response := []ServerResponse{}

	for _, server := range servers {
		containers, err := docker.DockerClient.ContainerList(context.Background(), types.ContainerListOptions{
			All: true,
		})

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		for _, container := range containers {
			if container.Names[0] == "/"+server.ContainerName {
				response = append(response, ServerResponse{
					ID:            server.ID,
					Name:          server.Name,
					Status:        "online",
					State:         container.State,
					Health:        container.Status,
					IP:            server.IP,
					Region:        server.Region,
					CreatedAt:     server.CreatedAt,
					ContainerName: server.ContainerName,
					Port:          server.Port,
				})
			}
		}
	}

	c.JSON(200, response)
}
