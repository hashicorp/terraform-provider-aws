// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webfunctions

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/webfunctions"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_webfunctions_endpoint", sweepEndpoints)
	awsv2.Register("aws_webfunctions_function", sweepFunctions, "aws_webfunctions_endpoint")
}

func sweepFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WebFunctionsClient(ctx)
	var sweepResources []sweep.Sweepable

	input := webfunctions.ListWebFunctionsInput{}
	for {
		out, err := conn.ListWebFunctions(ctx, &input)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range out.Functions {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newFunctionResource, client,
				sweepfw.NewAttribute("function_name", aws.ToString(v.FunctionName))),
			)
		}

		if out.NextToken == nil {
			break
		}
		input.NextToken = out.NextToken
	}

	return sweepResources, nil
}

func sweepEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WebFunctionsClient(ctx)
	var sweepResources []sweep.Sweepable

	functionsInput := webfunctions.ListWebFunctionsInput{}
	for {
		functionsOut, err := conn.ListWebFunctions(ctx, &functionsInput)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, function := range functionsOut.Functions {
			functionName := aws.ToString(function.FunctionName)

			endpointsInput := webfunctions.ListWebFunctionEndpointsInput{
				FunctionName: aws.String(functionName),
			}
			for {
				endpointsOut, err := conn.ListWebFunctionEndpoints(ctx, &endpointsInput)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, endpoint := range endpointsOut.Endpoints {
					sweepResources = append(sweepResources, sweepfw.NewSweepResource(newEndpointResource, client,
						sweepfw.NewAttribute("function_name", functionName),
						sweepfw.NewAttribute("endpoint_name", aws.ToString(endpoint.EndpointName))),
					)
				}

				if endpointsOut.NextToken == nil {
					break
				}
				endpointsInput.NextToken = endpointsOut.NextToken
			}
		}

		if functionsOut.NextToken == nil {
			break
		}
		functionsInput.NextToken = functionsOut.NextToken
	}

	return sweepResources, nil
}
