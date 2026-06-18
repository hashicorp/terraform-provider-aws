// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// TestFlattenBucketNotificationDataSourceModel verifies the AWS-to-model
// transformation in isolation, without an acceptance test. It exercises every
// field AutoFlex is expected to handle and the pieces that need help: the
// EventBridge presence-as-bool patch and the Filter.Key.FilterRules pivot
// implemented via Flattener on the per-destination models.
func TestFlattenBucketNotificationDataSourceModel(t *testing.T) {
	ctx := context.Background()

	output := &s3.GetBucketNotificationConfigurationOutput{
		EventBridgeConfiguration: &awstypes.EventBridgeConfiguration{},
		LambdaFunctionConfigurations: []awstypes.LambdaFunctionConfiguration{{
			Id:                aws.String("notification-lambda"),
			LambdaFunctionArn: aws.String("arn:aws:lambda:us-east-1:111111111111:function:fn"),
			Events:            []awstypes.Event{awstypes.EventS3ObjectCreated, awstypes.EventS3ObjectRemovedDelete},
			Filter: &awstypes.NotificationConfigurationFilter{
				Key: &awstypes.S3KeyFilter{
					FilterRules: []awstypes.FilterRule{
						{Name: "Prefix", Value: aws.String("tf-acc-test/")},
						{Name: "Suffix", Value: aws.String(".png")},
					},
				},
			},
		}},
		QueueConfigurations: []awstypes.QueueConfiguration{{
			Id:       aws.String("notification-queue"),
			QueueArn: aws.String("arn:aws:sqs:us-east-1:111111111111:queue"),
			Events:   []awstypes.Event{awstypes.EventS3ObjectCreated},
			Filter: &awstypes.NotificationConfigurationFilter{
				Key: &awstypes.S3KeyFilter{
					FilterRules: []awstypes.FilterRule{
						{Name: "Prefix", Value: aws.String("queues/")},
					},
				},
			},
		}},
		TopicConfigurations: []awstypes.TopicConfiguration{{
			Id:       aws.String("notification-topic"),
			TopicArn: aws.String("arn:aws:sns:us-east-1:111111111111:topic"),
			Events:   []awstypes.Event{awstypes.EventS3ObjectCreated},
		}},
	}

	var data bucketNotificationDataSourceModel
	if diags := flattenBucketNotificationDataSourceModel(ctx, output, &data); diags.HasError() {
		t.Fatalf("flatten produced errors: %v", diags)
	}

	if got, want := data.EventBridge.ValueBool(), true; got != want {
		t.Errorf("EventBridge: got %v, want %v", got, want)
	}

	lambdaFns, d := data.LambdaFunction.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("LambdaFunction.ToSlice: %v", d)
	}
	if got, want := len(lambdaFns), 1; got != want {
		t.Fatalf("LambdaFunction len: got %d, want %d", got, want)
	}
	l := lambdaFns[0]
	if got, want := l.ID.ValueString(), "notification-lambda"; got != want {
		t.Errorf("lambda.ID: got %q, want %q", got, want)
	}
	if got, want := l.LambdaFunctionARN.ValueString(), "arn:aws:lambda:us-east-1:111111111111:function:fn"; got != want {
		t.Errorf("lambda.LambdaFunctionARN: got %q, want %q", got, want)
	}
	if got, want := l.FilterPrefix.ValueString(), "tf-acc-test/"; got != want {
		t.Errorf("lambda.FilterPrefix: got %q, want %q", got, want)
	}
	if got, want := l.FilterSuffix.ValueString(), ".png"; got != want {
		t.Errorf("lambda.FilterSuffix: got %q, want %q", got, want)
	}
	if l.Events.IsNull() || l.Events.IsUnknown() {
		t.Errorf("lambda.Events: null/unknown, want set with two elements")
	}

	queues, d := data.Queue.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("Queue.ToSlice: %v", d)
	}
	if got, want := len(queues), 1; got != want {
		t.Fatalf("Queue len: got %d, want %d", got, want)
	}
	q := queues[0]
	if got, want := q.ID.ValueString(), "notification-queue"; got != want {
		t.Errorf("queue.ID: got %q, want %q", got, want)
	}
	if got, want := q.QueueARN.ValueString(), "arn:aws:sqs:us-east-1:111111111111:queue"; got != want {
		t.Errorf("queue.QueueARN: got %q, want %q", got, want)
	}
	if got, want := q.FilterPrefix.ValueString(), "queues/"; got != want {
		t.Errorf("queue.FilterPrefix: got %q, want %q", got, want)
	}
	if got, want := q.FilterSuffix.ValueString(), ""; got != want {
		t.Errorf("queue.FilterSuffix: got %q, want %q", got, want)
	}

	topics, d := data.Topic.ToSlice(ctx)
	if d.HasError() {
		t.Fatalf("Topic.ToSlice: %v", d)
	}
	if got, want := len(topics), 1; got != want {
		t.Fatalf("Topic len: got %d, want %d", got, want)
	}
	tp := topics[0]
	if got, want := tp.ID.ValueString(), "notification-topic"; got != want {
		t.Errorf("topic.ID: got %q, want %q", got, want)
	}
	if got, want := tp.TopicARN.ValueString(), "arn:aws:sns:us-east-1:111111111111:topic"; got != want {
		t.Errorf("topic.TopicARN: got %q, want %q", got, want)
	}
	if got, want := tp.FilterPrefix.ValueString(), ""; got != want {
		t.Errorf("topic.FilterPrefix: got %q, want %q", got, want)
	}
	if got, want := tp.FilterSuffix.ValueString(), ""; got != want {
		t.Errorf("topic.FilterSuffix: got %q, want %q", got, want)
	}
}
