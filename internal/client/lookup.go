package client

import (
	"fmt"

	"github.com/Krzysztofz01/apikit/internal/config"
)

type EndpointLookup interface {
	GetEndpointSourcesWithSourceValueNames(endpointName string) (map[string][]string, error)
	GetEndpointValueName(sourceName, sourceValueName string) (string, error)
}

type endpointLookup struct {
	endpointSourcesWithSourceValuesLookup    map[string]map[string][]string
	sourceValueNameToEndpointValueNameLookup map[string]string
}

func (l endpointLookup) GetEndpointSourcesWithSourceValueNames(endpointName string) (map[string][]string, error) {
	if endpointSourcesWithSourceValuesNames, ok := l.endpointSourcesWithSourceValuesLookup[endpointName]; !ok {
		return nil, fmt.Errorf("client: specified endpoint not present in lookup")
	} else {
		return endpointSourcesWithSourceValuesNames, nil
	}
}

func (l endpointLookup) GetEndpointValueName(sourceName string, sourceValueName string) (string, error) {
	lookupKey := fmt.Sprintf("%s%s%s", sourceName, lookupKeySeparator, sourceValueName)

	if endpointValueName, ok := l.sourceValueNameToEndpointValueNameLookup[lookupKey]; !ok {
		return "", fmt.Errorf("client: specified pair of source and value are not matching and endpoint value")
	} else {
		return endpointValueName, nil
	}
}

const (
	lookupKeySeparator = ";"
)

func createEndpointLookup(c *config.ApiKitConfiguration) (EndpointLookup, error) {
	var (
		endpointSourcesWithSourceValuesLookup    = make(map[string]map[string][]string, len(c.Endpoints))
		sourceValueNameToEndpointValueNameLookup = make(map[string]string)
	)

	for _, endpointConfig := range c.Endpoints {
		sourcesWithSourceValuesLookup := make(map[string][]string)
		for _, endpointValueConfig := range endpointConfig.Values {
			// NOTE: Part related to "endpointSourcesWithSourceValuesLookup"
			if _, ok := sourcesWithSourceValuesLookup[endpointValueConfig.SourceName]; !ok {
				sourcesWithSourceValuesLookup[endpointValueConfig.SourceName] = []string{endpointValueConfig.SourceValueName}
			} else {
				sourcesWithSourceValuesLookup[endpointValueConfig.SourceName] = append(sourcesWithSourceValuesLookup[endpointValueConfig.SourceName], endpointValueConfig.SourceValueName)
			}

			// NOTE: Part related to "sourceValueNameToEndpointValueNameLookup"
			lookupKey := fmt.Sprintf("%s%s%s", endpointValueConfig.SourceName, lookupKeySeparator, endpointValueConfig.SourceValueName)
			sourceValueNameToEndpointValueNameLookup[lookupKey] = endpointValueConfig.Name
		}

		endpointSourcesWithSourceValuesLookup[endpointConfig.Name] = sourcesWithSourceValuesLookup
	}

	return endpointLookup{
		endpointSourcesWithSourceValuesLookup:    endpointSourcesWithSourceValuesLookup,
		sourceValueNameToEndpointValueNameLookup: sourceValueNameToEndpointValueNameLookup,
	}, nil
}
