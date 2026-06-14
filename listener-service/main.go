package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()  

	// start listening for messages

	// create a consumer

	// watch the queue and consume the events
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err != nil {
			fmt.Printf("RabbitMQ is not yet ready...")
			counts++
		}else {
			connection = conn
			break
		}
		if counts > 5 {
			fmt.Printf("Giving up on RabbitMQ...")
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Printf("Retrying in %v...", backOff)
		time.Sleep(backOff)
		continue
	}
	return connection, nil
}