// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_sqs_queue", sweepQueues,
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
	conn := client.SQSClient(ctx)
	input := &sqs.ListQueuesInput{}
	var sweepResources []sweep.Sweepable

	pages := sqs.NewListQueuesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sweepResources, err
		}

		for _, v := range page.QueueUrls {
			r := resourceQueue()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
