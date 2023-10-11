// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package fis

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_fis_experiment_template", &resource.Sweeper{
		Name: "aws_fis_experiment_template",
		F:    sweepExperimentTemplates,
	})
}

func sweepExperimentTemplates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.FISClient(ctx)
	input := &fis.ListExperimentTemplatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pg := fis.NewListExperimentTemplatesPaginator(conn, input)

	for pg.HasMorePages() {
		page, err := pg.NextPage(ctx)

		if err != nil {
			sweeperErr := fmt.Errorf("error listing FIS Experiment Templates: %w", err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, experimentTemplate := range page.ExperimentTemplates {
			r := ResourceExperimentTemplate()
			d := r.Data(nil)
			d.SetId(aws.ToString(experimentTemplate.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping FIS Experiment Templates (%s): %w", region, err)
	}

	return nil
}
