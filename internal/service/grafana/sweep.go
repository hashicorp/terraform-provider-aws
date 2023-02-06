//go:build sweep
// +build sweep

package grafana

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_grafana_workspace", &resource.Sweeper{
		Name: "aws_grafana_workspace",
		F:    sweepWorkSpaces,
	})
}

func sweepWorkSpaces(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GrafanaConn()

	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &managedgrafana.ListWorkspacesInput{}

	err = conn.ListWorkspacesPagesWithContext(ctx, input, func(page *managedgrafana.ListWorkspacesOutput, lastPage bool) bool {
		if len(page.Workspaces) == 0 {
			log.Printf("[INFO] No Grafana Workspaces to sweep")
			return false
		}
		for _, workspace := range page.Workspaces {

			id := aws.StringValue(workspace.Id)
			log.Printf("[INFO] Deleting Grafana Workspace: %s", id)
			r := ResourceWorkspace()
			d := r.Data(nil)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("reading Grafana Workspace (%s): %w", id, err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing Grafana Workspace for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Grafana Workspace for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Grafana Workspace sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
