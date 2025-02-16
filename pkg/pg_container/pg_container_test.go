package pgcontainer

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		require.NoError(t, err)
	}()
}
