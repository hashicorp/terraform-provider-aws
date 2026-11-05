// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_cloudwatch_log_stream")
func newStreamResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceStream(), framework.SDKListSpec[awstypes.LogStream, listStreamModel]{
		ConfigSchema: listStreamConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listStreamModel) (iter.Seq2[awstypes.LogStream, error], diag.Diagnostics) {
			var input cloudwatchlogs.DescribeLogStreamsInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listLogStreams(ctx, c.LogsClient(ctx), &input), nil
		},
		ID: func(ls awstypes.LogStream, _ listStreamModel) string {
			return aws.ToString(ls.LogStreamName)
		},
		Prepare: func(ctx context.Context, _ awstypes.LogStream, query listStreamModel, rd *schema.ResourceData) {
			// The log group is a required query argument and part of this
			// resource's identity, and no field of the item carries it.
			rd.Set(names.AttrLogGroupName, fwflex.StringValueFromFramework(ctx, query.LogGroupName))
		},
		Flatten: resourceStreamFlatten,
	})
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

func listStreamConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
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
