package service

import (
	"context"
	"time"

	db "github.com/MaksimovDenis/Avito_merch_shop/internal/client"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/models"
	"github.com/MaksimovDenis/Avito_merch_shop/internal/repository"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/token"
	"github.com/MaksimovDenis/Avito_merch_shop/pkg/util"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const durationAccessToken time.Duration = 24 * time.Hour

type Authorization interface {
	Auth(ctx context.Context, req *models.AuthReq) (string, error)
}

type AuthService struct {
	appRepository repository.Repository
	txManager     db.TxManager
	token         token.JWTMaker
	log           zerolog.Logger
}

func newAuthService(
	appRepository repository.Repository,
	txManager db.TxManager,
	token token.JWTMaker,
	log zerolog.Logger,
) *AuthService {
	return &AuthService{
		appRepository: appRepository,
		txManager:     txManager,
		token:         token,
		log:           log,
	}
}

func (auth *AuthService) Auth(ctx context.Context, req *models.AuthReq) (string, error) {
	user, err := auth.appRepository.Authorization.GetUser(ctx, req.Username)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			accessToken, err := auth.CreateUser(ctx, req)
			if err != nil {
				return "", err
			}

			return accessToken, nil
		} else {
			auth.log.Error().Err(err).Msg("failed to get user from storage")
			return "", err
		}
	}

	if err = util.CheckPassword(req.Password, user.Password); err != nil {
		auth.log.Error().Err(err).Msg("password mismatch")
		return "", err
	}

	return auth.generateToken(user)

}

func (auth *AuthService) CreateUser(ctx context.Context, req *models.AuthReq) (string, error) {
	hashedPwd, err := util.HashPassword(req.Password)
	if err != nil {
		auth.log.Error().Err(err).Msg("failed to hash password")
		return "", err
	}

	req.Password = hashedPwd

	newUser, err := auth.appRepository.Authorization.CreateUser(ctx, req)
	if err != nil {
		auth.log.Error().Err(err).Msg("failed to create new user in storage")
		return "", err
	}

	return auth.generateToken(newUser)
}

func (auth *AuthService) generateToken(user *models.User) (string, error) {
	accessToken, _, err := auth.token.CreateToken(int64(user.Id), user.Username, durationAccessToken)
	if err != nil {
		auth.log.Error().Err(err).Msg("failed to create access token")
		return "", err
	}
	return accessToken, nil
}
