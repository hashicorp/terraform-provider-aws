//go:build sweep
// +build sweep

package kendra

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_kendra_index", &resource.Sweeper{
		Name: "aws_kendra_index",
		F:    sweepIndex,
	})
}

func sweepIndex(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).KendraClient()
	sweepResources := make([]sweep.Sweepable, 0)
	in := &kendra.ListIndicesInput{}
	var errs *multierror.Error

	pages := kendra.NewListIndicesPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Kendra Indices sweep for %s: %s", region, err)
			return errs.ErrorOrNil()
		}

		if err != nil {
			return multierror.Append(errs, fmt.Errorf("retrieving Kendra Indices: %w", err))
		}

		for _, index := range page.IndexConfigurationSummaryItems {
			r := ResourceIndex()
			d := r.Data(nil)
			d.SetId(aws.ToString(index.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Kendra Indices for %s: %w", region, err))
	}

	return errs.ErrorOrNil()
}
