package main

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func (c *Client) consume() {
	receiveCount := 0
	if token := c.conn.Subscribe(*topic, byte(*qos), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	for receiveCount < *num {
		incoming := <-c.msgchan
		fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", incoming.Topic(), incoming.Payload())
		receiveCount++
	}

	c.conn.Disconnect(250)
	fmt.Println("Sample Subscriber Disconnected")
}

func (c *Client) onPublish(_ mqtt.Client, m mqtt.Message) {
	c.msgchan <- m
}
