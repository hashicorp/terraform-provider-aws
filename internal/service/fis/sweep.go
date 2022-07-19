//go:build sweep
// +build sweep

package fis

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_fis_experiment_template", &resource.Sweeper{
		Name: "aws_fis_experiment_template",
		F:    sweepExperimentTemplates,
	})
}

func sweepExperimentTemplates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).FISConn
	input := &fis.ListExperimentTemplatesInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListExperimentTemplatesPages(input, func(page *fis.ListExperimentTemplatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, experimentTemplate := range page.ExperimentTemplates {
			r := ResourceExperimentTemplate()
			d := r.Data(nil)
			d.SetId(aws.StringValue(experimentTemplate.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping FIS Experiment Template sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing FIS Experiment Templates (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping FIS Experiment Templates (%s): %w", region, err)
	}

	return nil
}
