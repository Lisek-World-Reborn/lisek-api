package docker

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/Lisek-World-Reborn/lisek-api/channels"
	"github.com/Lisek-World-Reborn/lisek-api/db"
	"github.com/Lisek-World-Reborn/lisek-api/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var DockerClient *client.Client

const SERVER_IMAGE = "docker.io/itzg/minecraft-server"

var DATA_DIR = os.Getenv("DATA_DIR")
var PRELOADED_DIR = os.Getenv("PRELOADED_DIR")

type PreloadedServer struct {
	Mounts  []interface{}     `json:"mounts,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Name    string            `json:"name,omitempty"`
	Folders map[string]string `json:"folders,omitempty"`
}

func Init() {

	logger.Info("Initializing docker")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		logger.Info("Error initializing docker client: " + err.Error())
		os.Exit(0)
		return
	}

	DockerClient = cli

	logger.Info("Docker initialized")
}

func CreateServer(server db.Server) {

	logger.Info("Creating container for server " + server.Name + "(" + server.ContainerName + ")")

	ctx, _ := context.WithTimeout(context.TODO(), time.Minute*5)

	logger.Info("Pulling container image")

	_, err := DockerClient.ImagePull(ctx, SERVER_IMAGE, types.ImagePullOptions{})

	if err != nil {
		logger.Info("Error pulling container image: " + err.Error())
		os.Exit(0)
		return
	}

	serverBind := path.Join(DATA_DIR, "servers", server.ContainerName)

	os.MkdirAll(path.Join("/data", "servers", server.ContainerName), os.ModePerm)

	resp, err := DockerClient.ContainerCreate(ctx, &container.Config{
		Image: SERVER_IMAGE,
		Env: []string{
			"VERSION=1.12.2",
			"EULA=TRUE",
			"TYPE=PAPER",
			"MAX_MEMORY=2048M",
			"TZ=Europe/Kiev",
			"USE_AIKAR_FLAGS=true",
			"ONLINE_MODE=false",
		},
	},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   serverBind,
					Target:   "/data",
					ReadOnly: false,
				},
			},
		},
		&network.NetworkingConfig{}, &v1.Platform{}, server.ContainerName)

	if err != nil {
		logger.Error("Error creating container: " + err.Error())
		return
	}

	logger.Info("Container created: " + resp.ID)

	err = DockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})

	if err != nil {
		logger.Error("Error starting container: " + err.Error())
		return
	}

	logger.Info("Container started: " + resp.ID)

	addedServerRequest := channels.ServerAddedRequest{
		ServerId: int(server.ID),
	}

	addedServerRequestJson, err := json.Marshal(addedServerRequest)

	if err != nil {
		logger.Error("Error marshalling server added request: " + err.Error())
		return
	}

	channels.RedisConnection.Publish(context.Background(), "servers:added", addedServerRequestJson)

}

func serverExists(name string) bool {

	containerList, err := DockerClient.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		logger.Error("Error listing containers: " + err.Error())
		return false
	}

	for _, container := range containerList {
		if container.Names[0] == "/"+name {
			return true
		}
	}

	return false
}

func createPreloadContainer(name string) {

	server := db.Server{}

	db.OpenedConnection.First("container_name = ?", name).First(&server)

	if server.ID == 0 {
		logger.Error("Server not found")
		return
	}

	preloadedServer := PreloadedServer{}

	preloadedServerJson, err := os.ReadFile(path.Join(PRELOADED_DIR, server.ContainerName+".json"))

	if err != nil {
		logger.Error("Error reading preloaded server file: " + err.Error())
		return
	}

	err = json.Unmarshal(preloadedServerJson, &preloadedServer)

	if err != nil {
		logger.Error("Error unmarshalling preloaded server file: " + err.Error())
		return
	}

	logger.Info("Creating container for server " + server.Name + "(" + server.ContainerName + ")")

	ctx, _ := context.WithTimeout(context.TODO(), time.Minute*5)

	logger.Info("Pulling container image")

	_, err = DockerClient.ImagePull(ctx, SERVER_IMAGE, types.ImagePullOptions{})

	if err != nil {
		logger.Info("Error pulling container image: " + err.Error())
		os.Exit(0)
		return
	}

	serverBind := path.Join(DATA_DIR, "servers", server.ContainerName)

	os.MkdirAll(path.Join("/data", "servers", server.ContainerName), os.ModePerm)

	mounts := []mount.Mount{
		{
			Type:     mount.TypeBind,
			Source:   serverBind,
			Target:   "/data",
			ReadOnly: false,
		},
	}

	envs := []string{}

	for key, value := range preloadedServer.Env {
		envs = append(envs, key+"="+value)
	}

	serverPort := strconv.Itoa(server.Port)

	container, err := DockerClient.ContainerCreate(context.Background(), &container.Config{
		Image: SERVER_IMAGE,
		Env:   envs,
	}, &container.HostConfig{
		Mounts: mounts,
		PortBindings: nat.PortMap{
			"25565/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: serverPort,
				},
			},
		},
	}, nil, nil, server.ContainerName)

	if err != nil {
		logger.Error("Error creating container: " + err.Error())
		return
	}

	logger.Info("Container created: " + container.ID)

	startPreloadedContainer(server.ContainerName)
}

func startPreloadedContainer(name string) {

	container, err := DockerClient.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})

	if err != nil {
		logger.Error("Error listing containers: " + err.Error())
		return
	}

	for _, container := range container {
		if container.Names[0] == "/"+name {
			err = DockerClient.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})

			if err != nil {
				logger.Error("Error starting container: " + err.Error())
				return
			}

			logger.Info("Container started: " + container.ID)
		}
	}

	createPreloadContainer(name)

}
func PreloadServers() {

	logger.Info("Preloading servers")

	files, err := os.ReadDir("preloaded")

	if err != nil {
		logger.Error("Error reading preloaded servers directory: " + err.Error())
		return
	}

	for _, file := range files {

		if file.IsDir() {

			if serverExists(file.Name()) {
				logger.Info("Server " + file.Name() + " already exists, starting...")

				startPreloadedContainer(file.Name())
				continue
			}

			logger.Info("Preloading server " + file.Name())

			preloadedServer := PreloadedServer{}

			preloadedServerJson, err := os.ReadFile(path.Join("preloaded", file.Name(), "info.json"))

			if err != nil {
				logger.Error("Error reading preloaded server info: " + err.Error())
				continue
			}

			err = json.Unmarshal(preloadedServerJson, &preloadedServer)

			if err != nil {
				logger.Error("Error unmarshalling preloaded server info: " + err.Error())
				continue
			}

			envs := []string{}

			for key, value := range preloadedServer.Env {
				envs = append(envs, key+"="+value)
			}

			var latestServer db.Server

			db.OpenedConnection.Last(&latestServer)

			lastId := 0

			if latestServer.ID > 1 {
				lastId = int(latestServer.ID)
				logger.Info("Last server was not null!")
			}

			serverPort := strconv.Itoa(25565 + lastId)

			container, err := DockerClient.ContainerCreate(context.Background(), &container.Config{
				Image: SERVER_IMAGE,
				Env:   envs,
			}, &container.HostConfig{
				PortBindings: nat.PortMap{
					"25565/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: serverPort,
						},
					},
				},
			}, nil, nil, file.Name())

			if err != nil {
				logger.Error("Error creating container: " + err.Error())
				continue
			}

			//Inserting in db

			portInt, err := strconv.Atoi(serverPort)

			if err != nil {
				logger.Error("Error converting port to int: " + err.Error())
				continue
			}

			server := db.Server{
				Name:          preloadedServer.Name,
				ContainerName: file.Name(),
				IP:            "0.0.0.0",
				Region:        "eu",
				Port:          portInt,
			}

			db.OpenedConnection.Create(&server)

			err = DockerClient.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
			if err != nil {
				logger.Error("Error starting container: " + err.Error())
				continue
			}

			logger.Info("Container started: " + container.ID)
		}
	}
}
