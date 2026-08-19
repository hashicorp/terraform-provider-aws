// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3_bucket_notification", name="Bucket Notification")
func newBucketNotificationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &bucketNotificationDataSource{}, nil
}

type bucketNotificationDataSource struct {
	framework.DataSourceWithModel[bucketNotificationDataSourceModel]
}

func (d *bucketNotificationDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrBucket: schema.StringAttribute{
				Required: true,
			},
			"eventbridge": schema.BoolAttribute{
				Computed: true,
			},
			"lambda_function": framework.DataSourceComputedListOfObjectAttribute[bucketNotificationLambdaFunctionModel](ctx),
			"queue":           framework.DataSourceComputedListOfObjectAttribute[bucketNotificationQueueModel](ctx),
			"topic":           framework.DataSourceComputedListOfObjectAttribute[bucketNotificationTopicModel](ctx),
		},
	}
}

func (d *bucketNotificationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bucketNotificationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket := data.Bucket.ValueString()
	conn := d.Meta().S3Client(ctx)
	if isDirectoryBucket(bucket) {
		conn = d.Meta().S3ExpressClient(ctx)
	}

	output, err := findBucketNotificationConfiguration(ctx, conn, bucket, "")
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, bucket)
		return
	}

	resp.Diagnostics.Append(bucketNotificationDataSourceFlatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// bucketNotificationDataSourceFlatten maps the GetBucketNotificationConfiguration
// API output onto the data source model. Split out from Read so it can be
// reasoned about (and unit tested) without standing up an acceptance test.
//
// AutoFlex does the bulk of the work, given one hint: WithFieldNameSuffix("Configurations")
// lets it match LambdaFunctionConfigurations / QueueConfigurations / TopicConfigurations
// against the singular nested-block names in the model. AutoFlex then dispatches to
// each per-destination model's Flatten method (Flattener interface) for the
// FilterRules{Name,Value} pivot into filter_prefix / filter_suffix. EventBridge
// presence-as-bool is patched here since it's a single field that doesn't justify a
// custom Flattener on the top model.
func bucketNotificationDataSourceFlatten(ctx context.Context, output *s3.GetBucketNotificationConfigurationOutput, data *bucketNotificationDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, output, data, fwflex.WithFieldNameSuffix("Configurations"))...)
	if diags.HasError() {
		return diags
	}
	data.EventBridge = types.BoolValue(output.EventBridgeConfiguration != nil)
	return diags
}

func bucketNotificationFilterRulePrefixSuffix(filter *awstypes.NotificationConfigurationFilter) (types.String, types.String) {
	prefix, suffix := types.StringNull(), types.StringNull()
	if filter == nil || filter.Key == nil {
		return prefix, suffix
	}
	// AWS returns "Prefix"/"Suffix" with title case while the SDK enum
	// constants are lowercase, so normalize before matching.
	for _, rule := range filter.Key.FilterRules {
		switch strings.ToLower(string(rule.Name)) {
		case string(awstypes.FilterRuleNamePrefix):
			prefix = types.StringValue(aws.ToString(rule.Value))
		case string(awstypes.FilterRuleNameSuffix):
			suffix = types.StringValue(aws.ToString(rule.Value))
		}
	}
	return prefix, suffix
}

func eventsToStrings(events []awstypes.Event) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = string(e)
	}
	return out
}

type bucketNotificationDataSourceModel struct {
	framework.WithRegionModel
	Bucket         types.String                                                           `tfsdk:"bucket"`
	EventBridge    types.Bool                                                             `tfsdk:"eventbridge"`
	LambdaFunction fwtypes.ListNestedObjectValueOf[bucketNotificationLambdaFunctionModel] `tfsdk:"lambda_function"`
	Queue          fwtypes.ListNestedObjectValueOf[bucketNotificationQueueModel]          `tfsdk:"queue"`
	Topic          fwtypes.ListNestedObjectValueOf[bucketNotificationTopicModel]          `tfsdk:"topic"`
}

type bucketNotificationLambdaFunctionModel struct {
	Events            fwtypes.SetOfString `tfsdk:"events"`
	FilterPrefix      types.String        `tfsdk:"filter_prefix"`
	FilterSuffix      types.String        `tfsdk:"filter_suffix"`
	ID                types.String        `tfsdk:"id"`
	LambdaFunctionARN types.String        `tfsdk:"lambda_function_arn"`
}

func (m *bucketNotificationLambdaFunctionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	var c *awstypes.LambdaFunctionConfiguration
	switch t := v.(type) {
	case awstypes.LambdaFunctionConfiguration:
		c = &t
	case *awstypes.LambdaFunctionConfiguration:
		c = t
	default:
		if v == nil {
			return diags
		}
		diags.AddError(
			"Unexpected source type",
			fmt.Sprintf("flattening bucket_notification.lambda_function: got %T, expected awstypes.LambdaFunctionConfiguration", v),
		)
		return diags
	}
	if c == nil {
		return diags
	}
	m.ID = types.StringPointerValue(c.Id)
	m.LambdaFunctionARN = types.StringPointerValue(c.LambdaFunctionArn)
	m.Events = fwflex.FlattenFrameworkStringValueSetOfString(ctx, eventsToStrings(c.Events))
	m.FilterPrefix, m.FilterSuffix = bucketNotificationFilterRulePrefixSuffix(c.Filter)
	return diags
}

type bucketNotificationQueueModel struct {
	Events       fwtypes.SetOfString `tfsdk:"events"`
	FilterPrefix types.String        `tfsdk:"filter_prefix"`
	FilterSuffix types.String        `tfsdk:"filter_suffix"`
	ID           types.String        `tfsdk:"id"`
	QueueARN     types.String        `tfsdk:"queue_arn"`
}

func (m *bucketNotificationQueueModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	var c *awstypes.QueueConfiguration
	switch t := v.(type) {
	case awstypes.QueueConfiguration:
		c = &t
	case *awstypes.QueueConfiguration:
		c = t
	default:
		if v == nil {
			return diags
		}
		diags.AddError(
			"Unexpected source type",
			fmt.Sprintf("flattening bucket_notification.queue: got %T, expected awstypes.QueueConfiguration", v),
		)
		return diags
	}
	if c == nil {
		return diags
	}
	m.ID = types.StringPointerValue(c.Id)
	m.QueueARN = types.StringPointerValue(c.QueueArn)
	m.Events = fwflex.FlattenFrameworkStringValueSetOfString(ctx, eventsToStrings(c.Events))
	m.FilterPrefix, m.FilterSuffix = bucketNotificationFilterRulePrefixSuffix(c.Filter)
	return diags
}

type bucketNotificationTopicModel struct {
	Events       fwtypes.SetOfString `tfsdk:"events"`
	FilterPrefix types.String        `tfsdk:"filter_prefix"`
	FilterSuffix types.String        `tfsdk:"filter_suffix"`
	ID           types.String        `tfsdk:"id"`
	TopicARN     types.String        `tfsdk:"topic_arn"`
}

func (m *bucketNotificationTopicModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	var c *awstypes.TopicConfiguration
	switch t := v.(type) {
	case awstypes.TopicConfiguration:
		c = &t
	case *awstypes.TopicConfiguration:
		c = t
	default:
		if v == nil {
			return diags
		}
		diags.AddError(
			"Unexpected source type",
			fmt.Sprintf("flattening bucket_notification.topic: got %T, expected awstypes.TopicConfiguration", v),
		)
		return diags
	}
	if c == nil {
		return diags
	}
	m.ID = types.StringPointerValue(c.Id)
	m.TopicARN = types.StringPointerValue(c.TopicArn)
	m.Events = fwflex.FlattenFrameworkStringValueSetOfString(ctx, eventsToStrings(c.Events))
	m.FilterPrefix, m.FilterSuffix = bucketNotificationFilterRulePrefixSuffix(c.Filter)
	return diags
}
