package service

import (
	"context"
	"errors"
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

func TestValidateData(t *testing.T) {
	tests := []struct {
		name    string
		user    models.AuthReq
		wantErr error
	}{
		{
			name:    "Пустой логин",
			user:    models.AuthReq{Username: "", Password: "password123"},
			wantErr: errors.New("заполните поле логин"),
		},
		{
			name:    "Пустой пароль",
			user:    models.AuthReq{Username: "user1", Password: ""},
			wantErr: errors.New("заполните поле пароль"),
		},
		{
			name:    "Логин совпадает с паролем",
			user:    models.AuthReq{Username: "user1", Password: "user1"},
			wantErr: errors.New("логин и пароль совпадают"),
		},
		{
			name:    "Логин содержит недопустимые символы",
			user:    models.AuthReq{Username: "user!name", Password: "password123"},
			wantErr: errors.New("логин содержит недопустимые символы"),
		},
		{
			name:    "Пароль содержит недопустимые символы",
			user:    models.AuthReq{Username: "username", Password: "pass@word"},
			wantErr: errors.New("пароль содержит недопустимые символы"),
		},
		{
			name:    "Корректные данные",
			user:    models.AuthReq{Username: "validUser", Password: "securePass"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateData(tt.user)
			if (err != nil) != (tt.wantErr != nil) || (err != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("validateData() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
