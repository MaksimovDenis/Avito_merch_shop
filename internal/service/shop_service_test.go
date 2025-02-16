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

func TestByItem(t *testing.T) {
	ctx := context.Background()
	port := "5992"

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
	auth := NewService(*repo, clientDb, token, log)

	tests := []struct {
		name    string
		args    models.AuthReq
		want    *models.Items
		wantErr bool
	}{
		{
			name: "OK",
			args: models.AuthReq{
				Username: "user",
				Password: "password",
			},
			want: &models.Items{
				Name:     "book",
				Quantity: 1,
			},
			wantErr: false,
		},
		{
			name: "Incorrect product name",
			args: models.AuthReq{
				Username: "user2",
				Password: "password",
			},
			want: &models.Items{
				Name:     "books",
				Quantity: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newUser, err := repo.Authorization.CreateUser(ctx, tt.args)
			require.NoError(t, err)

			if !tt.wantErr {
				err = auth.BuyItem(ctx, newUser.Id, "book")
				require.NoError(t, err)

				var item models.Items

				query := db.Query{
					Name: "BuyItem",
					QueryRow: `
						SELECT p.name as name, sum(pu.quantity) as quantity
						FROM purchases pu
						JOIN products p ON pu.products_id = p.id
						WHERE pu.user_id = $1
						GROUP BY name
					`,
				}

				err = clientDb.DB().QueryRowContext(ctx, query, newUser.Id).Scan(&item.Name, &item.Quantity)
				require.NoError(t, err)
				assert.Equal(t, tt.want, &item)
			} else {
				err = auth.BuyItem(ctx, newUser.Id, "books")
				require.Error(t, err)
				assert.EqualError(t, err, "[books] не найден")
			}
		})
	}
}

func TestSendCoin(t *testing.T) {
	ctx := context.Background()

	port := "5993"

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
	auth := NewService(*repo, clientDb, token, log)

	tests := []struct {
		name    string
		args    []models.AuthReq
		want    int
		wantErr bool
	}{
		{
			name: "OK",
			args: []models.AuthReq{
				{
					Username: "user1",
					Password: "password",
				},
				{
					Username: "user2",
					Password: "password",
				},
			},
			want:    1100,
			wantErr: false,
		},
		{
			name: "Not anough money",
			args: []models.AuthReq{
				{
					Username: "user3",
					Password: "password",
				},
				{
					Username: "user4",
					Password: "password",
				},
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user1, err := repo.Authorization.CreateUser(ctx, tt.args[0])
			require.NoError(t, err)
			user2, err := repo.Authorization.CreateUser(ctx, tt.args[1])
			require.NoError(t, err)

			if !tt.wantErr {
				err = auth.SendCoins(ctx, user1.Username, user2.Username, 100)
				require.NoError(t, err)

				var coins int

				query := db.Query{
					Name:     "SendCoin",
					QueryRow: `SELECT coins FROM users WHERE id = $1`,
				}

				err = clientDb.DB().QueryRowContext(ctx, query, user2.Id).Scan(&coins)
				require.NoError(t, err)
				assert.Equal(t, tt.want, coins)
			} else {
				err = auth.SendCoins(ctx, user1.Username, user2.Username, 1200)
				require.Error(t, err)
				assert.EqualError(t, err, "недостаточно монет для перевода")
			}
		})
	}
}

func TestInfo(t *testing.T) {
	ctx := context.Background()
	port := "5994"

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
	auth := NewService(*repo, clientDb, token, log)

	type wantStruct struct {
		coins         int
		items         []models.Items
		sentCoins     []models.SentCoins
		receivedCoins []models.ReceivedCoins
	}

	wantRes := &wantStruct{
		coins: 1450,
		items: []models.Items{
			{
				Name:     "book",
				Quantity: 1,
			},
		},
		sentCoins: []models.SentCoins{
			{
				ToUser: "user1",
				Amount: 500,
			},
		},
		receivedCoins: []models.ReceivedCoins{
			{
				FromUser: "user1",
				Amount:   1000,
			},
		},
	}

	tests := []struct {
		name string
		args []models.AuthReq
		want *wantStruct
	}{
		{
			name: "OK",
			args: []models.AuthReq{
				{
					Username: "user1",
					Password: "password",
				},
				{
					Username: "user2",
					Password: "password",
				},
			},
			want: wantRes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user1, err := repo.Authorization.CreateUser(ctx, tt.args[0])
			require.NoError(t, err)

			user2, err := repo.Authorization.CreateUser(ctx, tt.args[1])
			require.NoError(t, err)

			err = auth.SendCoins(ctx, user1.Username, user2.Username, 1000)
			require.NoError(t, err)

			err = auth.SendCoins(ctx, user2.Username, user1.Username, 500)
			require.NoError(t, err)

			err = auth.Shop.BuyItem(ctx, user2.Id, "book")
			require.NoError(t, err)

			coins, items, sentCoins, receivedCoins, err := auth.Shop.Info(ctx, user2.Username)
			require.NoError(t, err)

			assert.Equal(t, tt.want.coins, coins)
			assert.Equal(t, tt.want.items, items)
			assert.Equal(t, tt.want.sentCoins, sentCoins)
			assert.Equal(t, tt.want.receivedCoins, receivedCoins)
		})
	}
}
