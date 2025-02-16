package service

import (
	"context"
	"testing"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/client/db/pg"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/models"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/repository"
	pgcontainer "github.com/MaksimovDenis/Avito_merch_shop/pkg/pg_container"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	port := "5990"

	cli, containerID, err := pgcontainer.SetupPostgresContainer(ctx, port)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		require.NoError(t, err)
	}()

	conStr := "postgres://admin:admin@localhost:" + port + "/testDB?sslmode=disable"

	clientDb, err := pg.New(ctx, conStr)
	if err != nil {
		t.Fatal(err)
	}
	defer clientDb.Close()

	var log zerolog.Logger

	var token token.JWTMaker

	repo := repository.NewRepository(clientDb, log)
	authSvc := newAuthService(*repo, token, log)

	tests := []struct {
		name    string
		args    models.AuthReq
		want    string
		wantErr bool
	}{
		{
			name: "OK",
			args: models.AuthReq{
				Username: "userTest",
				Password: "passwordTest",
			},
			want:    "userTest",
			wantErr: false,
		},
		{
			name: "Error",
			args: models.AuthReq{
				Username: "userTest2",
				Password: "passwordTest",
			},
			want:    "userTest",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = authSvc.appRepository.Authorization.CreateUser(ctx, tt.args)
			require.NoError(t, err)

			var title string

			query := db.Query{
				Name:     "Create User",
				QueryRow: "SELECT username FROM users WHERE id = 1",
			}

			err = clientDb.DB().QueryRowContext(ctx, query).Scan(&title)
			require.NoError(t, err)

			if tt.wantErr {
				assert.NotEqual(t, tt.args.Username, tt.want)
			} else {
				assert.Equal(t, tt.args.Username, tt.want)
			}
		})
	}
}

func TestAuth(t *testing.T) {
	ctx := context.Background()
	port := "5991"

	cli, containerID, err := pgcontainer.SetupPostgresContainer(ctx, port)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		require.NoError(t, err)
	}()

	conStr := "postgres://admin:admin@localhost:" + port + "/testDB?sslmode=disable"

	clientDb, err := pg.New(ctx, conStr)
	if err != nil {
		t.Fatal(err)
	}
	defer clientDb.Close()

	var log zerolog.Logger

	var token token.JWTMaker

	repo := repository.NewRepository(clientDb, log)
	authSvc := NewService(*repo, clientDb, token, log)

	type args struct {
		auth models.AuthReq
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
				auth: models.AuthReq{
					Username: "userTest",
					Password: "passwordTest",
				},
			},
			want:    "userTest",
			wantErr: false,
		},
		{
			name: "Empty Username",
			args: args{
				auth: models.AuthReq{
					Username: "",
					Password: "passwordTest",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Empty Password",
			args: args{
				auth: models.AuthReq{
					Username: "userTest2",
					Password: "",
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "User==password",
			args: args{
				auth: models.AuthReq{
					Username: "admin",
					Password: "admin",
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := authSvc.Auth(ctx, tt.args.auth)
			if !tt.wantErr {
				require.NoError(t, err)

				var title string

				query := db.Query{
					Name:     "Create User",
					QueryRow: "SELECT username FROM users WHERE id = 1",
				}

				err = clientDb.DB().QueryRowContext(ctx, query).Scan(&title)
				require.NoError(t, err)

				assert.Equal(t, tt.want, title)
			} else {
				require.Error(t, err)
			}
		})
	}
}
