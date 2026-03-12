// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appflow

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appflow"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appflow/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_appflow_flow")
func newFlowResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceFlow{}
	l.SetResourceSchema(resourceFlow())
	return &l
}

var _ list.ListResource = &listResourceFlow{}

type listResourceFlow struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceFlow) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.AppFlowClient(ctx)

	var query listFlowModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing AppFlow Flow")
	stream.Results = func(yield func(list.ListResult) bool) {
		var input appflow.ListFlowsInput
		for item, err := range listFlows(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			flowName := aws.ToString(item.FlowName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), flowName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(flowName)
			rd.Set(names.AttrName, flowName)

			tflog.Info(ctx, "Reading AppFlow Flow")
			output, err := findFlowByName(ctx, conn, flowName)
			if err != nil {
				tflog.Error(ctx, "Reading AppFlow Flow", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			resourceFlowFlatten(ctx, output, rd)

			result.DisplayName = flowName

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

type listFlowModel struct {
	framework.WithRegionModel
}

func listFlows(ctx context.Context, conn *appflow.Client, input *appflow.ListFlowsInput) iter.Seq2[awstypes.FlowDefinition, error] {
	return func(yield func(awstypes.FlowDefinition, error) bool) {
		pages := appflow.NewListFlowsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.FlowDefinition{}, fmt.Errorf("listing AppFlow Flow resources: %w", err))
				return
			}

			for _, item := range page.Flows {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
