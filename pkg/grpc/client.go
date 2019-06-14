// Package grpc
package grpc

import (
	"context"
	pb "github.com/kzozulya1/webpage-word-freq-counter-protobuf/protobuf"
	logger "app/pkg/loggerutil"
	"google.golang.org/grpc"
	"os"
)

//Main routine
func SendPageWordFrequencyDocument(pageWordFreq *pb.PageWordFrequency) error {
	//Remove nil elements
	sanitizeNilElements(pageWordFreq)

	// Set up a connection to the server.
	//Is taken from docker-composer.yml file
	address := os.Getenv("GRPC_SERVICE_ADDRESS")
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		logger.Log(err.Error(), "error.log")
		return err
	}
	defer conn.Close()

	//Create new client instance
	client := pb.NewWordFrequencyServiceClient(conn)

	_, err = client.UpdateOrCreatePageWordFrequency(context.Background(), pageWordFreq)
	if err != nil {
		logger.Log(err.Error(), "error.log")
		return err
	}
	return nil
}

//Remove nil elements
func sanitizeNilElements(pageWordFreq *pb.PageWordFrequency) {
	for i, v := range pageWordFreq.Words {
		if v == nil {
			//Remove nil element from slice
			pageWordFreq.Words = append(pageWordFreq.Words[:i], pageWordFreq.Words[i+1:]...)
		}
	}
}
