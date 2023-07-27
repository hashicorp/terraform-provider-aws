// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package synthetics

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_synthetics_canary", &resource.Sweeper{
		Name: "aws_synthetics_canary",
		F:    sweepCanaries,
		Dependencies: []string{
			"aws_lambda_function",
			"aws_lambda_layer",
			"aws_cloudwatch_log_group",
		},
	})
}

func sweepCanaries(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.SyntheticsConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &synthetics.DescribeCanariesInput{}
	for {
		output, err := conn.DescribeCanariesWithContext(ctx, input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Synthetics Canary sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Synthetics Canaries: %w", err)
		}

		for _, canary := range output.Canaries {
			name := aws.StringValue(canary.Name)
			log.Printf("[INFO] Deleting Synthetics Canary: %s", name)

			r := ResourceCanary()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Synthetics Canaries: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
