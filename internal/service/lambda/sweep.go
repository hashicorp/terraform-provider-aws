// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_lambda_function", &resource.Sweeper{
		Name: "aws_lambda_function",
		F:    sweepFunctions,
	})

	resource.AddTestSweepers("aws_lambda_layer", &resource.Sweeper{
		Name: "aws_lambda_layer",
		F:    sweepLayerVersions,
		Dependencies: []string{
			"aws_lambda_function",
		},
	})
}

func sweepFunctions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.LambdaClient(ctx)
	input := &lambda.ListFunctionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lambda.NewListFunctionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lambda Function sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Lambda Functions (%s): %w", region, err)
		}

		for _, v := range page.Functions {
			r := resourceFunction()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FunctionName))
			d.Set("function_name", v.FunctionName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Lambda Functions (%s): %w", region, err)
	}

	return nil
}

func sweepLayerVersions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.LambdaClient(ctx)
	input := &lambda.ListLayersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lambda.NewListLayersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lambda Layer Version sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Lambda Layers (%s): %w", region, err)
		}

		for _, v := range page.Layers {
			layerName := aws.ToString(v.LayerName)
			input := &lambda.ListLayerVersionsInput{
				LayerName: aws.String(layerName),
			}

			pages := lambda.NewListLayerVersionsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.LayerVersions {
					r := resourceLayerVersion()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.LayerVersionArn))
					d.Set("layer_name", layerName)
					d.Set(names.AttrVersion, strconv.Itoa(int(v.Version)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Lambda Layer Versions (%s): %w", region, err)
	}

	return nil
}
