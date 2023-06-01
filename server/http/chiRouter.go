package http

import (
	"time"

	"github.com/rs/zerolog"
)

type Server struct {
}

func (s *Server) Shutdown() error {
	return nil
}

func NewChiRouter(s interface{}, config1 string, config2 time.Duration, logger *zerolog.Logger) (interface{}, error) {
	return new(interface{}), nil
}

func NewHTTP(r interface{}, address string, logger *zerolog.Logger) (Server, error) {
	return Server{}, nil
}