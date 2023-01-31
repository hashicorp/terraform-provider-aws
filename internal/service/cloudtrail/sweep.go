//go:build sweep
// +build sweep

package cloudtrail

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudtrail", &resource.Sweeper{
		Name: "aws_cloudtrail",
		F:    sweeps,
	})
}

func sweeps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudTrailConn()
	var sweeperErrs *multierror.Error

	err = conn.ListTrailsPagesWithContext(ctx, &cloudtrail.ListTrailsInput{}, func(page *cloudtrail.ListTrailsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, trail := range page.Trails {
			name := aws.StringValue(trail.Name)

			if name == "AWSMacieTrail-DO-NOT-EDIT" {
				log.Printf("[INFO] Skipping AWSMacieTrail-DO-NOT-EDIT for Macie Classic, which is not automatically recreated by the service")
				continue
			}

			output, err := conn.DescribeTrailsWithContext(ctx, &cloudtrail.DescribeTrailsInput{
				TrailNameList: aws.StringSlice([]string{name}),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error describing CloudTrail (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if len(output.TrailList) == 0 {
				log.Printf("[INFO] CloudTrail (%s) not found, skipping", name)
				continue
			}

			if aws.BoolValue(output.TrailList[0].IsOrganizationTrail) {
				log.Printf("[INFO] CloudTrail (%s) is an organization trail, skipping", name)
				continue
			}

			log.Printf("[INFO] Deleting CloudTrail: %s", name)
			_, err = conn.DeleteTrailWithContext(ctx, &cloudtrail.DeleteTrailInput{
				Name: aws.String(name),
			})
			if tfawserr.ErrCodeEquals(err, cloudtrail.ErrCodeTrailNotFoundException) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudTrail (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudTrail sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudTrails: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
