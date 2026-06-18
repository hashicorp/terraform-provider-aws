// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

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

	data.EventBridge = types.BoolValue(output.EventBridgeConfiguration != nil)

	lambdaFunction, diags := flattenBucketNotificationLambdaFunctions(ctx, output.LambdaFunctionConfigurations)
	resp.Diagnostics.Append(diags...)
	queue, diags := flattenBucketNotificationQueues(ctx, output.QueueConfigurations)
	resp.Diagnostics.Append(diags...)
	topic, diags := flattenBucketNotificationTopics(ctx, output.TopicConfigurations)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LambdaFunction = lambdaFunction
	data.Queue = queue
	data.Topic = topic

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

type bucketNotificationQueueModel struct {
	Events       fwtypes.SetOfString `tfsdk:"events"`
	FilterPrefix types.String        `tfsdk:"filter_prefix"`
	FilterSuffix types.String        `tfsdk:"filter_suffix"`
	ID           types.String        `tfsdk:"id"`
	QueueARN     types.String        `tfsdk:"queue_arn"`
}

type bucketNotificationTopicModel struct {
	Events       fwtypes.SetOfString `tfsdk:"events"`
	FilterPrefix types.String        `tfsdk:"filter_prefix"`
	FilterSuffix types.String        `tfsdk:"filter_suffix"`
	ID           types.String        `tfsdk:"id"`
	TopicARN     types.String        `tfsdk:"topic_arn"`
}

func flattenBucketNotificationLambdaFunctions(ctx context.Context, in []awstypes.LambdaFunctionConfiguration) (fwtypes.ListNestedObjectValueOf[bucketNotificationLambdaFunctionModel], diag.Diagnostics) {
	out := make([]*bucketNotificationLambdaFunctionModel, 0, len(in))
	for _, c := range in {
		prefix, suffix := bucketNotificationFilterRulePrefixSuffix(c.Filter)
		out = append(out, &bucketNotificationLambdaFunctionModel{
			Events:            fwflex.FlattenFrameworkStringValueSetOfString(ctx, eventsToStrings(c.Events)),
			FilterPrefix:      types.StringValue(prefix),
			FilterSuffix:      types.StringValue(suffix),
			ID:                types.StringPointerValue(c.Id),
			LambdaFunctionARN: types.StringPointerValue(c.LambdaFunctionArn),
		})
	}
	return fwtypes.NewListNestedObjectValueOfSlice(ctx, out, nil)
}

func flattenBucketNotificationQueues(ctx context.Context, in []awstypes.QueueConfiguration) (fwtypes.ListNestedObjectValueOf[bucketNotificationQueueModel], diag.Diagnostics) {
	out := make([]*bucketNotificationQueueModel, 0, len(in))
	for _, c := range in {
		prefix, suffix := bucketNotificationFilterRulePrefixSuffix(c.Filter)
		out = append(out, &bucketNotificationQueueModel{
			Events:       fwflex.FlattenFrameworkStringValueSetOfString(ctx, eventsToStrings(c.Events)),
			FilterPrefix: types.StringValue(prefix),
			FilterSuffix: types.StringValue(suffix),
			ID:           types.StringPointerValue(c.Id),
			QueueARN:     types.StringPointerValue(c.QueueArn),
		})
	}
	return fwtypes.NewListNestedObjectValueOfSlice(ctx, out, nil)
}

func flattenBucketNotificationTopics(ctx context.Context, in []awstypes.TopicConfiguration) (fwtypes.ListNestedObjectValueOf[bucketNotificationTopicModel], diag.Diagnostics) {
	out := make([]*bucketNotificationTopicModel, 0, len(in))
	for _, c := range in {
		prefix, suffix := bucketNotificationFilterRulePrefixSuffix(c.Filter)
		out = append(out, &bucketNotificationTopicModel{
			Events:       fwflex.FlattenFrameworkStringValueSetOfString(ctx, eventsToStrings(c.Events)),
			FilterPrefix: types.StringValue(prefix),
			FilterSuffix: types.StringValue(suffix),
			ID:           types.StringPointerValue(c.Id),
			TopicARN:     types.StringPointerValue(c.TopicArn),
		})
	}
	return fwtypes.NewListNestedObjectValueOfSlice(ctx, out, nil)
}

func bucketNotificationFilterRulePrefixSuffix(filter *awstypes.NotificationConfigurationFilter) (string, string) {
	if filter == nil || filter.Key == nil {
		return "", ""
	}
	var prefix, suffix string
	for _, rule := range filter.Key.FilterRules {
		switch rule.Name {
		case awstypes.FilterRuleNamePrefix:
			if rule.Value != nil {
				prefix = *rule.Value
			}
		case awstypes.FilterRuleNameSuffix:
			if rule.Value != nil {
				suffix = *rule.Value
			}
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
