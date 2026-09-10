// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_sfn_state_machine")
func newStateMachineResourceAsListResource() inttypes.ListResourceForSDK {
	l := stateMachineListResource{}
	l.SetResourceSchema(resourceStateMachine())
	return &l
}

var _ list.ListResource = &stateMachineListResource{}

type stateMachineListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *stateMachineListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.SFNClient(ctx)

	tflog.Info(ctx, "Listing Step Functions State Machines")

	stream.Results = func(yield func(list.ListResult) bool) {
		var input sfn.ListStateMachinesInput
		for item, err := range listStateMachines(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			arn := aws.ToString(item.StateMachineArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(arn)
			rd.Set(names.AttrARN, arn)

			if request.IncludeResource {
				tflog.Info(ctx, "Reading Step Functions State Machine")
				output, err := findStateMachineByARN(ctx, conn, arn)
				if err != nil {
					tflog.Error(ctx, "Reading Step Functions State Machine", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				diags := resourceStateMachineFlatten(ctx, conn, output, rd)
				if diags.HasError() {
					tflog.Error(ctx, "Flattening Step Functions State Machine", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
					})
					continue
				}
			}

			result.DisplayName = aws.ToString(item.Name)

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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

func listStateMachines(ctx context.Context, conn *sfn.Client, input *sfn.ListStateMachinesInput) iter.Seq2[awstypes.StateMachineListItem, error] {
	return func(yield func(awstypes.StateMachineListItem, error) bool) {
		pages := sfn.NewListStateMachinesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.StateMachineListItem{}, fmt.Errorf("listing Step Functions State Machines: %w", err))
				return
			}

			for _, item := range page.StateMachines {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
