package amqp

import (
	logger "app/pkg/loggerutil"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

const defaultConsumerName = "cnsm1"

//Rabbit Client data
type RabbitMqClient struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
	key      string
	queue    string
}

//Rabbit Client interface
//Create connection, channel and declare a queue for send URL parse jobs
func (r *RabbitMqClient) Init(dialURL, exchange, routeKey, queue string, workers int) {
	var err error
	//Connect until rabbit mq come online
	for {
		r.conn, err = amqp.Dial(dialURL)
		if err != nil {
			logger.Log(fmt.Sprintf("Can't dial at %s: %s",dialURL,err.Error()), "error.log")
		}else{
			logger.Log("Successfully connected!", "error.log")
			break
		}
		time.Sleep(3 * time.Second)
	}

	//Create a channel
	r.ch, err = r.conn.Channel()
	if err != nil {
		r.failOnError(err, "Can't get channel")
	}

	err = r.ch.ExchangeDeclare(exchange, //exchange point name
		amqp.ExchangeDirect, //kind
		false,               //durable
		false,               //auto-delete  - deleted if no binded queues
		false,               //internal, not accept accept publishings
		false,               //noWait When true, declare without waiting for a confirmation from the server
		nil)
	if err != nil {
		r.failOnError(err, "Can't declare excange")
	}

	//Assign exchange point name
	r.exchange = exchange
	//Assign route key for publishing into exchange point
	r.key = routeKey

	//Delare a queue
	_, err = r.ch.QueueDeclare(queue, //queue name
		false, //durable
		false, //autoDelete
		false, //exclusive
		false, //no wait
		nil)
	if err != nil {
		r.failOnError(err, "Can't declare queue")
	}
	//Bind new queue to exchange point (direct) with route key
	err = r.ch.QueueBind(queue, //queue name
		routeKey, //routing key
		exchange, //exchange name
		false,    //noWait
		nil)
	if err != nil {
		r.failOnError(err, "Can't bind queue to exchange")
	}

	//Assign queue name , used by Consume later
	r.queue = queue

	// Set our quality of service.  Since we're sharing `workers` consumers on the same
	// channel, we want at least `workers` messages in flight.
	err = r.ch.Qos(workers, // prefetchCount
		0,     // prefetchSize
		false) // global

	if err != nil {
		r.failOnError(err, "Can't manage QOS")
	}
}

//Consume jobs
func (r *RabbitMqClient) ConsumeJson() <-chan amqp.Delivery {

	deliverChan, err := r.ch.Consume(
		r.queue,             //queue
		defaultConsumerName, //consumer name
		false,               //autoAck
		false,               //exclusive
		false,               //noLocal
		false,               //noWait
		nil)
	if err != nil {
		r.failOnError(err, "Can't consume")
	}
	return deliverChan
}

//Fail over routine -
func (r *RabbitMqClient) failOnError(err error, msg string) {
	if err != nil {
		logger.Log(fmt.Sprintf("%s: %s", msg, err), "error.log")
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

//Gracefully destroy Rabbit MQ  client
func (r *RabbitMqClient) Cancel() {
	//Cancel consumer
	r.ch.Cancel(defaultConsumerName,
		false /*noWait, if true do not wait for the server to acknowledge the cancel. */)
}

//Rabbit Client Destruct
func (r *RabbitMqClient) Destruct() {
	// Close Channel
	r.ch.Close()
	// Close Connection
	r.conn.Close()
}
