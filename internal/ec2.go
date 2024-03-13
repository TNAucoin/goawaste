package internal

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/tnaucoin/goawaste/models"
	"log"
)

func GetAvailableVolumes(ctx context.Context, cfg aws.Config) []models.Finding {
	var availableVolumes []models.Finding
	svc := ec2.NewFromConfig(cfg)
	filter := types.Filter{
		Name:   aws.String("status"),
		Values: []string{"available"},
	}
	input := ec2.DescribeVolumesInput{
		Filters:    []types.Filter{filter},
		MaxResults: aws.Int32(500),
	}
	result, err := svc.DescribeVolumes(ctx, &input)
	if err != nil {
		panic("failed to describe volumes, " + err.Error())
	}
	if len(result.Volumes) > 0 {
		for _, volume := range result.Volumes {
			if len(volume.Attachments) == 0 {
				availableVolumes = append(availableVolumes, models.Finding{
					Id:     *volume.VolumeId,
					Type:   "AWS::EC2::Volume",
					Reason: "Volume is not attached to an instance",
				})
			}
		}
		for result.NextToken != nil {
			input.NextToken = result.NextToken
			result, err = svc.DescribeVolumes(context.TODO(), &input)
			if err != nil {
				panic("failed to describe volumes, " + err.Error())
			}
			if len(result.Volumes) > 0 {
				for _, volume := range result.Volumes {
					if len(volume.Attachments) == 0 {
						availableVolumes = append(availableVolumes, models.Finding{
							Id:     *volume.VolumeId,
							Type:   "AWS::EC2::Volume",
							Reason: "Volume is not attached to an instance",
						})
					}
				}
			}
		}
	}
	return availableVolumes
}

// TODO: figure out timing on the snapshots, and ownership.
// https://medium.com/@NickHystax/reduce-your-aws-bill-by-cleaning-orphaned-and-unused-disk-snapshots-c3142d6ab84
func GetUnusedEBSSnapshots(ctx context.Context, cfg aws.Config) []models.Finding {
	svc := ec2.NewFromConfig(cfg)
	var unusedEBSSnapshots []models.Finding
	var snapshotIds []string
	resp, err := svc.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
		MaxResults: aws.Int32(10),
		OwnerIds: []string{
			"self",
			"113191093292",
		},
	})
	if err != nil {
		panic("failed to describe snapshots, " + err.Error())
	}
	if len(resp.Snapshots) > 0 {
		for _, snapshot := range resp.Snapshots {
			snapshotIds = append(snapshotIds, *snapshot.SnapshotId)

		}
		for resp.NextToken != nil {
			resp, err = svc.DescribeSnapshots(context.TODO(), &ec2.DescribeSnapshotsInput{NextToken: resp.NextToken})
			if err != nil {
				panic("failed to describe snapshots, " + err.Error())
			}
			if len(resp.Snapshots) > 0 {
				snapshotIds = append(snapshotIds, *resp.Snapshots[0].SnapshotId)

			}
		}
		//check if the snapshots are attached to a volume.
		input := ec2.DescribeVolumesInput{
			Filters: []types.Filter{{
				Name:   aws.String("snapshot-id"),
				Values: snapshotIds,
			},
			},
		}
		result, err := svc.DescribeVolumes(ctx, &input)
		if err != nil {
			panic("failed to describe volumes, " + err.Error())
		}
		if len(result.Volumes) > 0 {
			for _, volume := range result.Volumes {
				//remove the snapshots that are attached to a volume
				for index, snapshot := range snapshotIds {
					if *volume.SnapshotId == snapshot {
						snapshotIds = append(snapshotIds[:index], snapshotIds[index+1:]...)
					}
				}

			}
		}
		// for each snapshot left in the list, generate findings
		for _, snapshotId := range snapshotIds {
			unusedEBSSnapshots = append(unusedEBSSnapshots, models.Finding{
				Id:     snapshotId,
				Type:   "AWS::EC2::Snapshot",
				Reason: "Snapshot is not attached to a volume",
			})
		}
	}
	return unusedEBSSnapshots
}

func GetLogsWithNoRetention(ctx context.Context, cfg aws.Config) []models.Finding {
	svc := cloudwatchlogs.NewFromConfig(cfg)
	logGroups := []string{"/aws", "API-Gateway", "RDSOSMetrics", "test", "/ecs"}
	log.Println("Checking for log groups without retention...", logGroups)
	var logsWithNoRetention []models.Finding
	for _, groupName := range logGroups {
		resp, err := svc.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: aws.String(groupName),
		})
		if err != nil {
			panic("failed to describe log groups, " + err.Error())
		}
		if len(resp.LogGroups) > 0 {
			for _, logGroup := range resp.LogGroups {
				if logGroup.RetentionInDays == nil {
					logsWithNoRetention = append(logsWithNoRetention, models.Finding{
						Id: *logGroup.LogGroupName, Type: "AWS::Logs::LogGroup", Reason: "No retention days specified",
					})
				}
			}
		}
	}
	return logsWithNoRetention
}

func GetNotAssociatedEIPs(ctx context.Context, cfg aws.Config) []models.Finding {
	svc := ec2.NewFromConfig(cfg)
	var notAssociatedEIPs []models.Finding
	resp, err := svc.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		panic("failed to describe addresses, " + err.Error())
	}
	if len(resp.Addresses) > 0 {
		for _, address := range resp.Addresses {
			if address.AssociationId == nil {
				notAssociatedEIPs = append(notAssociatedEIPs, models.Finding{
					Id: *address.AllocationId, Type: "AWS::EC2::EIP", Reason: "Not associated with an instance",
				})
			}
		}
	}
	return notAssociatedEIPs
}
