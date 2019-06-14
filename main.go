//Package main.go
package main

import (
	rmq "app/pkg/amqp"
	entity "app/pkg/entity"
	logger "app/pkg/loggerutil"
	"github.com/streadway/amqp"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

//Main routine
func main() {
	log.Printf("Starting consumer...")
	//Config
	url, exchange, routeKey, queue, workers := getConfig()

	//Init Rabbit MQ Consumer client
	rabbitClient := rmq.RabbitMqClient{}
	rabbitClient.Init(url, exchange, routeKey, queue, workers)
	defer rabbitClient.Destruct()

	//Declare delivery <-chan
	var delivery <-chan amqp.Delivery

	for {
		//Start consuming
		delivery = rabbitClient.ConsumeJson()

		//Start batch of go routines - all they share messages from `delivery` channel from consume
		for i := 0; i < workers; i++ {
			go entity.Worker(delivery, i)
		}
		//After 10s
		time.Sleep(300 * time.Second)
		//...cancel `delivery` channel, it causes finish the range with `delivery` in all goroutines and gracefully end them
		rabbitClient.Cancel()
		//Waiting for a while before next sprint
		time.Sleep(2 * time.Second)
	}
}

//Get Rabbit MQ client configuration
func getConfig() (string, string, string, string, int) {
	url := os.Getenv("RMQ_URL")
	if url == "" {
		panic("RabbitMQ URL is not specified")
	}
	exchange := os.Getenv("RMQ_EXCHANGE_NAME")
	if exchange == "" {
		panic("RabbitMQ exchange name is not specified")
	}
	routeKey := os.Getenv("RMQ_ROUTE_KEY")
	if routeKey == "" {
		panic("RabbitMQ route key is not specified")
	}
	queue := os.Getenv("RMQ_QUEUE_NAME")
	if queue == "" {
		panic("RabbitMQ queue name is not specified")
	}

	var workers int = runtime.NumCPU()
	if workersEnv := os.Getenv("WORKERS"); workersEnv != "" {
		workersEnvInt, err := strconv.Atoi(workersEnv)
		if err != nil {
			logger.Log(err, "Can't convert WORKERS env val to int")
			workers = 3
		} else {
			workers = workersEnvInt
		}
	}
	return url, exchange, routeKey, queue, workers
}
