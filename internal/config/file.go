package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

const (
	configFileName string = "config"
	configFileType string = "json"
)

type apiKitConfiguration struct {
	Sources   []*sourceConfiguration   `mapstructure:"sources"`
	Endpoints []*endpointConfiguration `mapstructure:"endpoints"`
}

type apiKitServerConfiguration struct {
	ApiKit      *apiKitConfiguration                 `mapstructure:"general"`
	Endpoints   []*apiKitServerEndpointConfiguration `mapstructure:"endpoints"`
	ApiKeys     []*apiKitServerKeyConfiguration      `mapstructure:"api-keys"`
	VerboseMode bool                                 `mapstructure:"verbose-mode"`
	Host        string                               `mapstructure:"host"`
}

type apiKitServerKeyConfiguration struct {
	Name   string `mapstructure:"name"`
	Secret string `mapstructure:"secret"`
}

type apiKitServerEndpointConfiguration struct {
	EndpointName       string   `mapstructure:"name"`
	Path               string   `mapstructure:"path"`
	RequiredApiKeyPool []string `mapstructure:"required-api-key-name-pool"`
}

type endpointConfiguration struct {
	Name   string                        `mapstructure:"name"`
	Values []*endpointValueConfiguration `mapstructure:"values"`
}

type endpointValueConfiguration struct {
	Name            string `mapstructure:"name"`
	SourceName      string `mapstructure:"source-name"`
	SourceValueName string `mapstructure:"source-value-name"`
}

type sourceConfiguration struct {
	Name                   string                      `mapstructure:"name"`
	Url                    string                      `mapstructure:"url"`
	CachingEnable          bool                        `mapstructure:"caching-enabled"`
	CachingLifeTimeSeconds int                         `mapstructure:"caching-life-time-seconds"`
	Retries                int                         `mapstructure:"retries-count"`
	HttpHeader             map[string]string           `mapstructure:"http-headers"`
	TimeoutSeconds         int                         `mapstructure:"timeout-seconds"`
	Values                 []*sourceValueConfiguration `mapstructure:"values"`
}

type sourceValueConfiguration struct {
	Name                 string `mapstructure:"name"`
	Xpath                string `mapstructure:"xpath"`
	ExtractionStrategy   string `mapstructure:"extraction-strategy"`
	ExtractionTrim       bool   `mapstructure:"extraction-trim"`
	ExtractionRegex      string `mapstructure:"extraction-regex"`
	ExtractionRegexIndex int    `mapstructure:"extraction-regex-match-index"`
	Type                 string `mapstructure:"type"`
}

func LoadServerConfigurationFromFile() (*ApiKitServerConfiguration, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("config: failed to access the current working directory path: %w", err)
	}

	viper.AddConfigPath(cwd)
	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: failed to read the config file: %w", err)
	}

	configuration := new(apiKitServerConfiguration)
	if err := viper.Unmarshal(&configuration); err != nil {
		return nil, fmt.Errorf("config: failed to unmarshal the config file content: %w", err)
	}

	if config, err := buildConfiguration(configuration); err != nil {
		return nil, fmt.Errorf("config: failed to build the configuration from file: %w", err)
	} else {
		return config, nil
	}
}

func buildConfiguration(c *apiKitServerConfiguration) (*ApiKitServerConfiguration, error) {
	config := &ApiKitServerConfiguration{
		ApiKit: &ApiKitConfiguration{
			Sources:   make([]*SourceConfiguration, 0, len(c.ApiKit.Sources)),
			Endpoints: make([]*EndpointConfiguration, 0, len(c.ApiKit.Endpoints)),
		},
		Endpoints:   make([]*ApiKitServerEndpointConfiguration, 0, len(c.Endpoints)),
		ApiKeys:     make([]*ApiKitServerKeyConfiguration, 0, len(c.ApiKeys)),
		VerboseMode: c.VerboseMode,
		Host:        c.Host,
	}

	for _, apiKey := range c.ApiKeys {
		config.ApiKeys = append(config.ApiKeys, &ApiKitServerKeyConfiguration{
			Name:   apiKey.Name,
			Secret: apiKey.Secret,
		})
	}

	for _, endpoint := range c.Endpoints {
		config.Endpoints = append(config.Endpoints, &ApiKitServerEndpointConfiguration{
			EndpointName:       endpoint.EndpointName,
			Path:               endpoint.Path,
			RequiredApiKeyPool: endpoint.RequiredApiKeyPool,
		})
	}

	for _, endpoint := range c.ApiKit.Endpoints {
		endpointValues := make([]*EndpointValueConfiguration, 0, len(endpoint.Values))
		for _, value := range endpoint.Values {
			endpointValues = append(endpointValues, &EndpointValueConfiguration{
				Name:            value.Name,
				SourceName:      value.SourceName,
				SourceValueName: value.SourceValueName,
			})
		}

		config.ApiKit.Endpoints = append(config.ApiKit.Endpoints, &EndpointConfiguration{
			Name:   endpoint.Name,
			Values: endpointValues,
		})
	}

	for _, source := range c.ApiKit.Sources {
		sourceValues := make([]*SourceValueConfiguration, 0, len(source.Values))
		for _, value := range source.Values {
			var extractionStrategy ExtractionStrategy
			switch strings.ToLower(value.ExtractionStrategy) {
			case "first":
				extractionStrategy = First
			case "single":
				extractionStrategy = Single
			default:
				return nil, fmt.Errorf("config: invalid extraction strategy for %s in %s", value.Name, source.Name)
			}

			var variableType VariableType
			switch strings.ToLower(value.Type) {
			case "string":
				variableType = String
			case "int":
				variableType = Int
			case "float":
				variableType = Float
			default:
				return nil, fmt.Errorf("config: invalid variable type for %s in %s", value.Name, source.Name)
			}

			sourceValues = append(sourceValues, &SourceValueConfiguration{
				Name:                 value.Name,
				Xpath:                value.Xpath,
				ExtractionStrategy:   extractionStrategy,
				ExtractionTrim:       value.ExtractionTrim,
				ExtractionRegex:      value.ExtractionRegex,
				ExtractionRegexIndex: value.ExtractionRegexIndex,
				Type:                 variableType,
			})
		}

		config.ApiKit.Sources = append(config.ApiKit.Sources, &SourceConfiguration{
			Name:                   source.Name,
			Url:                    source.Url,
			CachingEnable:          source.CachingEnable,
			CachingLifeTimeSeconds: source.CachingLifeTimeSeconds,
			Retries:                source.Retries,
			HttpHeaders:            source.HttpHeader,
			TimeoutSeconds:         source.TimeoutSeconds,
			Values:                 sourceValues,
		})
	}

	if valid, msg := ValidateServer(config); !valid {
		return nil, fmt.Errorf("config: validation failed due to %s", msg)
	} else {
		return config, nil
	}
}
