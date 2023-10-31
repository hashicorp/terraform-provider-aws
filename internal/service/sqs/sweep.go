// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func RegisterSweepers() {
	sweep.Register("aws_sqs_queue", sweepQueues,
		"aws_autoscaling_group",
		"aws_cloudwatch_event_rule",
		"aws_elastic_beanstalk_environment",
		"aws_iot_topic_rule",
		"aws_lambda_function",
		"aws_s3_bucket",
		"aws_sns_topic",
	)
}

func sweepQueues(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SQSConn(ctx)

	var sweepResources []sweep.Sweepable

	input := &sqs.ListQueuesInput{}

	err := conn.ListQueuesPagesWithContext(ctx, input, func(page *sqs.ListQueuesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queueUrl := range page.QueueUrls {
			r := ResourceQueue()
			d := r.Data(nil)
			d.SetId(aws.StringValue(queueUrl))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	return sweepResources, err
}
