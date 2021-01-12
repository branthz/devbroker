package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

/*
Options:
 [-help]                      Display help
 [-a pub|sub]                 Action pub (publish) or sub (subscribe)
 [-m <message>]               Payload to send
 [-n <number>]                Number of messages to send or receive
 [-q 0|1|2]                   Quality of Service
 [-clean]                     CleanSession (true if -clean is present)
 [-id <clientid>]             CliendID
 [-user <user>]               User
 [-password <password>]       Password
 [-broker <uri>]              Broker URI
 [-topic <topic>]             Topic
*/

var (
	topic     = flag.String("topic", "", "The topic name to/from which to publish/subscribe")
	broker    = flag.String("broker", "tcp://127.0.0.1:9898", "The broker URI. ex: tcp://10.10.1.1:1883")
	password  = flag.String("password", "be fool", "The password (optional)")
	user      = flag.String("user", "qiongshi", "The User (optional)")
	id        = flag.String("id", "testgoid", "The ClientID (optional)")
	cleansess = flag.Bool("clean", false, "Set Clean Session (default false)")
	qos       = flag.Int("qos", 0, "The Quality of Service 0,1,2 (default 0)")
	num       = flag.Int("num", 1, "The number of messages to publish or subscribe (default 1)")
	payload   = flag.String("message", "first hand", "The message text to publish (default empty)")
	action    = flag.String("action", "", "Action publish or subscribe (required)")
)

func main() {
	flag.Parse()
	lg := log.New(os.Stdout, "client: ", log.Lshortfile)
	MQTT.DEBUG = lg
	MQTT.CRITICAL = lg
	MQTT.ERROR = lg

	if *action != "pub" && *action != "sub" {
		fmt.Println("Invalid setting for -action, must be pub or sub")
		return
	}

	if *topic == "" {
		fmt.Println("Invalid setting for -topic, must not be empty")
		return
	}

	fmt.Printf("Sample Info:\n")
	fmt.Printf("\taction:    %s\n", *action)
	fmt.Printf("\tbroker:    %s\n", *broker)
	fmt.Printf("\tclientid:  %s\n", *id)
	fmt.Printf("\tuser:      %s\n", *user)
	fmt.Printf("\tpassword:  %s\n", *password)
	fmt.Printf("\ttopic:     %s\n", *topic)
	fmt.Printf("\tmessage:   %s\n", *payload)
	fmt.Printf("\tqos:       %d\n", *qos)
	fmt.Printf("\tcleansess: %v\n", *cleansess)
	fmt.Printf("\tnum:       %d\n", *num)

	cli := NewClient()
	cli.Connect()
	if *action == "pub" {
		cli.produce()
	} else {
		cli.consume()
	}
}
