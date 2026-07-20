// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_sns_platform_application", sweepPlatformApplications)
	awsv2.Register("aws_sns_topic", sweepTopics,
		"aws_autoscaling_group",
		"aws_backup_vault_notifications",
		"aws_budgets_budget",
		"aws_config_delivery_channel",
		"aws_dax_cluster",
		"aws_db_event_subscription",
		"aws_elasticache_cluster",
		"aws_elasticache_replication_group",
		"aws_glacier_vault",
		"aws_iot_topic_rule",
		"aws_redshift_event_subscription",
		"aws_s3_bucket",
		"aws_ses_configuration_set",
		"aws_ses_domain_identity",
		"aws_ses_email_identity",
		"aws_ses_receipt_rule_set",
		"aws_sns_platform_application",
	)
	awsv2.Register("aws_sns_topic_subscription", sweepTopicSubscriptions)
}

func sweepPlatformApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SNSClient(ctx)

	input := &sns.ListPlatformApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sns.NewListPlatformApplicationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.PlatformApplications {
			r := resourcePlatformApplication()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.PlatformApplicationArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepTopics(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SNSClient(ctx)

	input := &sns.ListTopicsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sns.NewListTopicsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Topics {
			r := resourceTopic()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TopicArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepTopicSubscriptions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SNSClient(ctx)

	input := &sns.ListSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sns.NewListSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Subscriptions {
			arn := aws.ToString(v.SubscriptionArn)

			if arn == "PendingConfirmation" {
				continue
			}

			r := resourceTopicSubscription()
			d := r.Data(nil)
			d.SetId(arn)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
