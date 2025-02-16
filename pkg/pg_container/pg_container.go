package pgcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func SetupPostgresContainer(ctx context.Context, port string) (*client.Client, string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, "", fmt.Errorf("failed to create docker client: %w", err)
	}

	containerConfig := &container.Config{
		Image: "postgres:latest",
		Env: []string{
			"POSTGRES_USER=admin",
			"POSTGRES_PASSWORD=admin",
			"POSTGRES_DB=testDB",
		},
		ExposedPorts: nat.PortSet{
			"5432/tcp": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"5432/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
	}

	networkingConfig := &network.NetworkingConfig{}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, "my_postgres"+port)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create docker container: %w", err)
	}

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to start docker container: %w", err)
	}

	fmt.Println("PostgreSQL контейнер запущен с ID:", resp.ID)
	time.Sleep(5 * time.Second)

	return cli, resp.ID, nil
}
