package traefik_http2amqp

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/wagslane/go-rabbitmq"
	"io"
	"log"
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
		Host:     HOST,
		Port:     PORT,
		Vhost:    VHOST,
		Username: "",
		Password: "",
	}
}

type Http2Amqp struct {
	next      http.Handler
	name      string
	config    *Config
	publisher *rabbitmq.Publisher
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

	conn, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:%d%s", config.Username, config.Password, config.Host, config.Port, config.Vhost),
		rabbitmq.WithConnectionOptionsLogging,
	)

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
	)

	if err != nil {
		return nil, err
	}
	fmt.Println("[Http2Amqp] make new publisher instance")

	defer conn.Close()
	if err != nil {
		return nil, err
	}

	return &Http2Amqp{
		next:      next,
		name:      name,
		config:    config,
		publisher: publisher,
	}, nil
}

// ServeHTTP method to skip at next request step
func (h *Http2Amqp) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		if request.Header.Get(h.config.HeaderExchangeName) != "" && request.Header.Get(h.config.HeaderQueueName) != "" {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				writer.Write([]byte("Bad body request"))
				return
			}

			fmt.Println("[Http2Amqp] published body to exchange", string(body))
			err = h.publisher.Publish(
				body,
				[]string{h.config.HeaderQueueName},
				rabbitmq.WithPublishOptionsContentType("application/json"),
				rabbitmq.WithPublishOptionsExchange(h.config.HeaderExchangeName),
				rabbitmq.WithPublishOptionsCorrelationID(uuid.NewV4().String()),
				rabbitmq.WithPublishOptionsPersistentDelivery,
				/*rabbitmq.WithPublishOptionsHeaders(rabbitmq.Table{

				})*/
			)

			if err != nil {
				log.Println(err)
			}

			fmt.Println("[Http2Amqp] body was published")
		}

		fmt.Println("[Http2Amqp] Method is not a post")
	}

	fmt.Println("[Http2Amqp] Skip publishing to rabbit, its simple request")
	h.next.ServeHTTP(writer, request)
}
