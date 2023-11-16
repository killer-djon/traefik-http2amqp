package traefik_http2amqp

import (
	"context"
	"log"
	"net/http"
)

// Config the plugin configuration.
type Config struct {
	Headers map[string]string `yaml:"headers"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Headers: nil,
	}
}

type Http2Amqp struct {
	next    http.Handler
	name    string
	headers map[string]string
}

// New created a new GeoBlock plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Http2Amqp{
		next:    next,
		name:    name,
		headers: nil,
	}, nil
}

// ServeHTTP method to skip at next request step
func (h *Http2Amqp) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.headers["auth"] = request.Header.Get("Authorization")
	log.Println("[AFS] Debug log for plugin with headers", h)
}
