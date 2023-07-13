// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package lambda

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.LambdaConn(ctx)
	input := &lambda.ListFunctionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListFunctionsPagesWithContext(ctx, input, func(page *lambda.ListFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Functions {
			r := ResourceFunction()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.FunctionName))
			d.Set("function_name", v.FunctionName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lambda Function sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Lambda Functions (%s): %w", region, err)
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
	conn := client.LambdaConn(ctx)
	input := &lambda.ListLayersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListLayersPagesWithContext(ctx, input, func(page *lambda.ListLayersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Layers {
			layerName := aws.StringValue(v.LayerName)
			input := &lambda.ListLayerVersionsInput{
				LayerName: aws.String(layerName),
			}

			err := conn.ListLayerVersionsPagesWithContext(ctx, input, func(page *lambda.ListLayerVersionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.LayerVersions {
					r := ResourceLayerVersion()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.LayerVersionArn))
					d.Set("layer_name", layerName)
					d.Set("version", strconv.Itoa(int(aws.Int64Value(v.Version))))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Lambda Layer Versions (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Lambda Layer Version sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Lambda Layers (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Lambda Layer Versions (%s): %w", region, err))
	}

	return nil
}
