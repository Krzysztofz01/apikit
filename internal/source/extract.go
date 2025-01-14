package source

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Krzysztofz01/apikit/internal/config"
	"github.com/Krzysztofz01/apikit/internal/constants"
	"github.com/Krzysztofz01/apikit/internal/log"
)

func GetHtmlViaHttp(h *http.Client, ctx context.Context, cfg *config.SourceConfiguration, l log.Loggerp) (string, error) {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, cfg.Url, nil)
	if err != nil {
		return "", fmt.Errorf("source: failed to create the extraction http request: %w", err)
	}

	header := make(http.Header, len(cfg.HttpHeaders)+1)
	header.Set("user-agent", fmt.Sprintf("ApiKit/%s", constants.Version))

	for key, value := range cfg.HttpHeaders {
		if len(key) == 0 {
			return "", fmt.Errorf("source: header with invalid key provided")
		}

		if len(value) == 0 {
			return "", fmt.Errorf("source: header with invalid value provided")
		}

		header.Set(key, value)
	}

	request.Header = header

	var (
		attemptsLeft int            = cfg.Retries + 1
		response     *http.Response = nil
		requestErr   error          = nil
	)

	for {
		if attemptsLeft <= 0 {
			return "", fmt.Errorf("source: extraction http request retries count exceeded")
		} else {
			attemptsLeft -= 1
		}

		if response, requestErr = h.Do(request); requestErr == nil && response.StatusCode == http.StatusOK {
			defer func() {
				if err := response.Body.Close(); err != nil {
					l.Warnf("Failed to close the response body with error: %s", err)
				}
			}()

			break
		} else {
			statusCode := "\"none\""
			if response != nil {
				statusCode = strconv.Itoa(response.StatusCode)
			}

			l.Warnf("Request attempt %d of %d failed with code %s: %s", cfg.Retries+1-attemptsLeft, cfg.Retries+1, statusCode, err)
		}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("source: failed to read the response body content: %w", err)
	}

	if len(body) == 0 {
		return "", fmt.Errorf("source: the response body is empty")
	}

	contentEncoding := response.Header.Get("Content-Encoding")
	l.Debugf("Response content encoding %s", contentEncoding)

	decodedBody, err := GetDecodedHttpBody(body, contentEncoding)
	if err != nil {
		return "", fmt.Errorf("source: failed to decode the response body: %w", err)
	}

	return decodedBody, nil
}

func GetDecodedHttpBody(body []byte, contentEncoding string) (string, error) {
	switch contentEncoding {
	case "":
		{
			return string(body), nil
		}
	case "gzip":
		{
			bodyReader := bytes.NewReader(body)

			gzipReader, err := gzip.NewReader(bodyReader)
			if err != nil {
				return "", fmt.Errorf("source: failed to create the gzip reader: %w", err)
			}

			defer gzipReader.Close()

			if decoded, err := io.ReadAll(gzipReader); err != nil {
				return "", fmt.Errorf("source: failed to read the decoded gzip body: %w", err)
			} else {
				return string(decoded), nil
			}
		}
	case "deflate":
		{
			bodyReader := bytes.NewReader(body)

			flateReader := flate.NewReader(bodyReader)

			defer flateReader.Close()

			if decoded, err := io.ReadAll(flateReader); err != nil {
				return "", fmt.Errorf("source: failed to read the decoded deflate body: %w", err)
			} else {
				return string(decoded), nil
			}
		}
	default:
		{
			return "", fmt.Errorf("source: unsupported content encoding")
		}
	}
}
