// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_osis_resource_policy")
func newResourcePolicyResourceAsListResource() list.ListResourceWithConfigure {
	return &resourcePolicyListResource{}
}

var _ list.ListResource = &resourcePolicyListResource{}

type resourcePolicyListResource struct {
	resourcePolicyResource
	framework.WithList
}

func (l *resourcePolicyListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().OpenSearchIngestionClient(ctx)

	var query listResourcePolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		var input osis.ListPipelinesInput
		for output, err := range listResourcePolicies(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			resourceARN := aws.ToString(output.ResourceArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrResourceARN), resourceARN)

			result := request.NewListResult(ctx)

			var data resourcePolicyResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, output, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				result.DisplayName = resourceARN
			})

			if !yield(result) {
				return
			}
		}
	}
}

type listResourcePolicyModel struct {
	framework.WithRegionModel
}

func listResourcePolicies(ctx context.Context, conn *osis.Client, input *osis.ListPipelinesInput) iter.Seq2[*osis.GetResourcePolicyOutput, error] {
	return func(yield func(*osis.GetResourcePolicyOutput, error) bool) {
		for pipeline, err := range listPipelines(ctx, conn, input) {
			if err != nil {
				yield(nil, err)
				return
			}

			pipelineARN := aws.ToString(pipeline.PipelineArn)
			output, err := findResourcePolicyByResourceARN(ctx, conn, pipelineARN)

			if retry.NotFound(err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				yield(nil, fmt.Errorf("getting resource policy for pipeline %s: %w", pipelineARN, err))
				return
			}

			if !yield(output, nil) {
				return
			}
		}
	}
}
