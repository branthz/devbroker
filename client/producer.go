package main

import (
	"fmt"
)

func (c *Client) produce() {
	if token := c.conn.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Sample Publisher Started")
	for i := 0; i < *num; i++ {
		fmt.Println("---- doing publish ----")
		token := c.conn.Publish(*topic, byte(*qos), false, *payload)
		token.Wait()
	}

	c.conn.Disconnect(250)
	fmt.Println("Sample Publisher Disconnected")
}
