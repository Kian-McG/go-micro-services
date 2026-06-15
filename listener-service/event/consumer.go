package event

import (
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn     *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection, queueName string) (Consumer, error) {
	consumer := Consumer{
		conn:     conn,
		queueName: queueName,
	}
	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}
	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	q, err := declareRandomQueue(channel)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		err := channel.QueueBind(
			q.Name,
			topic,
			"logs_topic",
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	messages, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for msg := range messages {
			var payload Payload
			_ = json.Unmarshal(msg.Body, &payload)
			go handlePayload(payload)
		}
	}()
	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			fmt.Printf("Error handling payload: %v\n", err)
		}
	
	case "auth":
		// Handle auth payload
	default:
		err := logEvent(payload)
		if err != nil {
			fmt.Printf("Error handling payload: %v\n", err)
		}
	}
}

func logEvent(entry Payload) error {
		jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to log event: %v", err)
	}

	return nil
}
