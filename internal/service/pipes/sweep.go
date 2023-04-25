//go:build sweep
// +build sweep

package pipes

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_pipes_pipe", &resource.Sweeper{
		Name: "aws_pipes_pipe",
		F:    sweepPipes,
	})
}

func sweepPipes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).PipesClient()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	paginator := pipes.NewListPipesPaginator(conn, &pipes.ListPipesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("listing Pipes for %s: %w", region, err))
			break
		}

		for _, it := range page.Pipes {
			name := aws.ToString(it.Name)

			r := ResourcePipe()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Pipe for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Pipe sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
