// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_cloudwatch_log_stream")
func newStreamResourceAsListResource() inttypes.ListResourceForSDK {
	l := streamListResource{}
	l.SetResourceSchema(resourceStream())
	return &l
}

var _ list.ListResource = &streamListResource{}

type streamListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *streamListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"descending": listschema.BoolAttribute{
				Optional:    true,
				Description: "If `true`, results are returned in descending order.",
			},
			names.AttrLogGroupName: listschema.StringAttribute{
				Required:    true,
				Description: "Name of the log group.",
			},
			"order_by": listschema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.OrderBy](),
				Optional:    true,
				Description: "The method used to sort the log streams. Valid values are `LogStreamName` or `LastEventTime`.",
			},
		},
	}
}

func (l *streamListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().LogsClient(ctx)

	var query listStreamModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	logGroupName := fwflex.StringValueFromFramework(ctx, query.LogGroupName)

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey(names.AttrLogGroupName): logGroupName,
	})

	var input cloudwatchlogs.DescribeLogStreamsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for output, err := range listLogStreams(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			result := request.NewListResult(ctx)

			name := aws.ToString(output.LogStreamName)
			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrLogGroupName, logGroupName)
			resourceStreamFlatten(ctx, rd, output)

			result.DisplayName = name

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

type listStreamModel struct {
	framework.WithRegionModel
	Descending   types.Bool                           `tfsdk:"descending"`
	LogGroupName types.String                         `tfsdk:"log_group_name"`
	OrderBy      fwtypes.StringEnum[awstypes.OrderBy] `tfsdk:"order_by"`
}

func listLogStreams(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) iter.Seq2[awstypes.LogStream, error] { // nosemgrep:ci.logs-in-func-name
	return tfiter.ConcatValuesWithError(listLogStreamPages(ctx, conn, input, optFns...))
}
