package main

import (
	"broker/event"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {

	// try to connect to rabbitmq
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()  

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	// define http server
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
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

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJSON(w, err)
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "Event logged successfully via RabbitMQ"
	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error  {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data:  msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}