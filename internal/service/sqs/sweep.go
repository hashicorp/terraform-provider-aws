//go:build sweep
// +build sweep

package sqs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_sqs_queue", &resource.Sweeper{
		Name: "aws_sqs_queue",
		F:    sweepQueues,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_cloudwatch_event_rule",
			"aws_elastic_beanstalk_environment",
			"aws_iot_topic_rule",
			"aws_lambda_function",
			"aws_s3_bucket",
			"aws_sns_topic",
		},
	})
}

func sweepQueues(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SQSConn()

	input := &sqs.ListQueuesInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListQueuesPagesWithContext(ctx, input, func(page *sqs.ListQueuesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queueUrl := range page.QueueUrls {
			r := ResourceQueue()
			d := r.Data(nil)
			d.SetId(aws.StringValue(queueUrl))
			err = sweep.DeleteResource(ctx, r, d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SQS Queue sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing SQS Queues: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
