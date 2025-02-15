package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/client/db/pg"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/models"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/repository"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgres(t *testing.T) {
	ctx := context.Background()
	connStr := "postgres://postgres:password@localhost:5454/shop?sslmode=disable"

	_, err := pg.New(ctx, connStr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuth(t *testing.T) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Error().Err(err).Msg("failed to create docker client")
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
					HostPort: "5999",
				},
			},
		},
	}

	networkingConfig := &network.NetworkingConfig{}
	rmOpts := container.RemoveOptions{Force: true}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, "my_postgres")
	if err != nil {
		log.Error().Err(err).Msg("failed to create docker container")
	}
	defer cli.ContainerRemove(ctx, resp.ID, rmOpts)

	var options container.StartOptions
	if err := cli.ContainerStart(ctx, resp.ID, options); err != nil {
		log.Error().Err(err).Msg("failed to start docker container")
	}

	fmt.Println("PostgreSQL контейнер запущен с ID:", resp.ID)

	time.Sleep(5 * time.Second)

	conStr := "postgres://admin:admin@localhost:5999/testDB?sslmode=disable"
	clientDb, err := pg.New(ctx, conStr)
	if err != nil {
		t.Fatal(err)
	}
	defer clientDb.Close()

	var log zerolog.Logger
	var token token.JWTMaker

	repo := repository.NewRepository(clientDb, log)
	authSvc := newAuthService(*repo, token, log)

	type args struct {
		auth *models.AuthReq
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				auth: &models.AuthReq{
					Username: "userTest",
					Password: "passwordTest",
				},
			},
			want:    "userTest",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authSvc.appRepository.Authorization.CreateUser(ctx, tt.args.auth)
			require.NoError(t, err)
			if !tt.wantErr {
				var title string

				query := db.Query{
					Name:     "Create User",
					QueryRow: "SELECT username FROM users WHERE id = 1",
				}

				_ = clientDb.DB().QueryRowContext(ctx, query).Scan(&title)
				require.NoError(t, err)

				fmt.Println(title)
				assert.Equal(t, tt.args.auth.Username, tt.want)
			}
		})
	}
}
