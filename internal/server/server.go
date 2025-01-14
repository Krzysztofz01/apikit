package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Krzysztofz01/apikit/internal/client"
	"github.com/Krzysztofz01/apikit/internal/config"
	"github.com/Krzysztofz01/apikit/internal/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ApiKitServer interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type apiKitServer struct {
	server                 *echo.Echo
	apiKitClient           client.ApiKitClient
	endpointNamePathLookup map[string]string
	logger                 log.Loggerp
	isStarted              bool
	cfg                    *config.ApiKitServerConfiguration
}

func CreateApiKitServer(h *http.Client, c *config.ApiKitServerConfiguration, l log.Logger) (ApiKitServer, error) {
	if h == nil {
		return nil, fmt.Errorf("server: invalid nil reference http client provided")
	}

	if c == nil {
		return nil, fmt.Errorf("server: invalid nil reference configuration provided")
	}

	if l == nil {
		return nil, fmt.Errorf("server: provided logger reference is nil")
	}

	logger := log.CreatePrefixedLogger("Server", l)

	logger.Infof("Apikit client setup started")

	apiKitClient, err := client.CreateApiKitClient(h, c.ApiKit, l)
	if err != nil {
		return nil, fmt.Errorf("server: failed to create apikit client: %w", err)
	}

	logger.Infof("Apikit client setup finished")
	logger.Infof("Apikit server setup started")

	server := echo.New()

	server.Use(middleware.Recover())
	server.Use(middleware.CORS())

	logger.Infof("Apikit server setup finished")

	endpointNamePathLookup := make(map[string]string, len(c.Endpoints))
	for _, endpoint := range c.Endpoints {
		endpointNamePathLookup[endpoint.Path] = endpoint.EndpointName
	}

	apiKitServer := &apiKitServer{
		server:                 server,
		apiKitClient:           apiKitClient,
		endpointNamePathLookup: endpointNamePathLookup,
		logger:                 logger,
		isStarted:              false,
		cfg:                    c,
	}

	if err := apiKitServer.RegisterEndpoints(); err != nil {
		return nil, fmt.Errorf("server: failed to register endpoints: %w", err)
	}

	return apiKitServer, nil
}

func (s *apiKitServer) Shutdown(ctx context.Context) error {
	if !s.isStarted {
		return nil
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server: server shutdown failure: %w", err)
	} else {
		s.isStarted = false
	}

	return nil
}

func (s *apiKitServer) Start() error {
	if s.isStarted {
		return fmt.Errorf("sever: can not start a running server")
	}

	if err := s.server.Start(s.cfg.Host); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: runtime failure occured: %w", err)
	} else {
		s.isStarted = true
	}

	return nil
}

func (s *apiKitServer) RegisterEndpoints() error {
	if s.isStarted {
		return fmt.Errorf("server: failed to register endpoints with the server running")
	}

	for _, endpoint := range s.cfg.Endpoints {
		s.server.GET(endpoint.Path, s.GetRequestHandle)
	}

	return nil
}

func (s *apiKitServer) GetRequestHandle(c echo.Context) error {
	t := time.Now()

	endpointName, ok := s.endpointNamePathLookup[c.Path()]
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}

	result, err := s.apiKitClient.Get(endpointName)
	if err != nil {
		// TODO: Better error handling that will be able to tell the difference between 4xx and 5xx
		s.logger.Errorf("HTTP 500 %s in %dms failed with %s", c.Path(), time.Since(t).Milliseconds(), err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}

	s.logger.Infof("HTTP 200 %s in %dms", c.Path(), time.Since(t).Milliseconds())
	return c.JSON(http.StatusOK, result)
}
