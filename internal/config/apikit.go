package config

import (
	"net/url"
	"regexp"

	"github.com/Krzysztofz01/apikit/internal/utils"
)

func Validate(c *ApiKitConfiguration) (bool, string) {
	return c.isValid()
}

type ApiKitConfiguration struct {
	Sources   []*SourceConfiguration
	Endpoints []*EndpointConfiguration
}

func (c *ApiKitConfiguration) isValid() (bool, string) {
	if c.Sources == nil {
		return false, "uninitialized sources collection"
	}

	if c.Endpoints == nil {
		return false, "uninitialized endpoints collection"
	}

	sourcesValues := make(map[string]utils.Set[string], len(c.Sources))
	for _, source := range c.Sources {
		// NOTE: Inner source config values validation
		if valid, msg := source.isValid(); !valid {
			return false, msg
		}

		// NOTE: Map sourcesValues and value names unique validation
		sourceValues := utils.NewEmptySet[string]()
		for _, value := range source.Values {
			// NOTE: Inner source value config values validation
			if valid, msg := value.isValid(); !valid {
				return false, msg
			}

			if !sourceValues.Add(value.Name) {
				return false, "duplicate source value name found"
			}
		}

		// NOTE: Map sourcesValues and source names unique validation
		if _, exist := sourcesValues[source.Name]; exist {
			return false, "duplicate source name found"
		} else {
			sourcesValues[source.Name] = sourceValues
		}
	}

	endpointsValues := make(map[string]utils.Set[string], len(c.Endpoints))
	for _, endpoint := range c.Endpoints {
		// NOTE: Inner endpoint config values validation
		if valid, msg := endpoint.isValid(); !valid {
			return false, msg
		}

		// NOTE: Map endpointsValues and value names unique validation
		endpointValues := utils.NewEmptySet[string]()
		for _, value := range endpoint.Values {
			// NOTE: Inner endpoint value config values check
			if valid, msg := value.isValid(); !valid {
				return false, msg
			}

			if !endpointValues.Add(value.Name) {
				return false, "duplicate endpoint value name found"
			}

			// NOTE: Endpoint source name check
			targetSource, targetSourceExist := sourcesValues[value.SourceName]
			if !targetSourceExist {
				return false, "endpoint references non existing source"
			}

			// NOTE: Endpoint source value name check
			if !targetSource.Contains(value.SourceValueName) {
				return false, "endpoint references non existing source value"
			}
		}

		// NOTE: Map endpointsValues and endpoint names unique validation
		if _, exist := endpointsValues[endpoint.Name]; exist {
			return false, "duplicate endpoint name found"
		} else {
			sourcesValues[endpoint.Name] = endpointValues
		}
	}

	return true, ""
}

type EndpointConfiguration struct {
	Name   string
	Values []*EndpointValueConfiguration
}

func (c *EndpointConfiguration) isValid() (bool, string) {
	if len(c.Name) == 0 {
		return false, "invalid endpoint name"
	}

	return true, ""
}

type EndpointValueConfiguration struct {
	Name            string
	SourceName      string
	SourceValueName string
}

func (c *EndpointValueConfiguration) isValid() (bool, string) {
	if len(c.Name) == 0 {
		return false, "invalid endpoint value name"
	}

	if len(c.SourceName) == 0 {
		return false, "invalid endpoint value source name"
	}

	if len(c.SourceValueName) == 0 {
		return false, "invalid endpoint value source value name"
	}

	return true, ""
}

type SourceConfiguration struct {
	Name                   string
	Url                    string
	CachingEnable          bool
	CachingLifeTimeSeconds int
	Retries                int
	HttpHeaders            map[string]string
	TimeoutSeconds         int
	Values                 []*SourceValueConfiguration
}

func (c *SourceConfiguration) isValid() (bool, string) {
	if len(c.Name) == 0 {
		return false, "invalid source name"
	}

	if _, err := url.Parse(c.Url); err != nil {
		return false, "invalid source url"
	}

	if c.CachingLifeTimeSeconds < 0 {
		return false, "invalid caching life time that is out of range"
	}

	if c.Retries < 0 {
		return false, "invalid retries count that is out of range"
	}

	if c.TimeoutSeconds < 0 {
		return false, "invalid timeout seconds that is out of range"
	}

	return true, ""
}

type VariableType int

const (
	String VariableType = iota
	Int
	Float
)

type ExtractionStrategy int

const (
	First ExtractionStrategy = iota
	Single
)

type SourceValueConfiguration struct {
	Name                 string
	Xpath                string
	ExtractionStrategy   ExtractionStrategy
	ExtractionTrim       bool
	ExtractionRegex      string
	ExtractionRegexIndex int
	Type                 VariableType
}

func (c *SourceValueConfiguration) isValid() (bool, string) {
	if len(c.Name) == 0 {
		return false, "invalid source value name"
	}

	if len(c.Xpath) == 0 {
		return false, "invalid xpath value"
	}

	if _, err := regexp.Compile(c.ExtractionRegex); err != nil {
		return false, "invalid regex that could not be parsed"
	}

	if c.ExtractionRegexIndex < 0 {
		return false, "invalid regex match index that is out of range"
	}

	return true, ""
}
