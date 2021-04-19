package main

import (
	"fmt"
	"strconv"
)

var payload = "produce msg hello world"

func (c *Client) produce() {
	if token := c.conn.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	var pp string
	for i := 0; i < *num; i++ {
		fmt.Println("---- doing publish ----")
		pp = payload + strconv.Itoa(i)
		token := c.conn.Publish(*topic, byte(*qos), false, pp)
		token.Wait()
	}

	c.conn.Disconnect(250)
	fmt.Println("Sample Publisher Disconnected")
}
