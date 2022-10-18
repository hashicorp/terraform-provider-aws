//go:build sweep
// +build sweep

package dms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_dms_replication_instance", &resource.Sweeper{
		Name: "aws_dms_replication_instance",
		F:    sweepReplicationInstances,
		Dependencies: []string{
			"aws_dms_replication_task",
		},
	})

	resource.AddTestSweepers("aws_dms_replication_task", &resource.Sweeper{
		Name: "aws_dms_replication_task",
		F:    sweepReplicationTasks,
	})
}

func sweepReplicationInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).DMSConn
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeReplicationInstancesPages(&dms.DescribeReplicationInstancesInput{}, func(page *dms.DescribeReplicationInstancesOutput, lastPage bool) bool {
		for _, instance := range page.ReplicationInstances {
			r := ResourceReplicationInstance()
			d := r.Data(nil)
			d.Set("replication_instance_arn", instance.ReplicationInstanceArn)
			d.SetId(aws.StringValue(instance.ReplicationInstanceIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing DMS Replication Instances: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DMS Replication Instances for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DMS Replication Instance sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepReplicationTasks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).DMSConn
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &dms.DescribeReplicationTasksInput{
		WithoutSettings: aws.Bool(true),
	}
	err = conn.DescribeReplicationTasksPages(input, func(page *dms.DescribeReplicationTasksOutput, lastPage bool) bool {
		for _, instance := range page.ReplicationTasks {
			r := ResourceReplicationTask()
			d := r.Data(nil)
			d.SetId(aws.StringValue(instance.ReplicationTaskIdentifier))
			d.Set("replication_task_arn", instance.ReplicationTaskArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing DMS Replication Tasks: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DMS Replication Tasks for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DMS Replication Instance sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
