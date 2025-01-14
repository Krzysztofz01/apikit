package config

import (
	"net/url"

	"github.com/Krzysztofz01/apikit/internal/utils"
)

func ValidateServer(c *ApiKitServerConfiguration) (bool, string) {
	return c.isValid()
}

type ApiKitServerConfiguration struct {
	ApiKit      *ApiKitConfiguration
	Endpoints   []*ApiKitServerEndpointConfiguration
	ApiKeys     []*ApiKitServerKeyConfiguration
	VerboseMode bool
	Host        string
}

func (c *ApiKitServerConfiguration) isValid() (bool, string) {
	// NOTE: Inner apikit config values validation
	if valid, msg := c.ApiKit.isValid(); !valid {
		return false, msg
	}

	if len(c.Host) == 0 {
		return false, "invalid server host"
	}

	serverKeyName := utils.NewEmptySet[string]()
	serverKeySecret := utils.NewEmptySet[string]()

	// NOTE: Server api keys names and secrets unique validation
	for _, serverKey := range c.ApiKeys {
		// NOTE: Inner api key config validation
		if valid, msg := serverKey.isValid(); !valid {
			return false, msg
		}

		if !serverKeyName.Add(serverKey.Name) {
			return false, "duplicate api key name found"
		}

		if !serverKeySecret.Add(serverKey.Secret) {
			return false, "duplicate api key secret found"
		}
	}

	for _, endpoint := range c.Endpoints {
		// NOTE: Inner server endpoint config values validation
		if valid, msg := endpoint.isValid(); !valid {
			return false, msg
		}

		// NOTE: Server endpoints api key pool name existance check
		for _, apiKey := range endpoint.RequiredApiKeyPool {
			if !serverKeyName.Contains(apiKey) {
				return false, "endpoint referencing non existing api key"
			}
		}
	}

	return true, ""
}

type ApiKitServerKeyConfiguration struct {
	Name   string
	Secret string
}

func (c *ApiKitServerKeyConfiguration) isValid() (bool, string) {
	if len(c.Name) == 0 {
		return false, "invalid server key name"
	}

	if len(c.Name) == 0 {
		return false, "invalid server key secret"
	}

	return true, ""
}

type ApiKitServerEndpointConfiguration struct {
	EndpointName       string
	Path               string
	RequiredApiKeyPool []string
}

func (c *ApiKitServerEndpointConfiguration) isValid() (bool, string) {
	if len(c.EndpointName) == 0 {
		return false, "invalid endpoint name"
	}

	if parsedPath, err := url.Parse(c.Path); err != nil || parsedPath.Host != "" || parsedPath.Scheme != "" {
		return false, "invalid path format"
	}

	if c.RequiredApiKeyPool == nil {
		return false, "uninitialized required api key pool collection"
	}

	return true, ""
}
