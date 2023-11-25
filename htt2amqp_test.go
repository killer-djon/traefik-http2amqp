package traefik_http2amqp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	config := &Config{
		Host:               HOST,
		Port:               PORT,
		Vhost:              VHOST,
		Username:           "",
		Password:           "",
		HeaderExchangeName: "",
		HeaderQueueName:    "",
		HeaderExchangeType: EXCHANGE_TYPE,
	}
	newConfig := CreateConfig()

	if fmt.Sprint(config) != fmt.Sprint(newConfig) {
		t.Errorf("Created config struct is not equal as expected")
	}
}

func TestNew(t *testing.T) {
	ctx := context.Background()

	queueName := "test-http2amqp-queue"
	exchangeName := "test-http2amqp-exchange"

	name := "traefik-http2amqp"

	config := CreateConfig()
	config.Password = "guest"
	config.Username = "guest"
	config.HeaderQueueName = "X-QUEUE"
	config.HeaderExchangeName = "X-EXCHANGE"

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Header.Set(config.HeaderExchangeName, exchangeName)
		req.Header.Set(config.HeaderQueueName, queueName)

		return
	})

	handler, err := New(ctx, next, config, name)
	if err != nil {
		t.Errorf("Error when create plugin instance: %x", err)
	}

	jsonBody := []byte(`{"client_message": "hello, server!"}`)
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost", bodyReader)
	if err != nil {
		t.Fatal("Error when request to plugin", err)
	}

	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Result().StatusCode != http.StatusOK {
		t.Errorf("Request is bad with status code = %d", recorder.Result().StatusCode)
	}

	if req.Header.Get(config.HeaderExchangeName) == "" {
		t.Error("HeaderName for exchnage name is empty")
	}

	if req.Header.Get(config.HeaderQueueName) == "" {
		t.Error("HeaderName for queue name is empty")
	}
}
