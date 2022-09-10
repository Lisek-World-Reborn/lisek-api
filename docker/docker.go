package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/Lisek-World-Reborn/lisek-api/channels"
	"github.com/Lisek-World-Reborn/lisek-api/db"
	"github.com/Lisek-World-Reborn/lisek-api/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

var DockerClient *client.Client

const SERVER_IMAGE = "docker.io/itzg/minecraft-server"

var DATA_DIR = os.Getenv("DATA_DIR")

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

func PreloadServers() {

	logger.Info("Preloading servers")

	files, err := os.ReadDir("preloaded")

	if err != nil {
		logger.Error("Error reading preloaded servers directory: " + err.Error())
		return
	}

	for _, file := range files {

		if file.IsDir() {

			logger.Info("Preloading server " + file.Name())

			tarPath := path.Join("preloaded", file.Name(), "server.tar.gz")

			buildContext, err := os.Open(tarPath)

			if err != nil {
				logger.Error("Error opening build context: " + err.Error())
				return
			}

			defer buildContext.Close()

			resp, err := DockerClient.ImageBuild(context.Background(), nil, types.ImageBuildOptions{
				Dockerfile: path.Join("./preloaded", file.Name(), "Dockerfile"),
				Tags:       []string{file.Name()},
				Remove:     false,
				Context:    buildContext,
			})

			if err != nil {
				logger.Error("Error building image: " + err.Error())
				return
			}

			scanner := bufio.NewScanner(resp.Body)

			for scanner.Scan() {
				logger.Info(scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				logger.Error("Error reading image build output: " + err.Error())
				return
			}

			logger.Info("Server preloaded: " + file.Name())

			logger.Info("Launching server " + file.Name())

			ctx, _ := context.WithTimeout(context.TODO(), time.Minute*5)

			serverBind := path.Join(DATA_DIR, "servers", file.Name())

			os.MkdirAll(path.Join("/data", "servers", file.Name()), os.ModePerm)

			response, err := DockerClient.ContainerCreate(ctx, &container.Config{
				Image: file.Name(),
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
				&network.NetworkingConfig{}, &v1.Platform{}, file.Name())

			if err != nil {
				logger.Error("Error creating container: " + err.Error())
				return
			}

			logger.Info("Container created: " + response.ID)

		}
	}
}
