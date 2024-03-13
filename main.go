package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/tnaucoin/goawaste/internal"
	"github.com/tnaucoin/goawaste/models"
	"log"
)

// https://github.com/karthickcse05/aws_unused_resources/blob/master/Lambda/src/aws_resources.py

func printResults(findings []models.Finding) {
	for _, finding := range findings {
		message := fmt.Sprintf("Resource ID: %s for service: %s, reason: %s", finding.Id, finding.Type, finding.Reason)
		fmt.Println(message)
	}
}

func HandleRequest(ctx context.Context, event *models.Event) (*string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	deadline, _ := ctx.Deadline()
	log.Println("Setting parent context to control all child contexts...")
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	noRetention := internal.GetLogsWithNoRetention(ctx, cfg)
	if len(noRetention) > 0 {
		printResults(noRetention)
	}
	availVolumes := internal.GetAvailableVolumes(ctx, cfg)
	if len(availVolumes) > 0 {
		printResults(availVolumes)
	}
	notAssociatedEIPs := internal.GetNotAssociatedEIPs(ctx, cfg)
	if len(notAssociatedEIPs) > 0 {
		printResults(notAssociatedEIPs)
	}
	unusedRDSSnapshots := internal.GetUnusedRDSSnapshots(ctx, cfg)
	if len(unusedRDSSnapshots) > 0 {
		printResults(unusedRDSSnapshots)
	}
	unusedEBSSnapshots := internal.GetUnusedEBSSnapshots(ctx, cfg, event)
	if len(unusedEBSSnapshots) > 0 {
		printResults(unusedEBSSnapshots)
	}
	return nil, nil
}

func main() {
	lambda.Start(HandleRequest)
}
