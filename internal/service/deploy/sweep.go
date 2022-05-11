//go:build sweep
// +build sweep

package deploy

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codedeploy"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_codedeploy_app", &resource.Sweeper{
		Name: "aws_codedeploy_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).DeployConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &codedeploy.ListApplicationsInput{}

	err = conn.ListApplicationsPages(input, func(page *codedeploy.ListApplicationsOutput, lastPage bool) bool {
		for _, app := range page.Applications {
			if app == nil {
				continue
			}

			appName := aws.StringValue(app)
			r := ResourceApp()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", "xxxx", appName))
			d.Set("name", appName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing CodeDeploy Applications for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping CodeDeploy Applications for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping CodeDeploy Applications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
