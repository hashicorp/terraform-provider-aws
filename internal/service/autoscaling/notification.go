// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscaling_notification", name="Notification")
func resourceNotification() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNotificationCreate,
		ReadWithoutTimeout:   resourceNotificationRead,
		UpdateWithoutTimeout: resourceNotificationUpdate,
		DeleteWithoutTimeout: resourceNotificationDelete,

		Schema: map[string]*schema.Schema{
			"group_names": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"notifications": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTopicARN: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNotificationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	topic := d.Get(names.AttrTopicARN).(string)
	if err := addNotificationConfigToGroupsWithTopic(ctx, conn, flex.ExpandStringSet(d.Get("group_names").(*schema.Set)), flex.ExpandStringSet(d.Get("notifications").(*schema.Set)), topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Notification (%s): %s", topic, err)
	}

	// ARNs are unique, and these notifications are per ARN, so we re-use the ARN
	// here as the ID.
	d.SetId(topic)

	return append(diags, resourceNotificationRead(ctx, d, meta)...)
}

func resourceNotificationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	notifications, err := findNotificationsByTwoPartKey(ctx, conn, flex.ExpandStringValueSet(d.Get("group_names").(*schema.Set)), d.Id())

	if err == nil && len(notifications) == 0 {
		err = tfresource.NewEmptyResultError(nil)
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Notification %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Notification (%s): %s", d.Id(), err)
	}

	// Grab all applicable notification configurations for this Topic.
	// Each NotificationType will have a record, so 1 Group with 3 Types results
	// in 3 records, all with the same Group name.
	gRaw := make(map[string]struct{})
	nRaw := make(map[string]struct{})
	for _, n := range notifications {
		gRaw[aws.ToString(n.AutoScalingGroupName)] = struct{}{}
		nRaw[aws.ToString(n.NotificationType)] = struct{}{}
	}

	d.Set("group_names", tfmaps.Keys(gRaw))
	d.Set("notifications", tfmaps.Keys(nRaw))

	return diags
}

func resourceNotificationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	// Notifications API call is a PUT, so we don't need to diff the list, just
	// push whatever it is and AWS sorts it out
	nl := flex.ExpandStringSet(d.Get("notifications").(*schema.Set))

	o, n := d.GetChange("group_names")
	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	remove := flex.ExpandStringSet(o.(*schema.Set))
	add := flex.ExpandStringSet(n.(*schema.Set))

	topic := d.Get(names.AttrTopicARN).(string)

	if err := removeNotificationConfigToGroupsWithTopic(ctx, conn, remove, topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Notification (%s): %s", topic, err)
	}

	var update []*string
	if d.HasChange("notifications") {
		update = flex.ExpandStringSet(d.Get("group_names").(*schema.Set))
	} else {
		update = add
	}

	if err := addNotificationConfigToGroupsWithTopic(ctx, conn, update, nl, topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Notification (%s): %s", topic, err)
	}

	return append(diags, resourceNotificationRead(ctx, d, meta)...)
}

func resourceNotificationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	topic := d.Get(names.AttrTopicARN).(string)
	if err := removeNotificationConfigToGroupsWithTopic(ctx, conn, flex.ExpandStringSet(d.Get("group_names").(*schema.Set)), topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Notification (%s): %s", topic, err)
	}

	return diags
}

func addNotificationConfigToGroupsWithTopic(ctx context.Context, conn *autoscaling.Client, groups, notificationTypes []*string, topic string) error {
	for _, group := range groups {
		input := &autoscaling.PutNotificationConfigurationInput{
			AutoScalingGroupName: group,
			NotificationTypes:    aws.ToStringSlice(notificationTypes),
			TopicARN:             aws.String(topic),
		}

		_, err := conn.PutNotificationConfiguration(ctx, input)

		if err != nil {
			return err
		}
	}

	return nil
}

func removeNotificationConfigToGroupsWithTopic(ctx context.Context, conn *autoscaling.Client, groups []*string, topic string) error {
	for _, group := range groups {
		input := &autoscaling.DeleteNotificationConfigurationInput{
			AutoScalingGroupName: group,
			TopicARN:             aws.String(topic),
		}

		_, err := conn.DeleteNotificationConfiguration(ctx, input)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "doesn't exist") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func findNotificationsByTwoPartKey(ctx context.Context, conn *autoscaling.Client, groups []string, topic string) ([]awstypes.NotificationConfiguration, error) {
	input := &autoscaling.DescribeNotificationConfigurationsInput{
		AutoScalingGroupNames: groups,
	}

	return findNotifications(ctx, conn, input, func(v *awstypes.NotificationConfiguration) bool {
		return aws.ToString(v.TopicARN) == topic
	})
}

func findNotifications(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeNotificationConfigurationsInput, filter tfslices.Predicate[*awstypes.NotificationConfiguration]) ([]awstypes.NotificationConfiguration, error) {
	var output []awstypes.NotificationConfiguration

	pages := autoscaling.NewDescribeNotificationConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.NotificationConfigurations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
