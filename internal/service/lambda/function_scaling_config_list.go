// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @FrameworkListResource("aws_lambda_function_scaling_config")
func newFunctionScalingConfigResourceAsListResource() list.ListResourceWithConfigure {
	return &listResourceFunctionScalingConfig{}
}

var _ list.ListResource = &listResourceFunctionScalingConfig{}

type listResourceFunctionScalingConfig struct {
	functionScalingConfigResource
	framework.WithList
}

func (r *listResourceFunctionScalingConfig) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().LambdaClient(ctx)

	var query functionScalingConfigListModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		// Scaling configurations only apply to functions using a capacity provider,
		// so enumerate capacity providers, then the function versions attached to
		// each, and finally read the scaling configuration for each version.
		var cpInput lambda.ListCapacityProvidersInput
		for capacityProvider, err := range listCapacityProviders(ctx, conn, &cpInput) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			if capacityProvider.State == awstypes.CapacityProviderStateDeleting {
				continue
			}

			cpARN, err := arn.Parse(aws.ToString(capacityProvider.CapacityProviderArn))
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}
			capacityProviderName := strings.TrimPrefix(cpARN.Resource, "capacity-provider:")

			for functionVersion, err := range listFunctionVersionsByCapacityProvider(ctx, conn, capacityProviderName) {
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				functionName, qualifier, err := functionNameQualifierFromARN(aws.ToString(functionVersion.FunctionArn))
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				scOut, err := findFunctionScalingConfigByTwoPartKey(ctx, conn, functionName, qualifier)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				result := request.NewListResult(ctx)
				var data functionScalingConfigResourceModel
				r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
					data.FunctionName = flex.StringValueToFramework(ctx, functionName)
					data.Qualifier = flex.StringValueToFramework(ctx, qualifier)

					// Maps function_arn.
					if diags := flex.Flatten(ctx, scOut, &data); diags.HasError() {
						result.Diagnostics.Append(diags...)
						return
					}

					// The scaling configuration is reported under RequestedFunctionScalingConfig
					// while settling and AppliedFunctionScalingConfig once in effect; prefer the
					// requested (most recently desired) values.
					scalingConfig := scOut.RequestedFunctionScalingConfig
					if scalingConfig == nil {
						scalingConfig = scOut.AppliedFunctionScalingConfig
					}
					if diags := flex.Flatten(ctx, scalingConfig, &data.FunctionScalingConfig); diags.HasError() {
						result.Diagnostics.Append(diags...)
						return
					}

					result.DisplayName = functionName + ", " + qualifier
				})

				if result.Diagnostics.HasError() {
					yield(result)
					return
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}

type functionScalingConfigListModel struct {
	framework.WithRegionModel
}

func listFunctionVersionsByCapacityProvider(ctx context.Context, conn *lambda.Client, capacityProviderName string) iter.Seq2[awstypes.FunctionVersionsByCapacityProviderListItem, error] {
	return func(yield func(awstypes.FunctionVersionsByCapacityProviderListItem, error) bool) {
		input := lambda.ListFunctionVersionsByCapacityProviderInput{
			CapacityProviderName: aws.String(capacityProviderName),
		}

		for {
			page, err := conn.ListFunctionVersionsByCapacityProvider(ctx, &input)
			if err != nil {
				yield(awstypes.FunctionVersionsByCapacityProviderListItem{}, fmt.Errorf("listing Lambda Function Versions for Capacity Provider (%s): %w", capacityProviderName, err))
				return
			}

			for _, functionVersion := range page.FunctionVersions {
				if !yield(functionVersion, nil) {
					return
				}
			}

			if aws.ToString(page.NextMarker) == "" {
				return
			}
			input.Marker = page.NextMarker
		}
	}
}

// functionNameQualifierFromARN extracts the function name and version qualifier
// from a function version ARN of the form
// arn:aws:lambda:<region>:<account>:function:<name>:<version>.
func functionNameQualifierFromARN(functionARN string) (string, string, error) {
	parsed, err := arn.Parse(functionARN)
	if err != nil {
		return "", "", err
	}

	// Resource is "function:<name>:<version>".
	parts := strings.Split(parsed.Resource, ":")
	if len(parts) != 3 || parts[0] != "function" || parts[1] == "" || parts[2] == "" {
		return "", "", fmt.Errorf("unexpected Lambda function version ARN: %s", functionARN)
	}

	return parts[1], parts[2], nil
}
