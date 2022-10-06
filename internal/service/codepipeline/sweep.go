//go:build sweep
// +build sweep

package codepipeline

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_codepipeline", &resource.Sweeper{
		Name: "aws_codepipeline",
		F:    sweepPipelines,
	})
}

func sweepPipelines(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).CodePipelineConn
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &codepipeline.ListPipelinesInput{}

	err = conn.ListPipelinesPages(input, func(page *codepipeline.ListPipelinesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, pipeline := range page.Pipelines {
			r := ResourceCodePipeline()
			d := r.Data(nil)

			d.SetId(aws.StringValue(pipeline.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Codepipeline Pipeline for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Codepipeline Pipeline for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Codepipeline Pipeline sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
