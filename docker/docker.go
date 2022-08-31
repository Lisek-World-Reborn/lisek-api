package docker

import (
	"context"
	"os"

	"github.com/Lisek-World-Reborn/lisek-api/db"
	"github.com/Lisek-World-Reborn/lisek-api/logger"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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

	return
}

func CreateServer(server db.Server) {
	ctx := context.Background()
	resp, err := DockerClient.ContainerCreate(ctx, &container.Config{
		Image: SERVER_IMAGE,
		Env: []string{
			"VERSION=1.12.2",
			"EULA=TRUE",
			"TYPE=PAPER",
			"MAX_MEMORY=2048M",
			"TZ=Europe/Kiev",
			"USE_AIKAR_FLAGS=true",
		},
	},
		&container.HostConfig{
			Binds: []string{
				DATA_DIR + "/" + server.Name + ":/data",
			}},
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

}
