package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type RabbitMQ struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		connection: conn,
		channel:    ch,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		r.connection.Close()
		return err
	}
	return r.connection.Close()
}

func (r *RabbitMQ) GetChannel() *amqp.Channel {
	return r.channel
}
