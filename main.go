package traefik_http2amqp

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	rmq "github.com/wagslane/go-rabbitmq"
	"io"
	"net/http"
)

const (
	HOST  = "localhost"
	PORT  = 5672
	VHOST = "/"
)

// Config the plugin configuration.
type Config struct {
	Host               string `yaml:"host" json:"host"`
	Port               int    `yaml:"port" json:"port"`
	Vhost              string `yaml:"vhost" json:"vhost"`
	Username           string `json:"username" yaml:"username"`
	Password           string `yaml:"password" json:"password"`
	HeaderExchangeName string `yaml:"headerExchangeName" json:"headerExchangeName"`
	HeaderQueueName    string `yaml:"headerQueueName" json:"headerQueueName"`
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
	}
}

type Http2Amqp struct {
	next       http.Handler
	name       string
	config     *Config
	connection *rmq.Conn
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

	conn, err := rmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:%d%s", config.Username, config.Password, config.Host, config.Port, config.Vhost),
		rmq.WithConnectionOptionsLogging,
	)

	defer conn.Close()
	if err != nil {
		return nil, err
	}

	return &Http2Amqp{
		next:       next,
		name:       name,
		config:     config,
		connection: conn,
	}, nil
}

// ServeHTTP method to skip at next request step
func (h *Http2Amqp) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		if request.Header.Get(h.config.HeaderExchangeName) != "" && request.Header.Get(h.config.HeaderQueueName) != "" {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(400)
				writer.Write([]byte("Bad body request to trasport amqp"))
				return
			}

			publisher, err := rmq.NewPublisher(
				h.connection,
				rmq.WithPublisherOptionsLogging,
			)

			if err != nil {
				fmt.Println("[Http2Amqp] error occurred to establish connection to amqp")
				writer.WriteHeader(400)
				writer.Write([]byte("[Http2Amqp] error occurred to establish connection to amqp"))
				return
			}
			fmt.Println("[Http2Amqp] make new publisher instance")

			defer publisher.Close()
			fmt.Println("[Http2Amqp] published body to exchange", string(body))

			err = publisher.Publish(
				body,
				[]string{h.config.HeaderQueueName},
				rmq.WithPublishOptionsContentType("application/json"),
				rmq.WithPublishOptionsExchange(h.config.HeaderExchangeName),
				rmq.WithPublishOptionsCorrelationID(uuid.NewV4().String()),
				rmq.WithPublishOptionsPersistentDelivery,
				/*rabbitmq.WithPublishOptionsHeaders(rabbitmq.Table{

				})*/
			)

			if err != nil {
				fmt.Println("[Http2Amqp] error publisher", err)
				writer.WriteHeader(400)
				writer.Write([]byte("[Http2Amqp] error publisher"))
				return
			}

			fmt.Println("[Http2Amqp] body was published")
		}

		fmt.Println("[Http2Amqp] Method is not a post")
	}

	fmt.Println("[Http2Amqp] Skip publishing to rabbit, its simple request")
	h.next.ServeHTTP(writer, request)
}
