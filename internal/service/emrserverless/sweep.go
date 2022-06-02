//go:build sweep
// +build sweep

package emrserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_emrserverless_application", &resource.Sweeper{
		Name: "aws_emrserverless_application",
		F:    sweepApplications,
	})
}

func sweepApplications(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EMRServerlessConn
	input := &emrserverless.ListApplicationsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListApplicationsPages(input, func(page *emrserverless.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Applications {
			if aws.StringValue(v.State) == emrserverless.ApplicationStateTerminated {
				continue
			}

			r := ResourceApplication()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EMR Serverless Application sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EMR Serverless Applications (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EMR Serverless Applications (%s): %w", region, err)
	}

	return nil
}
