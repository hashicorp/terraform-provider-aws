// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_lambdaweb_endpoint", sweepEndpoints)
	awsv2.Register("aws_lambdaweb_function", sweepFunctions, "aws_lambdaweb_endpoint")
}

func sweepFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LambdaWebClient(ctx)
	var sweepResources []sweep.Sweepable

	input := lambdaweb.ListWebFunctionsInput{}
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
	conn := client.LambdaWebClient(ctx)
	var sweepResources []sweep.Sweepable

	functionsInput := lambdaweb.ListWebFunctionsInput{}
	for {
		functionsOut, err := conn.ListWebFunctions(ctx, &functionsInput)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, function := range functionsOut.Functions {
			functionName := aws.ToString(function.FunctionName)

			endpointsInput := lambdaweb.ListWebFunctionEndpointsInput{
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
