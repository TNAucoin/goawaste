package internal

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/tnaucoin/goawaste/models"
	"time"
)

// TODO: Figure out the time difference between the creation of the snapshot and the deletion of the instance.
func GetUnusedRDSSnapshots(ctx context.Context, cfg aws.Config) []models.Finding {
	svc := rds.NewFromConfig(cfg)
	var unusedSnapshots []models.Finding
	resp, err := svc.DescribeDBClusterSnapshots(ctx, &rds.DescribeDBClusterSnapshotsInput{
		MaxRecords: aws.Int32(100),
	})
	if err != nil {
		panic("failed to describe db cluster snapshots, " + err.Error())
	}
	if len(resp.DBClusterSnapshots) > 0 {
		for _, snapshot := range resp.DBClusterSnapshots {
			if snapshot.SnapshotCreateTime.Before(time.Now().AddDate(0, 0, -7)) {
				unusedSnapshots = append(unusedSnapshots, models.Finding{Id: *snapshot.DBClusterSnapshotIdentifier, Type: "RDS"})
			}
		}
		for resp.Marker != nil {
			resp, err = svc.DescribeDBClusterSnapshots(ctx, &rds.DescribeDBClusterSnapshotsInput{Marker: resp.Marker})
			if err != nil {
				panic("failed to describe db cluster snapshots, " + err.Error())
			}
			if len(resp.DBClusterSnapshots) > 0 {
				for _, snapshot := range resp.DBClusterSnapshots {
					if snapshot.SnapshotCreateTime.Before(time.Now().AddDate(0, 0, -7)) {
						unusedSnapshots = append(unusedSnapshots, models.Finding{Id: *snapshot.DBClusterSnapshotIdentifier, Type: "RDS"})
					}
				}
			}
		}
	}
	return unusedSnapshots
}
