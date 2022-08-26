//go:build sweep
// +build sweep

package rum

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_rum_app_monitor", &resource.Sweeper{
		Name: "aws_rum_app_monitor",
		F:    sweepAppMonitors,
	})
}

func sweepAppMonitors(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RUMConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.ListAppMonitorsPages(&cloudwatchrum.ListAppMonitorsInput{}, func(resp *cloudwatchrum.ListAppMonitorsOutput, lastPage bool) bool {
		if len(resp.AppMonitorSummaries) == 0 {
			log.Print("[DEBUG] No RUM App Monitors to sweep")
			return !lastPage
		}

		for _, c := range resp.AppMonitorSummaries {
			r := ResourceAppMonitor()
			d := r.Data(nil)
			d.SetId(aws.StringValue(c.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing RUM App Monitors: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping RUM App Monitors for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping RUM App Monitor sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
