// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_notifications_notification_configuration", sweepNotificationConfigurations)
	awsv2.Register("aws_notifications_event_rule", sweepEventRules, "aws_notifications_notification_configuration")
	awsv2.Register("aws_notifications_channel_association", sweepChannelAssociations, "aws_notifications_notification_configuration", "aws_notificationscontacts_email_contact")
}
