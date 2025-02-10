package config

import (
	"net"
	"os"

	"github.com/pkg/errors"
)

const (
	hostenvName = "SERVER_HOST"
	portenvName = "SERVER_PORT"
)

type ServerConfig interface {
	Address() string
}

type serverConfig struct {
	host string
	port string
}

func NewServerConfig() (ServerConfig, error) {
	host := os.Getenv(hostenvName)
	if len(host) == 0 {
		return nil, errors.New("server host not found")
	}

	port := os.Getenv(portenvName)
	if len(port) == 0 {
		return nil, errors.New("server port not found")
	}

	return &serverConfig{
		host: host,
		port: port,
	}, nil
}

func (cfg *serverConfig) Address() string {
	return net.JoinHostPort(cfg.host, cfg.port)
}
