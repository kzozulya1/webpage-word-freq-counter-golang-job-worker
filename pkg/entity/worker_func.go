package entity

import (
	logger "app/pkg/loggerutil"
	grpc "app/pkg/grpc"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
)

//Go routine Worker func
func Worker(works <-chan amqp.Delivery, workerNum int) {
	var job Job
	for work := range works {
		//Get body and write to log
		//Decode posted job
		err := json.Unmarshal(work.Body, &job)
		if err != nil {
			logger.Log(err.Error(), "error.log")
			fmt.Printf("Error: %s", err.Error())
			continue
		} else {
			crawler := &CrawlAgent{}
			pageWordFreq, err := crawler.Process(&job)
			if err != nil {
				logger.Log(err.Error(), "error.log")
				fmt.Printf("Error: %s", err.Error())
				continue
			}

			err = grpc.SendPageWordFrequencyDocument(pageWordFreq)
			if err != nil {
				logger.Log( fmt.Sprintf("Error send gRPC msg: %s", err.Error()), "error.log")
				fmt.Printf("Error send gRPC msg: %s", err.Error())
				continue
			}
		}
		//Eventualy mark delivery as processed
		work.Ack(false)
	}
}
