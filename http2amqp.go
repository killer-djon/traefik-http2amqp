package traefik_http2amqp

import (
	"context"
	"fmt"
	rmq "github.com/killer-djon/rabbitmq-go"
	uuid "github.com/satori/go.uuid"
	"io"
	"net/http"
	"time"
)

const (
	HOST          = "localhost"
	PORT          = 5672
	VHOST         = "/"
	EXCHANGE_TYPE = "direct"
)

// Config the plugin configuration.
type Config struct {
	Host               string `yaml:"host,omitempty" json:"host,omitempty" toml:"host,omitempty"`
	Port               int    `yaml:"port,omitempty" json:"port,omitempty" toml:"port,omitempty"`
	Vhost              string `yaml:"vhost,omitempty" json:"vhost,omitempty" toml:"vhost,omitempty"`
	Username           string `json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty"`
	Password           string `yaml:"password,omitempty" json:"password,omitempty" toml:"password,omitempty"`
	HeaderExchangeName string `yaml:"headerExchangeName,omitempty" json:"headerExchangeName,omitempty" toml:"headerExchangeName,omitempty"`
	HeaderQueueName    string `yaml:"headerQueueName,omitempty" json:"headerQueueName,omitempty" toml:"headerQueueName,omitempty"`
	HeaderExchangeType string `yaml:"headerExchangeType,omitempty" json:"headerExchangeType,omitempty" toml:"headerExchangeType,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Host:               HOST,
		Port:               PORT,
		Vhost:              VHOST,
		Username:           "",
		Password:           "",
		HeaderExchangeName: "",
		HeaderQueueName:    "",
		HeaderExchangeType: EXCHANGE_TYPE,
	}
}

type Http2Amqp struct {
	next       http.Handler
	name       string
	config     *Config
	connection *rmq.Connection
}

// New created a new GeoBlock plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if config.Username == "" {
		return nil, fmt.Errorf("[Http2Amqp] Username is empty must be set to connect to RabbitMQ")
	}

	if config.Password == "" {
		return nil, fmt.Errorf("[Http2Amqp] Password is empty must be set to connect to RabbitMQ")
	}

	if config.HeaderQueueName == "" || config.HeaderExchangeName == "" {
		return nil, fmt.Errorf("[Http2Amqp] you must set queueName and exchangeName to publish message into")
	}

	connection, err := rmq.DialConfig(
		fmt.Sprintf("amqp://%s:%s@%s:%d%s", config.Username, config.Password, config.Host, config.Port, config.Vhost),
		rmq.Config{
			Heartbeat: 30 * time.Second,
		})

	if err != nil {
		return nil, fmt.Errorf("[Http2Amqp] connection error try to check you connection")
	}

	return &Http2Amqp{
		next:       next,
		name:       name,
		config:     config,
		connection: connection,
	}, nil
}

// ServeHTTP method to skip at next request step
func (h *Http2Amqp) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPost {
		if request.Header.Get(h.config.HeaderExchangeName) != "" && request.Header.Get(h.config.HeaderQueueName) != "" {
			queueName := request.Header.Get(h.config.HeaderQueueName)
			exchangeName := request.Header.Get(h.config.HeaderExchangeName)

			exchangeType := request.Header.Get(h.config.HeaderExchangeType)
			if exchangeType == "" {
				exchangeType = EXCHANGE_TYPE
			}

			body, err := io.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(400)
				writer.Write([]byte("Bad body request to transport amqp"))
				return
			}

			channel, err := h.connection.Channel()
			if err != nil {
				fmt.Println("[Http2Amqp] error occurred to get channel from connection", err)

				writer.WriteHeader(400)
				writer.Write([]byte("[Http2Amqp] error occurred to get channel from connection"))
				return
			}

			defer channel.Close()

			err = channel.ExchangeDeclare(
				exchangeName,
				exchangeType,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				fmt.Printf(
					"[Http2Amqp] error occurred for exchange type declaration, type: %s, name: %s",
					h.config.HeaderExchangeType,
					h.config.HeaderExchangeName,
				)

				writer.WriteHeader(400)
				writer.Write([]byte("[Http2Amqp] error occurred for exchange type declaration"))
				return
			}

			queue, err := channel.QueueDeclare(
				queueName,
				true,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				fmt.Printf(
					"[Http2Amqp] error occurred for queue declaration, name: %s",
					h.config.HeaderQueueName,
				)

				writer.WriteHeader(400)
				writer.Write([]byte("[Http2Amqp] error occurred for queue declaration"))
				return
			}

			err = channel.QueueBind(queue.Name, queueName, exchangeName, false, nil)
			if err != nil {
				fmt.Printf(
					"[Http2Amqp] error occurred when bind queue to exchange, queue: %s, exchange: %s",
					h.config.HeaderQueueName,
					h.config.HeaderExchangeName,
				)

				writer.WriteHeader(400)
				writer.Write([]byte("[Http2Amqp] error occurred for queue declaration"))
				return
			}
			correlationId := uuid.NewV4().String()

			msg := rmq.Publishing{
				Headers: rmq.Table{
					"Correlation-id": correlationId,
				},
				ContentType:   "application/json",
				DeliveryMode:  rmq.Persistent,
				CorrelationId: correlationId,
				Timestamp:     time.Time{},
				UserId:        request.Header.Get("userId"),
				Body:          body,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			err = channel.PublishWithContext(
				ctx,
				exchangeName,
				queue.Name,
				false,
				false,
				msg,
			)

			if err != nil {
				fmt.Println("[Http2Amqp] error publishing body to amqp", err)
				writer.WriteHeader(400)
				writer.Write([]byte("[Http2Amqp] error publishing body to amqp"))

				return
			}

			fmt.Println("[Http2Amqp] body was published")
		}

		fmt.Println("[Http2Amqp] Method is not a post")
	}

	fmt.Println("[Http2Amqp] Skip publishing to rabbit, its simple request")
	h.next.ServeHTTP(writer, request)
}
