package main

import (
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func (c *Client) Connect() {
	token := c.conn.Connect()
	ok := token.WaitTimeout(c.timeout)
	if !ok {
		panic("connect token timeout")
	}
	if err := token.Error(); err != nil {
		panic(err)
	}
}

func (c *Client) initOptions() {
	c.opts.AddBroker(*broker)
	c.opts.SetClientID(*id)
	c.opts.SetUsername(*user)
	c.opts.SetPassword(*password)
	c.opts.SetKeepAlive(4 * time.Second)
	//c.opts.SetCleanSession(*cleansess)
	c.opts.SetStore(MQTT.NewFileStore(c.storePath))
	//c.opts.SetOnConnectHandler(c.onConnect)
	c.opts.SetDefaultPublishHandler(c.onPublish)
	return
}

type Client struct {
	conn      MQTT.Client
	opts      *MQTT.ClientOptions
	id        string
	timeout   time.Duration
	storePath string
	msgchan   chan MQTT.Message
}

func NewClient() *Client {
	c := &Client{
		opts:    MQTT.NewClientOptions(),
		timeout: 60 * time.Second,
	}
	c.initOptions()
	c.conn = MQTT.NewClient(c.opts)
	c.msgchan = make(chan MQTT.Message)
	return c
}
