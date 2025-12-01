// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"strconv"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_lambda_capacity_provider", sweepCapacityProviders, "aws_lambda_function")
	awsv2.Register("aws_lambda_function", sweepFunctions)
	awsv2.Register("aws_lambda_layer", sweepLayerVersions, "aws_lambda_function")
}

func sweepCapacityProviders(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := lambda.ListCapacityProvidersInput{}
	conn := client.LambdaClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := lambda.NewListCapacityProvidersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.CapacityProviders {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceCapacityProvider, client,
				sweepfw.NewAttribute(names.AttrName, aws.ToString(v.CapacityProviderArn))),
			)
		}
	}

	return sweepResources, nil
}

func sweepFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LambdaClient(ctx)
	input := &lambda.ListFunctionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lambda.NewListFunctionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Functions {
			r := resourceFunction()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FunctionName))
			d.Set("function_name", v.FunctionName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepLayerVersions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LambdaClient(ctx)
	input := &lambda.ListLayersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := lambda.NewListLayersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, smarterr.NewError(err)
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

	return sweepResources, nil
}
