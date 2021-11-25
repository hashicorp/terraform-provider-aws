//go:build sweep
// +build sweep

package cloud9

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloud9_environment_ec2", &resource.Sweeper{
		Name: "aws_cloud9_environment_ec2",
		F:    sweepEnvironmentEC2s,
	})
}

func sweepEnvironmentEC2s(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloud9conn
	sweepResources := make([]*testSweepResource, 0)
	var sweeperErrs *multierror.Error

	input := &cloud9.ListEnvironmentsInput{}
	err = conn.ListEnvironmentsPages(input, func(page *cloud9.ListEnvironmentsOutput, lastPage bool) bool {
		if len(page.EnvironmentIds) == 0 {
			log.Printf("[INFO] No Cloud9 Environment EC2s to sweep")
			return false
		}
		for _, envID := range page.EnvironmentIds {
			id := aws.StringValue(envID)

			log.Printf("[INFO] Deleting Cloud9 Environment EC2: %s", id)
			r := ResourceEnvironmentEC2()
			d := r.Data(nil)
			d.SetId(id)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}
		return !lastPage
	})

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Cloud9 Environment EC2s: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Cloud9 Environment EC2 for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Cloud9 Environment EC2s for %s: %s", region, errs)
		return nil
	}

	return sweeperErrs.ErrorOrNil()
}
