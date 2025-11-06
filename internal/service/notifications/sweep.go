// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_notifications_notification_configuration", sweepNotificationConfigurations)
	awsv2.Register("aws_notifications_event_rule", sweepEventRules, "aws_notifications_notification_configuration")
	awsv2.Register("aws_notifications_channel_association", sweepChannelAssociations, "aws_notifications_notification_configuration", "aws_notificationscontacts_email_contact")
}

func sweepEventRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NotificationsClient(ctx)
	var input notifications.ListNotificationConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := notifications.NewListNotificationConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.NotificationConfigurations {
			input := notifications.ListEventRulesInput{
				NotificationConfigurationArn: v.Arn,
			}

			pages := notifications.NewListEventRulesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.EventRules {
					sweepResources = append(sweepResources, framework.NewSweepResource(newEventRuleResource, client,
						framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
					)
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepNotificationConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NotificationsClient(ctx)
	var input notifications.ListNotificationConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := notifications.NewListNotificationConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.NotificationConfigurations {
			sweepResources = append(sweepResources, framework.NewSweepResource(newNotificationConfigurationResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}

func sweepChannelAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NotificationsClient(ctx)
	var input notifications.ListNotificationConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := notifications.NewListNotificationConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.NotificationConfigurations {
			input := notifications.ListChannelsInput{
				NotificationConfigurationArn: v.Arn,
			}

			pages := notifications.NewListChannelsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, v := range page.Channels {
					sweepResources = append(sweepResources, framework.NewSweepResource(newChannelAssociationResource, client,
						framework.NewAttribute(names.AttrARN, v)),
					)
				}
			}
		}
	}

	return sweepResources, nil
}
