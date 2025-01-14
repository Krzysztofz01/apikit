package source

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Krzysztofz01/apikit/internal/config"
	"github.com/Krzysztofz01/apikit/internal/content"
	"github.com/Krzysztofz01/apikit/internal/log"
	"github.com/Krzysztofz01/apikit/internal/utils"
)

type Source interface {
	GetValue(key string) (interface{}, error)
	GetValues(keys []string) (map[string]interface{}, error)
}

type source struct {
	httpClient       *http.Client
	htmlContentCache utils.Cacheable[content.HtmlContent]
	valueKeys        map[string]bool
	valueRegex       map[string]*regexp.Regexp
	logger           log.Loggerp
	cfg              *config.SourceConfiguration
	mu               sync.Mutex
}

func CreateSource(h *http.Client, c *config.SourceConfiguration, l log.Logger) (Source, error) {
	if h == nil {
		return nil, fmt.Errorf("source: provided http client reference is nil")
	}

	if c == nil {
		return nil, fmt.Errorf("source: provided config reference is nil")
	}

	if l == nil {
		return nil, fmt.Errorf("source: provided logger reference is nil")
	}

	prefix := fmt.Sprintf("Source - %s", c.Name)
	logger := log.CreatePrefixedLogger(prefix, l)

	valueKeys := make(map[string]bool, len(c.Values))
	for _, sourceValue := range c.Values {
		valueKeys[sourceValue.Name] = true
	}

	valueRegex := make(map[string]*regexp.Regexp, len(c.Values))
	for _, sourceValue := range c.Values {
		if len(sourceValue.ExtractionRegex) == 0 {
			continue
		}

		if regex, err := regexp.Compile(sourceValue.ExtractionRegex); err != nil {
			return nil, fmt.Errorf("source: failed to compile the source value extraction regex: %w", err)
		} else {
			valueRegex[sourceValue.ExtractionRegex] = regex
		}
	}

	return &source{
		httpClient:       h,
		htmlContentCache: utils.NewCacheable[content.HtmlContent](),
		valueKeys:        valueKeys,
		valueRegex:       valueRegex,
		logger:           logger,
		cfg:              c,
		mu:               sync.Mutex{},
	}, nil
}

func (s *source) GetValue(key string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if values, err := s.GetValuesNoLock(key); err != nil {
		return nil, fmt.Errorf("source: failed to access the source values: %w", err)
	} else {
		return values[key], nil
	}
}

func (s *source) GetValues(keys []string) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if values, err := s.GetValuesNoLock(keys...); err != nil {
		return nil, fmt.Errorf("source: failed to access the source values: %w", err)
	} else {
		return values, nil
	}
}

func (s *source) GetValuesNoLock(keys ...string) (map[string]interface{}, error) {
	if !s.AreValueKeysValid(keys...) {
		return nil, fmt.Errorf("source: invalid values keys provided")
	}

	htmlContent, err := s.GetHtmlContent()
	if err != nil {
		return nil, fmt.Errorf("source: failed to access the html content: %w", err)
	}

	result := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		if value, err := s.GetSourceValue(key, htmlContent); err != nil {
			return nil, fmt.Errorf("source: failed to access the source value: %w", err)
		} else {
			result[key] = value
		}
	}

	return result, nil
}

func (s *source) AreValueKeysValid(keys ...string) bool {
	if !utils.IsDistinct(keys) {
		return false
	}

	for _, key := range keys {
		if _, ok := s.valueKeys[key]; !ok {
			return false
		}
	}

	return true
}

func (s *source) GetHtmlContent() (content.HtmlContent, error) {
	if html, ok := s.htmlContentCache.Get(); ok && html != nil {
		s.logger.Infof("Cached content used to resolve %s access", s.cfg.Url)
		return html, nil
	}

	html, err := GetHtmlViaHttp(s.httpClient, context.Background(), s.cfg, s.logger)
	if err != nil {
		return nil, fmt.Errorf("source: failed to access html content via http: %w", err)
	}

	htmlContent, err := content.CreateHtmlContent(html)
	if err != nil {
		return nil, fmt.Errorf("source: failed to create the html content: %w", err)
	}

	if s.cfg.CachingEnable {
		ttl := time.Duration(s.cfg.CachingLifeTimeSeconds) * time.Second

		s.htmlContentCache.SetWithTTL(htmlContent, ttl)
	}

	s.logger.Infof("Request to resource made to resolve %s access", s.cfg.Url)
	return htmlContent, nil
}

func (s *source) GetSourceValue(key string, html content.HtmlContent) (interface{}, error) {
	var sourceValueConfig *config.SourceValueConfiguration = nil
	for _, config := range s.cfg.Values {
		if config.Name == key {
			sourceValueConfig = config
		}
	}

	if sourceValueConfig == nil {
		return nil, fmt.Errorf("source: failed to access the target source value configuration")
	}

	var sourceValueElement content.HtmlContentElement
	switch sourceValueConfig.ExtractionStrategy {
	case config.First:
		{
			if element, found, err := html.GetFirstElement(sourceValueConfig.Xpath); err != nil {
				return nil, fmt.Errorf("source: failed to extract first element via xpath: %w", err)
			} else if !found {
				return nil, fmt.Errorf("source: target first element to extract not found: %w", err)
			} else {
				sourceValueElement = element
			}
		}
	case config.Single:
		{
			if element, found, err := html.GetSingleElement(sourceValueConfig.Xpath); err != nil {
				return nil, fmt.Errorf("source: failed to extract single element via xpath: %w", err)
			} else if !found {
				return nil, fmt.Errorf("source: target single element to extract not found: %w", err)
			} else {
				sourceValueElement = element
			}
		}
	default:
		return nil, fmt.Errorf("source: invliad extraction strategy specified")
	}

	var sourceValuePreprocess content.HtmlContentValuePreprocess = func(in string) (string, error) {
		if sourceValueConfig.ExtractionTrim {
			in = strings.TrimSpace(in)
		}

		if regex, ok := s.valueRegex[sourceValueConfig.ExtractionRegex]; ok {
			matches := regex.FindStringSubmatch(in)
			s.logger.Debugf("Regex \"%s\" matching result for value %s: %+v", sourceValueConfig.ExtractionRegex, key, matches)

			if sourceValueConfig.ExtractionRegexIndex >= len(matches) {
				return "", fmt.Errorf("source: source value extraction regex index out of matches range")
			}

			return matches[sourceValueConfig.ExtractionRegexIndex], nil
		}

		return in, nil
	}

	var sourceValue interface{}
	switch sourceValueConfig.Type {
	case config.Float:
		{
			if value, err := sourceValueElement.GetInnerTextFloat(sourceValuePreprocess); err != nil {
				return nil, fmt.Errorf("source: failed to get inner text float value: %w", err)
			} else {
				sourceValue = value
			}
		}
	case config.Int:
		{
			if value, err := sourceValueElement.GetInnerTextInt(sourceValuePreprocess); err != nil {
				return nil, fmt.Errorf("source: failed to get inner text int value: %w", err)
			} else {
				sourceValue = value
			}
		}
	case config.String:
		{
			if value, err := sourceValueElement.GetInnerTextString(sourceValuePreprocess); err != nil {
				return nil, fmt.Errorf("source: failed to get inner text string value: %w", err)
			} else {
				sourceValue = value
			}
		}
	default:
		{
			return nil, fmt.Errorf("source: invalid source value type specified")
		}
	}

	return sourceValue, nil
}
