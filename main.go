package traefik_http2amqp

import (
	"context"
	"log"
	"net/http"
)

// Config the plugin configuration.
type Config struct {
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type Http2Amqp struct {
	next http.Handler
	name string
}

// New created a new GeoBlock plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &Http2Amqp{
		next: next,
		name: name,
	}, nil
}

// ServeHTTP method to skip at next request step
func (h *Http2Amqp) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("[AFS] Debug log for plugin with headers")
}
