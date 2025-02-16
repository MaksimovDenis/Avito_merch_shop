package pgcontainer

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestSetupPostgresContainer(t *testing.T) {
	ctx := context.Background()
	port := "5433"

	cli, containerID, err := SetupPostgresContainer(ctx, port)
	assert.NoError(t, err)
	assert.NotEmpty(t, containerID)

	containerInfo, err := cli.ContainerInspect(ctx, containerID)
	assert.NoError(t, err)
	assert.True(t, containerInfo.State.Running)

	defer func() {
		cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	}()
}
