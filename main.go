package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tnaucoin/goawaste/internal"
)

type Event struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, event *Event) (*string, error) {
	if event == nil {
		return nil, fmt.Errorf("recieved nil event")
	}
	availableVolumes := internal.GetAvailableVolumes()
	for _, volume := range availableVolumes {
		fmt.Println(volume)
	}
	return nil, nil
}

func main() {
	lambda.Start(HandleRequest)
}
