package internal

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"log"
)

func GetAvailableVolumes() []string {
	var availableVolumes []string
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	svc := ec2.NewFromConfig(cfg)
	filter := types.Filter{
		Name:   aws.String("status"),
		Values: []string{"available"},
	}
	input := ec2.DescribeVolumesInput{
		Filters:    []types.Filter{filter},
		MaxResults: aws.Int32(500),
	}
	result, err := svc.DescribeVolumes(context.TODO(), &input)
	if err != nil {
		panic("failed to describe volumes, " + err.Error())
	}
	log.Println("Checking for available volumes...")
	if len(result.Volumes) > 0 {
		for _, volume := range result.Volumes {
			if len(volume.Attachments) == 0 {
				availableVolumes = append(availableVolumes, *volume.VolumeId)
			}
		}
		log.Println("Checking for additional volumes...")
		for result.NextToken != nil {
			input.NextToken = result.NextToken
			result, err = svc.DescribeVolumes(context.TODO(), &input)
			if err != nil {
				panic("failed to describe volumes, " + err.Error())
			}
			if len(result.Volumes) > 0 {
				for _, volume := range result.Volumes {
					if len(volume.Attachments) == 0 {
						availableVolumes = append(availableVolumes, *volume.VolumeId)
					}
				}
			}
		}
	}
	return availableVolumes
}
