### About plugin
This traefik plugin make the transport to resend http requests body directly to amqp message broker. 
### Configure (default)
    Host               string `yaml:"host,omitempty" json:"host,omitempty" toml:"host,omitempty"`
	Port               int    `yaml:"port,omitempty" json:"port,omitempty" toml:"port,omitempty"`
	Vhost              string `yaml:"vhost,omitempty" json:"vhost,omitempty" toml:"vhost,omitempty"`
	Username           string `json:"username,omitempty" yaml:"username,omitempty" toml:"username,omitempty"`
	Password           string `yaml:"password,omitempty" json:"password,omitempty" toml:"password,omitempty"`
	HeaderExchangeName string `yaml:"headerExchangeName,omitempty" json:"headerExchangeName,omitempty" toml:"headerExchangeName,omitempty"`
	HeaderQueueName    string `yaml:"headerQueueName,omitempty" json:"headerQueueName,omitempty" toml:"headerQueueName,omitempty"`
	HeaderExchangeType string `yaml:"headerExchangeType,omitempty" json:"headerExchangeType,omitempty" toml:"headerExchangeType,omitempty"`

```yaml
...
    http2amqp:
      host: 'localhost' # Main host for connection
      port: 5672 # Main port for connection
      vhost: '/' # Main vhost for rabbitmq instance
      username: 'guest' # Rabbitmq instance username
      password: 'guest' # Rabbitmq instance password
      headerExchangeName: 'X-EXCHANGE' # A name of the exchange to publish on
      headerQueueName: 'X-QUEUE' # A name of the queue to publish on
      headerExchangeType: 'X-TYPE' # If is empty default type is direct
...
```
If send the request at post method and if into headers presents this names like headerExchangeName value and headerQueueName value, then body should be parsed and restranslit to amqp message broker
```shell
curl -XPOST https://<some domain>/<some api uri> \
  -H 'X-EXCHANGE: <exchange name to publish on>' \
  -H 'X-QUEUE: <queue name to publish on>' \
  -H 'X-TYPE: <exchange type>' \
  -d '{"key": "value"}'
```
Because in the header we have exchange name and queue name the body will be parsed and transmit to amqp

### Install
For install this middleware plugin you can configure them like this

#### Configure traefik-ingress for kubernates
First step is to add experimental plugin
```yaml
# anywhere/traefik.yml
experimental:
  plugins:
    http2amqp:
      moduleName: github.com/killer-djon/traefik-http2amqp
      version: v1.0.0 # Current version you could view into repository
```
Or you can set as localPlugin instance like this:
```yaml
experimental:
  localPlugins:
    http2amqp:
      moduleName: github.com/killer-djon/traefik-http2amqp
```
Next step is to add customDefenition for middleware
```yaml
...
-   apiVersion: traefik.containo.us/v1alpha1
    kind: Middleware
    metadata:
      name: my-traefik-http2amqp
      namespace: default # you could rename default namespace
    spec:
      plugin:
        http2amqp:
          host: 'localhost' # Main host for connection
          port: 5672 # Main port for connection
          vhost: '/' # Main vhost for rabbitmq instance
          username: 'guest' # Rabbitmq instance username
          password: 'guest' # Rabbitmq instance password
          headerExchangeName: 'X-EXCHANGE' # A name of the exchange to publish on
          headerQueueName: 'X-QUEUE' # A name of the queue to publish on
          headerExchangeType: 'X-TYPE'
...
```
And as the third step is to add ingress annotation for this middleware (if need to be at single middleware for single service) or set this plugin for avery services as entryPoint (web, websecure ...)
1. At ingress annotation
```yaml
...
traefik.ingress.kubernetes.io/router.middlewares: default-my-traefik-http2amqp@kubernetescrd
# in this case you installed plugin will works just for this service
```
2. At general entrypoint of traefik instance (web, websecure ...)
```yaml
ports:
  ...
  websecure:
    port: 8443
    expose: true
    exposedPort: 443
    protocol: TCP
    appProtocol: https
    middlewares:
    - default-my-traefik-http2amqp@kubernetescrd
# in this case you installed plugin will works for every request
```