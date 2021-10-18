//go:build sweep
// +build sweep

package s3control

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_s3_access_point", &resource.Sweeper{
		Name: "aws_s3_access_point",
		F:    sweepAccessPoints,
	})

	resource.AddTestSweepers("aws_s3_multi_region_access_point", &resource.Sweeper{
		Name: "aws_s3_multi_region_access_point",
		F:    sweepMultiRegionAccessPoints,
	})
}

func sweepAccessPoints(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	accountId := client.(*conns.AWSClient).AccountID
	conn := client.(*conns.AWSClient).S3ControlConn

	input := &s3control.ListAccessPointsInput{
		AccountId: aws.String(accountId),
	}
	var sweeperErrs *multierror.Error

	err = conn.ListAccessPointsPages(input, func(page *s3control.ListAccessPointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, accessPoint := range page.AccessPointList {
			input := &s3control.DeleteAccessPointInput{
				AccountId: aws.String(accountId),
				Name:      accessPoint.Name,
			}
			name := aws.StringValue(accessPoint.Name)

			log.Printf("[INFO] Deleting S3 Access Point: %s", name)
			_, err := conn.DeleteAccessPoint(input)

			if tfawserr.ErrMessageContains(err, "NoSuchAccessPoint", "") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting S3 Access Point (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Access Point sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Access Points: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepMultiRegionAccessPoints(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	if client.(*AWSClient).region != endpoints.UsWest2RegionID {
		log.Printf("[WARN] Skipping sweep for region: %s", client.(*AWSClient).region)
		return nil
	}

	accountId := client.(*AWSClient).accountid
	conn := client.(*AWSClient).s3controlconn

	input := &s3control.ListMultiRegionAccessPointsInput{
		AccountId: aws.String(accountId),
	}
	var sweeperErrs *multierror.Error

	err = conn.ListMultiRegionAccessPointsPages(input, func(page *s3control.ListMultiRegionAccessPointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, multiRegionAccessPoint := range page.AccessPoints {
			input := &s3control.DeleteMultiRegionAccessPointInput{
				AccountId: aws.String(accountId),
				Details: &s3control.DeleteMultiRegionAccessPointInput_{
					Name: multiRegionAccessPoint.Name,
				},
			}

			name := aws.StringValue(multiRegionAccessPoint.Name)

			log.Printf("[INFO] Deleting S3 Multi-Region Access Point: %s", name)
			_, err := conn.DeleteMultiRegionAccessPoint(input)

			if tfawserr.ErrCodeEquals(err, tfs3control.ErrCodeNoSuchMultiRegionAccessPoint) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting S3 Multi-Region Access Point (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Multi-Region Access Point sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Multi-Region Access Points: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}
