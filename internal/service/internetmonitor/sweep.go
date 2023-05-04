//go:build sweep
// +build sweep

package internetmonitor

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/internetmonitor"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_internetmonitor_monitor", &resource.Sweeper{
		Name: "aws_internetmonitor_monitor",
		F:    sweepMonitors,
	})
}

func sweepMonitors(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).InternetMonitorConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.ListMonitorsPagesWithContext(ctx, &internetmonitor.ListMonitorsInput{}, func(resp *internetmonitor.ListMonitorsOutput, lastPage bool) bool {
		if len(resp.Monitors) == 0 {
			log.Print("[DEBUG] No InternetMonitor Monitors to sweep")
			return !lastPage
		}

		for _, c := range resp.Monitors {
			r := ResourceMonitor()
			d := r.Data(nil)
			d.SetId(aws.StringValue(c.MonitorName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing InternetMonitor Monitors: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping InternetMonitor Monitors for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping InternetMonitor Monitor sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
