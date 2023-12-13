// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_autoscaling_notification")
func ResourceNotification() *schema.Resource {
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
			"topic_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNotificationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)
	gl := flex.ExpandStringSet(d.Get("group_names").(*schema.Set))
	nl := flex.ExpandStringSet(d.Get("notifications").(*schema.Set))

	topic := d.Get("topic_arn").(string)
	if err := addNotificationConfigToGroupsWithTopic(ctx, conn, gl, nl, topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Autoscaling Group Notification (%s): %s", topic, err)
	}

	// ARNs are unique, and these notifications are per ARN, so we re-use the ARN
	// here as the ID
	d.SetId(topic)
	return append(diags, resourceNotificationRead(ctx, d, meta)...)
}

func resourceNotificationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)
	gl := flex.ExpandStringSet(d.Get("group_names").(*schema.Set))

	opts := &autoscaling.DescribeNotificationConfigurationsInput{
		AutoScalingGroupNames: gl,
	}

	topic := d.Get("topic_arn").(string)
	// Grab all applicable notification configurations for this Topic.
	// Each NotificationType will have a record, so 1 Group with 3 Types results
	// in 3 records, all with the same Group name
	gRaw := make(map[string]bool)
	nRaw := make(map[string]bool)

	i := 0
	err := conn.DescribeNotificationConfigurationsPagesWithContext(ctx, opts, func(resp *autoscaling.DescribeNotificationConfigurationsOutput, lastPage bool) bool {
		if resp != nil {
			i++
			log.Printf("[DEBUG] Paging DescribeNotificationConfigurations for (%s), page: %d", d.Id(), i)
		} else {
			log.Printf("[DEBUG] Paging finished for DescribeNotificationConfigurations (%s)", d.Id())
			return false
		}

		for _, n := range resp.NotificationConfigurations {
			if n == nil {
				continue
			}

			if aws.StringValue(n.TopicARN) == topic {
				gRaw[aws.StringValue(n.AutoScalingGroupName)] = true
				nRaw[aws.StringValue(n.NotificationType)] = true
			}
		}
		return true // return false to stop paging
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Autoscaling Group Notification (%s): %s", topic, err)
	}

	// Grab the keys here as the list of Groups
	var gList []string
	for k := range gRaw {
		gList = append(gList, k)
	}

	// Grab the keys here as the list of Types
	var nList []string
	for k := range nRaw {
		nList = append(nList, k)
	}

	if err := d.Set("group_names", gList); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Autoscaling Group Notification (%s): %s", topic, err)
	}
	if err := d.Set("notifications", nList); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Autoscaling Group Notification (%s): %s", topic, err)
	}

	return diags
}

func resourceNotificationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)

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

	topic := d.Get("topic_arn").(string)

	if err := removeNotificationConfigToGroupsWithTopic(ctx, conn, remove, topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Autoscaling Group Notification (%s): %s", topic, err)
	}

	var update []*string
	if d.HasChange("notifications") {
		update = flex.ExpandStringSet(d.Get("group_names").(*schema.Set))
	} else {
		update = add
	}

	if err := addNotificationConfigToGroupsWithTopic(ctx, conn, update, nl, topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Autoscaling Group Notification (%s): %s", topic, err)
	}

	return append(diags, resourceNotificationRead(ctx, d, meta)...)
}

func addNotificationConfigToGroupsWithTopic(ctx context.Context, conn *autoscaling.AutoScaling, groups []*string, nl []*string, topic string) error {
	for _, group := range groups {
		opts := &autoscaling.PutNotificationConfigurationInput{
			AutoScalingGroupName: group,
			NotificationTypes:    nl,
			TopicARN:             aws.String(topic),
		}

		_, err := conn.PutNotificationConfigurationWithContext(ctx, opts)

		if err != nil {
			return fmt.Errorf("adding notifications for (%s): %w", aws.StringValue(group), err)
		}
	}

	return nil
}

func removeNotificationConfigToGroupsWithTopic(ctx context.Context, conn *autoscaling.AutoScaling, groups []*string, topic string) error {
	for _, group := range groups {
		opts := &autoscaling.DeleteNotificationConfigurationInput{
			AutoScalingGroupName: group,
			TopicARN:             aws.String(topic),
		}

		_, err := conn.DeleteNotificationConfigurationWithContext(ctx, opts)
		if err != nil {
			return fmt.Errorf("removing notifications for (%s): %w", aws.StringValue(group), err)
		}
	}
	return nil
}

func resourceNotificationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)

	gl := flex.ExpandStringSet(d.Get("group_names").(*schema.Set))

	topic := d.Get("topic_arn").(string)
	if err := removeNotificationConfigToGroupsWithTopic(ctx, conn, gl, topic); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Autoscaling Group Notification (%s): %s", topic, err)
	}
	return diags
}
