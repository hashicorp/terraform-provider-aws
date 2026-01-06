// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_ssm_parameter")
func newParameterResourceAsListResource() inttypes.ListResourceForSDK {
	l := parameterListResource{}
	l.SetResourceSchema(resourceParameter())
	return &l
}

type parameterListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type parameterListResourceModel struct {
	framework.WithRegionModel
}

func (l *parameterListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SSMClient(ctx)

	var query parameterListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ssm.DescribeParametersInput

	tflog.Info(ctx, "Listing SSM parameters")

	stream.Results = func(yield func(list.ListResult) bool) {
		for parameter, err := range listParameters(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(parameter.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), name)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)

			tflog.Info(ctx, "Reading SSM parameter")
			diags := resourceParameterRead(ctx, rd, awsClient)
			if diags.HasError() {
				result = fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("reading SSM parameter %s", name))
				yield(result)
				return
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			result.DisplayName = name

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
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

func listParameters(ctx context.Context, conn *ssm.Client, input *ssm.DescribeParametersInput) iter.Seq2[awstypes.ParameterMetadata, error] {
	return func(yield func(awstypes.ParameterMetadata, error) bool) {
		pages := ssm.NewDescribeParametersPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.ParameterMetadata{}, fmt.Errorf("listing SSM Parameters: %w", err))
				return
			}

			for _, parameter := range page.Parameters {
				if !yield(parameter, nil) {
					return
				}
			}
		}
	}
}
