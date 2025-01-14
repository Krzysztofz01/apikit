package client

import (
	"fmt"
	"net/http"

	"github.com/Krzysztofz01/apikit/internal/config"
	"github.com/Krzysztofz01/apikit/internal/log"
	"github.com/Krzysztofz01/apikit/internal/source"
)

type ApiKitClient interface {
	Get(endpointName string) (map[string]interface{}, error)
}

type apiKitClient struct {
	httpClient *http.Client
	sources    map[string]source.Source
	lookup     EndpointLookup
	logger     log.Loggerp
}

func CreateApiKitClient(h *http.Client, c *config.ApiKitConfiguration, l log.Logger) (ApiKitClient, error) {
	return CreateNamedApiKitClient("", h, c, l)
}

func CreateNamedApiKitClient(name string, h *http.Client, c *config.ApiKitConfiguration, l log.Logger) (ApiKitClient, error) {
	if h == nil {
		return nil, fmt.Errorf("client: invalid nil reference http client provided")
	}

	if c == nil {
		return nil, fmt.Errorf("client: invalid nil reference configuration provided")
	}

	if l == nil {
		return nil, fmt.Errorf("source: provided logger reference is nil")
	}

	prefix := "Client"
	if len(name) != 0 {
		prefix = fmt.Sprintf("Client - %s", name)
	}

	logger := log.CreatePrefixedLogger(prefix, l)

	sources := make(map[string]source.Source, len(c.Sources))
	for _, sourceConfig := range c.Sources {
		if source, err := source.CreateSource(h, sourceConfig, l); err != nil {
			return nil, fmt.Errorf("client: failed to create source instance: %w", err)
		} else {
			sources[sourceConfig.Name] = source
		}
	}

	lookup, err := createEndpointLookup(c)
	if err != nil {
		return nil, fmt.Errorf("client: failed to create the client endpoint lookup: %w", err)
	}

	return &apiKitClient{
		httpClient: h,
		sources:    sources,
		lookup:     lookup,
		logger:     logger,
	}, nil
}

func (c *apiKitClient) Get(endpointName string) (map[string]interface{}, error) {
	sourceValuesMap, err := c.lookup.GetEndpointSourcesWithSourceValueNames(endpointName)
	if err != nil {
		return nil, fmt.Errorf("client: failed to access the endpoint via lookup: %w", err)
	}

	result := make(map[string]interface{}, 0)
	for sourceName, sourceValueNames := range sourceValuesMap {
		source, ok := c.sources[sourceName]
		if !ok {
			return nil, fmt.Errorf("client: specified source not found")
		}

		sourceValues, err := source.GetValues(sourceValueNames)
		if err != nil {
			return nil, fmt.Errorf("client: failed to access the source values: %w", err)
		}

		for sourceValueName, sourceValue := range sourceValues {
			if endpointValueName, err := c.lookup.GetEndpointValueName(sourceName, sourceValueName); err != nil {
				return nil, fmt.Errorf("client: failed to access the endpoint value name via lookup: %w", err)
			} else {
				result[endpointValueName] = sourceValue
			}
		}
	}

	return result, nil
}
