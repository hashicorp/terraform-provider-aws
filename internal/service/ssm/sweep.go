//go:build sweep
// +build sweep

package ssm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_ssm_maintenance_window", &resource.Sweeper{
		Name: "aws_ssm_maintenance_window",
		F:    sweepMaintenanceWindows,
	})

	resource.AddTestSweepers("aws_ssm_resource_data_sync", &resource.Sweeper{
		Name: "aws_ssm_resource_data_sync",
		F:    sweepResourceDataSyncs,
	})
}

func sweepMaintenanceWindows(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).SSMConn
	input := &ssm.DescribeMaintenanceWindowsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeMaintenanceWindows(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSM Maintenance Window sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving SSM Maintenance Windows: %s", err)
		}

		for _, window := range output.WindowIdentities {
			id := aws.StringValue(window.WindowId)
			input := &ssm.DeleteMaintenanceWindowInput{
				WindowId: window.WindowId,
			}

			log.Printf("[INFO] Deleting SSM Maintenance Window: %s", id)

			_, err := conn.DeleteMaintenanceWindow(input)

			if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SSM Maintenance Window (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepResourceDataSyncs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).SSMConn
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &ssm.ListResourceDataSyncInput{}

	err = conn.ListResourceDataSyncPages(input, func(page *ssm.ListResourceDataSyncOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, resourceDataSync := range page.ResourceDataSyncItems {
			r := ResourceResourceDataSync()
			d := r.Data(nil)

			d.SetId(aws.StringValue(resourceDataSync.SyncName))
			d.Set("name", resourceDataSync.SyncName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing SSM Resource Data Sync for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping SSM Resource Data Sync for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping SSM Resource Data Sync sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
