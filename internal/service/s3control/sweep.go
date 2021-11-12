//go:build sweep
// +build sweep

package s3control

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_s3_access_point", &resource.Sweeper{
		Name: "aws_s3_access_point",
		F:    sweepAccessPoints,
	})

	resource.AddTestSweepers("aws_s3control_multi_region_access_point", &resource.Sweeper{
		Name: "aws_s3control_multi_region_access_point",
		F:    sweepMultiRegionAccessPoints,
	})
}

func sweepAccessPoints(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).S3ControlConn
	accountID := client.(*conns.AWSClient).AccountID
	input := &s3control.ListAccessPointsInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListAccessPointsPages(input, func(page *s3control.ListAccessPointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, accessPoint := range page.AccessPointList {
			r := ResourceAccessPoint()
			d := r.Data(nil)
			d.SetId(AccessPointCreateResourceID(aws.StringValue(accessPoint.AccessPointArn), accountID, aws.StringValue(accessPoint.Name)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Access Point sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing SS3 Access Points (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Access Points (%s): %w", region, err)
	}

	return nil
}

func sweepMultiRegionAccessPoints(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	if region != endpoints.UsWest2RegionID {
		log.Printf("[WARN] Skipping S3 Multi-Region Access Point sweep for region: %s", region)
		return nil
	}
	conn := client.(*conns.AWSClient).S3ControlConn
	accountID := client.(*conns.AWSClient).AccountID
	input := &s3control.ListMultiRegionAccessPointsInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListMultiRegionAccessPointsPages(input, func(page *s3control.ListMultiRegionAccessPointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, accessPoint := range page.AccessPoints {
			r := ResourceMultiRegionAccessPoint()
			d := r.Data(nil)
			d.SetId(MultiRegionAccessPointCreateResourceID(accountID, aws.StringValue(accessPoint.Name)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Multi-Region Access Point sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Multi-Region Access Points (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping S3 Multi-Region Access Points (%s): %w", region, err)
	}

	return nil
}
