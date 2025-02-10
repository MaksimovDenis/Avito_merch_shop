package config

import (
	"errors"
	"log"
	"os"
)

const (
	secretKey = "TOKEN_SECRET_KEY"

	minSecretKeySize = 32
)

type TokenConfig interface {
	SecretKey() string
}

type tokenConfig struct {
	secretKey string
}

func NewSecretKey() (*tokenConfig, error) {
	secretKey := os.Getenv(secretKey)
	if len(secretKey) == 0 {
		return nil, errors.New("secret key of jwt token not found")
	}

	if len(secretKey) < minSecretKeySize {
		log.Fatalf("SECRET_KEY must be at least %d characters", minSecretKeySize)
	}

	return &tokenConfig{
		secretKey: secretKey,
	}, nil
}

func (tkn *tokenConfig) SecretKey() string {
	return tkn.secretKey
}
