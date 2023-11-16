package traefik_http2amqp

import (
	"context"
	"log"
	"net/http"
)

// Config the plugin configuration.
type Config struct {
	Headers []string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers: []string{},
	}
}

type Http2Amqp struct {
	next    http.Handler
	name    string
	headers []string
}

// New created a new GeoBlock plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Http2Amqp{
		next:    next,
		name:    name,
		headers: config.Headers,
	}, nil
}

// ServeHTTP method to skip at next request step
func (h *Http2Amqp) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var headersMap = make(map[string]bool)
	for _, item := range h.headers {
		headersMap[item] = false

		if request.Header.Get(item) != "" {
			headersMap[item] = true
		}
	}

	log.Println("[AFS] Debug log for plugin with headers", headersMap)
}
