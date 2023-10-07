package mq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func GetChannel() (*amqp.Channel, func(), bool) {
	// setup RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Printf("failed to connect to RabbitMQ. %s\n", err)
		return nil, func() {}, false
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("failed to open a RabbitMQ channel. %s\n", err)
		return nil, func() {}, false
	}

	return ch, func() {
		log.Println("closing RabbitMQ channel")
		if err := ch.Close(); err != nil {
			fmt.Println(err.Error())
		}

		log.Println("closing RabbitMQ connection")
		if err := conn.Close(); err != nil {
			fmt.Println(err.Error())
		}
	}, true
}
