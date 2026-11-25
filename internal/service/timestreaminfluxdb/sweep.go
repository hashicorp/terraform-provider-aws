// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_timestreaminfluxdb_db_backup", sweepDBBackups)
	awsv2.Register("aws_timestreaminfluxdb_db_cluster", sweepDBClusters)
	awsv2.Register("aws_timestreaminfluxdb_db_instance", sweepDBInstances)
}

func sweepDBBackups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.TimestreamInfluxDBClient(ctx)
	var input timestreaminfluxdb.ListDbBackupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := timestreaminfluxdb.NewListDbBackupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			// Only on-demand backups are managed by aws_timestreaminfluxdb_db_backup.
			if v.Type != awstypes.DbBackupTypeOnDemand || v.Status == awstypes.DbBackupStatusDeleting {
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newDBBackupResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}

func sweepDBClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.TimestreamInfluxDBClient(ctx)
	var input timestreaminfluxdb.ListDbClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := timestreaminfluxdb.NewListDbClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDBClusterResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}

func sweepDBInstances(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.TimestreamInfluxDBClient(ctx)
	var input timestreaminfluxdb.ListDbInstancesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := timestreaminfluxdb.NewListDbInstancesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDBInstanceResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}
