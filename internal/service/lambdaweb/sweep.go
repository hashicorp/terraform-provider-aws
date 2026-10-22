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
	awsv2.Register("aws_lambdaweb_function", sweepFunctions, "aws_lambdaweb_function_revision")
	awsv2.Register("aws_lambdaweb_function_endpoint", sweepFunctionEndpoints)
	awsv2.Register("aws_lambdaweb_function_revision", sweepFunctionRevisions, "aws_lambdaweb_function_endpoint")
}

func sweepFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input lambdaweb.ListWebFunctionsInput
	conn := client.LambdaWebClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listFunctions(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		sweepResources = append(sweepResources, sweepfw.NewSweepResource(newFunctionResource, client,
			sweepfw.NewAttribute("function_name", aws.ToString(v.FunctionName))),
		)
	}

	return sweepResources, nil
}

func sweepFunctionEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input lambdaweb.ListWebFunctionsInput
	conn := client.LambdaWebClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listFunctions(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		functionName := aws.ToString(v.FunctionName)
		input := lambdaweb.ListWebFunctionEndpointsInput{
			FunctionName: aws.String(functionName),
		}
		for v, err := range listFunctionEndpoints(ctx, conn, &input) {
			if err != nil {
				return nil, smarterr.NewError(err)
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newFunctionEndpointResource, client,
				sweepfw.NewAttribute("function_name", functionName),
				sweepfw.NewAttribute("endpoint_name", aws.ToString(v.EndpointName))),
			)
		}
	}

	return sweepResources, nil
}

func sweepFunctionRevisions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input lambdaweb.ListWebFunctionsInput
	conn := client.LambdaWebClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listFunctions(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		functionName := aws.ToString(v.FunctionName)
		input := lambdaweb.ListWebFunctionRevisionsInput{
			FunctionName: aws.String(functionName),
		}
		for v, err := range listFunctionRevisions(ctx, conn, &input) {
			if err != nil {
				return nil, smarterr.NewError(err)
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newFunctionEndpointResource, client,
				sweepfw.NewAttribute("function_name", functionName),
				sweepfw.NewAttribute("revision_id", aws.ToString(v.RevisionId))),
			)
		}
	}

	return sweepResources, nil
}
