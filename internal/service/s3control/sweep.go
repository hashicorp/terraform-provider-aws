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
